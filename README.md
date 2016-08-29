# AWS Proxy

This app reverse proxies entry points for Amazon web services. Proxied
requests are signed using the [v4 signature](http://docs.aws.amazon.com/general/latest/gr/signature-version-4.html)
which allows direct access to the endpoint with tools such as `curl`
without having to sign the requests.

The primary use case for this app is proxying [Amazon Elasticsearch Service](https://aws.amazon.com/elasticsearch-service/)
domains so that developers can more easily use existing tools and libraries
that integrate with Elasticsearch, although other AWS services can be
proxied as well.

This project is inspired by the https://github.com/cllunsford/aws-signing-proxy
library and borrows some core techniques.

## Installation

Run the following command in the project's root to build the aws-proxy binary:

```shell
GOPATH=$PWD go build -o ./bin/aws-proxy aws-proxy
```

## Usage

This app reads configuration from environment variables, the AWS credentials
file, the CLI configuration file, and instance profile credentials. See
http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html#cli-metadata
for more details.

Run the proxy, replacing `my-domain` and `us-west-2` according to your environment.

```shell
./bin/aws-proxy --port 9200 --endpoint=https://my-domain.us-west-2.es.amazonaws.com
```

Consume the service with tools like `curl`:

```shell
curl http://localhost:9200
```

#### Proxying Kibana from the document root

If you just want to proxy Kibana and serve it from the document root, add
the Kibana plugin's path to the endpoint:

```shell
./bin/aws-proxy --port 5601 --endpoint=https://my-domain.us-west-2.es.amazonaws.com/_plugin/kibana
```

Be aware that there is some magic behind the scenes to make this possible.
Participate in https://github.com/acquia/aws-proxy/issues/6, pull back the
curtain, and make things less magical.

##### Securing Kibana

You probably don't want to expose Kibana to the world, so check out
[Bitly's Oauth2 Proxy](https://github.com/bitly/oauth2_proxy) and set the
AWS Proxy as its upstream endpoint.

If you do put AWS Proxy behind another reverse proxy, make sure to pass the
`--behind-reverse-proxy` option so that the IP of the host that made the
original request is logged.

### Running With Upstart

Use [Upstart](http://upstart.ubuntu.com/) to start aws-proxy during boot
and supervise it while the system is running. Add a file to `/etc/init` with
the following contents, replacing `/path/to` and `my-domain` according to
your environment.

```
description "AWS Proxy"
start on runlevel [2345]

respawn
respawn limit 10 5

exec /path/to/aws-proxy --port 9200 --endpoint=https://my-domain.us-west-2.es.amazonaws.com
```

## Alternate projects

We aren't in the business of pushing tools, so you should also look at the
projects below so that you can make the best decision for your use case.

* https://github.com/coreos/aws-auth-proxy
* https://github.com/cllunsford/aws-signing-proxy
* https://github.com/anomalizer/ngx_aws_auth

