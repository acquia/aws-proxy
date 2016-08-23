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
	director := func(req *http.Request) {

		// Rewrite the request to desired server host.
		req.URL.Scheme = url.Scheme
		req.URL.Host = url.Host
		req.Host = url.Host

		// We need the body to be an io.ReadSeeker to calculate the hash. We
		// do this via the method below, however ioutil.ReadAll() empties
		// the body from the request so we have to re-add it.
		// TODO Is there a more efficient way to do this?
		buf, _ := ioutil.ReadAll(req.Body)
		body := bytes.NewReader(buf)
		req.Body = ioutil.NopCloser(body)

		// Turn keep-alive off. If we don't do this then proxied requests
		// will succeed via curl but fail from the browser.
		req.Header.Set("Connection", "close")

		// Remove all x-forwarded-* headers.
		// https://github.com/acquia/aws-proxy/issues/4
		for header, _ := range req.Header {
			if strings.HasPrefix(strings.ToLower(header), "x-forwarded-") {
				req.Header.Del(header)
			}
		}

		// Read the credentials and sign the request.
		// TODO Don't parse this on every request. There has to be a more
		// efficient way to do this unless the SDK is already smart.
		sess := session.New()
		signer := v4.NewSigner(sess.Config.Credentials)
		signer.Sign(req, body, service, region, time.Now())
	}

	return &httputil.ReverseProxy{
		Director: director,
	}
}
