package main

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/acquia/aws-proxy/proxy"
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

	pflag.BoolP("behind-reverse-proxy", "b", false, "Set this flag if the proxy is being run behind another")
	conf.BindPFlag("behind-reverse-proxy", pflag.Lookup("behind-reverse-proxy"))
	conf.SetDefault("behind-reverse-proxy", false)

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

	region, service, err := proxy.ParseEndpointUrl(url)
	if err != nil {
		log.Fatal(err)
	}

	// Build the reverse proxy handler chain.
	var handler http.Handler
	proxy_handler := proxy.ReverseProxy(url, region, service)
	if conf.GetBool("behind-reverse-proxy") {
		handler = handlers.CombinedLoggingHandler(os.Stdout, handlers.ProxyHeaders(proxy_handler))
	} else {
		handler = handlers.CombinedLoggingHandler(os.Stdout, proxy_handler)
	}

	// Run the reverse proxy.
	port := strconv.Itoa(conf.GetInt("port"))
	http.ListenAndServe(":"+port, handler)
}
