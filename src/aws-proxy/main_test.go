package main

import (
	"net/url"
	"testing"
)

// TestParseEndpointWithNoRegion tests parsing the service and region from
// an Amazon web service's entry point with no region-specific endpoint.
func TestParseEndpointWithNoRegion(t *testing.T) {
	url, err := url.Parse("https://iam.amazonws.com")
	if err != nil {
		panic(err)
	}

	service, region, err := ParseEndpointUrl(url)
	if service != "iam" {
		t.Errorf("expecting service to be 'iam': got %s", service)
	}
	if region != "us-east-1" {
		t.Errorf("expecting region to be 'us-east-1': got %s", region)
	}
	if err != nil {
		t.Errorf("error parsing endpoint URL: got %s", err)
	}
}

// TestParseEndpointWithRegion tests parsing the service and region from an
// Amazon web service's entry point with a region-specific endpoint.
func TestParseEndpointWithRegion(t *testing.T) {
	url, err := url.Parse("https://dynamodb.us-west-2.amazonaws.com")
	if err != nil {
		panic(err)
	}

	service, region, err := ParseEndpointUrl(url)
	if service != "dynamodb" {
		t.Errorf("expecting service to be 'dynamodb': got %s", service)
	}
	if region != "us-west-2" {
		t.Errorf("expecting region to be 'us-west-2': got %s", region)
	}
	if err != nil {
		t.Errorf("error parsing endpoint URL: got %s", err)
	}
}

// TestParseEndpointWithPrefix tests parsing the service and region from an
// Amazon web service's entry point with a region-specific endpoint and
// prefix, e.g. Amazon's IoT and ES services.
func TestParseEndpointWithPrefix(t *testing.T) {
	url, err := url.Parse("https://my-domain.us-west-2.es.amazonaws.com")
	if err != nil {
		panic(err)
	}

	service, region, err := ParseEndpointUrl(url)
	if service != "es" {
		t.Errorf("expecting service to be 'es': got %s", service)
	}
	if region != "us-west-2" {
		t.Errorf("expecting region to be 'us-west-2': got %s", region)
	}
	if err != nil {
		t.Errorf("error parsing endpoint URL: got %s", err)
	}
}
