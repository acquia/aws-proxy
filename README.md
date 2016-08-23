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

This project uses the [GB build tool](https://getgb.io/). Assuming that GB
is installed, run the following command in the project's root directory to
build the `aws-proxy` binary:

```shell
gb build
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

