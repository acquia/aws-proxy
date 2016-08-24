package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/gorilla/handlers"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var conf *viper.Viper

func init() {
	conf = viper.New()

	conf.SetEnvPrefix("AWS_PROXY")
	conf.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	conf.AutomaticEnv()

	pflag.StringP("endpoint", "e", "", "The entry point for the web service, e.g. https://dynamodb.us-west-2.amazonaws.com")
	conf.BindPFlag("endpoint", pflag.Lookup("endpoint"))
	conf.SetDefault("endpoint", "")

	pflag.IntP("port", "p", 3000, "The port that the reverse proxy binds to")
	conf.BindPFlag("port", pflag.Lookup("port"))
	conf.SetDefault("port", 3000)

	pflag.Parse()
}

func main() {

	endpoint := conf.GetString("endpoint")
	if endpoint == "" {
		log.Fatal("missing required option --endpoint")
	}

	url, err := url.Parse(endpoint)
	if err != nil {
		log.Fatal(err)
	}

	region, service, err := ParseEndpointUrl(url)
	if err != nil {
		log.Fatal(err)
	}

	// Run the reverse proxy.
	port := strconv.Itoa(conf.GetInt("port"))
	handler := ReverseProxy(url, region, service)
	http.ListenAndServe(":"+port, handlers.CombinedLoggingHandler(os.Stdout, handler))
}

// ParseEndpointUrl parses the service and region from the endpoint.
// http://docs.aws.amazon.com/general/latest/gr/rande.html
func ParseEndpointUrl(url *url.URL) (service, region string, err error) {

	parts := strings.Split(url.Host, ".")
	size := len(parts)

	if size == 5 {
		service = parts[2]
		region = parts[1]
	} else if size == 4 {
		service = parts[1]
		region = parts[0]
	} else if size == 3 {
		service = parts[0]
		region = "us-east-1"
	} else {
		err = errors.New("url is not a valid aws entry point")
		return
	}

	return
}

// ReverseProxy modifies the request by signing it with the v4 signature.
func ReverseProxy(url *url.URL, service, region string) *httputil.ReverseProxy {
	targetQuery := url.RawQuery

	director := func(req *http.Request) {

		// Routes URLs to the scheme and host of the entry point.
		req.URL.Scheme = url.Scheme
		req.URL.Host = url.Host
		req.Host = url.Host

		// Routes URLs to the base path of the endpoint.
		//
		// Handle ES and Kibana quirks. Kibana is installed as a plugin and
		// is accessible via the /_plugin/kibana path, however it also
		// reaches back to the root path to make Elasticsearch queries.
		// Therefore we have to whitelise some of these paths for Kibnan to
		// function properly.
		// https://github.com/acquia/aws-proxy/issues/2
		if service == "es" && strings.HasPrefix(url.Path, "/_plugin/kibana") {
			switch {
			case req.URL.Path == "/_nodes":
			case strings.HasPrefix(req.URL.Path, "/.kibana-4"):
			case strings.HasSuffix(req.URL.Path, "/_mapping/field/*"):
			case strings.HasSuffix(req.URL.Path, "/_msearch"):
				// Don't rewrite the path.
			default:
				req.URL.Path = singleJoiningSlash(url.Path, req.URL.Path)
			}
		} else {
			req.URL.Path = singleJoiningSlash(url.Path, req.URL.Path)
		}

		// Stolen from the httputil.NewSingleHostReverseProxy function.
		// https://golang.org/src/net/http/httputil/reverseproxy.go
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}

		// We need the body to be an io.ReadSeeker to calculate the hash. We
		// do this via the method below, however ioutil.ReadAll() empties
		// the body from the request so we have to re-add it.
		// TODO Is there a more efficient way to do this?
		buf, _ := ioutil.ReadAll(req.Body)
		body := bytes.NewReader(buf)
		req.Body = ioutil.NopCloser(body)

		// If we don't set the Connection header's value to "closed", then
		// proxied requests sent via the browser will fail due to a
		// signature mismatch. The reason is that AWS always expects the
		// value of this header to be "closed" and generates the hash using
		// that value regardless of what is sent in the request.
		req.Header.Set("Connection", "close")

		// Remove all x-forwarded-* headers, as they cn cause a signature
		// mismatch. For example, AWS expects the x-forwarded-for header to
		// be the public IP of he EC2 instance, even if the request has
		// x-forwarded-for set for the IP that made the initial request.
		// https://github.com/acquia/aws-proxy/issues/4
		for header, _ := range req.Header {
			if strings.HasPrefix(strings.ToLower(header), "x-forwarded-") {
				req.Header.Del(header)
			}
		}

		// Read the credentials and sign the request.
		// TODO Don't parse this on every request. There has to be a more
		// efficient way to do this unless the SDK is already being smart.
		sess := session.New()
		signer := v4.NewSigner(sess.Config.Credentials)
		signer.Sign(req, body, service, region, time.Now())
	}

	return &httputil.ReverseProxy{
		Director: director,
	}
}

// Stolen from the httputil.singleJoiningSlash function.
// https://golang.org/src/net/http/httputil/reverseproxy.go
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
