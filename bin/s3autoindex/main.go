package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"

	"github.com/mattrobenolt/s3autoindex"
)

var (
	bind      = flag.String("b", ":8000", "bind address")
	accesskey = flag.String("accesskey", "", "S3 Access Key")
	secretkey = flag.String("secretkey", "", "S3 Secret Key")
	bucket    = flag.String("bucket", "", "S3 bucket")
)

func init() {
	flag.Parse()
}

func main() {
	auth, err := aws.GetAuth(*accesskey, *secretkey)
	if err != nil {
		log.Fatal(err)
	}

	if *bucket == "" {
		log.Fatal("Provide a bucket name")
	}

	server := s3autoindex.S3FileServer(
		s3.New(auth, aws.USEast).Bucket(*bucket),
	)

	log.Println("listening on...", *bind)
	log.Fatal(http.ListenAndServe(*bind, server))
}
