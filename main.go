package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"

	"github.com/certifi/gocertifi"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

var (
	bind       = flag.String("b", ":8000", "bind address")
	access_key = flag.String("access_key", "", "S3 Access Key")
	secret_key = flag.String("secret_key", "", "S3 Secret Key")
	bucket     = flag.String("bucket", "", "S3 bucket")
	proxy      = flag.Bool("proxy", false, "transparent proxy")
)

func init() {
	flag.Parse()
}

func main() {
	var auth aws.Auth
	var err error

	if *access_key != "" && *secret_key != "" {
		auth, err = aws.GetAuth(*access_key, *secret_key)
	} else {
		auth, err = aws.EnvAuth()
	}
	if err != nil {
		log.Fatal(err)
	}

	if *bucket == "" {
		log.Fatal("Provide a bucket name")
	}

	client := s3.New(auth, aws.USEast)
	client.HTTPClient = func() *http.Client {
		rootCAs, _ := gocertifi.CACerts()
		return &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{RootCAs: rootCAs},
			},
		}
	}

	server := S3FileServer(&S3FileServerConfig{
		Bucket:           client.Bucket(*bucket),
		TransparentProxy: *proxy,
	})

	log.Println("listening on...", *bind)
	log.Fatal(http.ListenAndServe(*bind, server))
}
