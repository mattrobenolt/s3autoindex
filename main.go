package main

import (
	"flag"
	"log"
	"net/http"
)

type Config struct {
	bind       string
	access_key string
	secret_key string
	key_file   string
	bucket     string
	proxy      bool
}

var config = &Config{}

func init() {
	flag.StringVar(&config.bind, "b", ":8000", "bind address")
	flag.StringVar(&config.access_key, "access_key", "", "S3 Access Key")
	flag.StringVar(&config.secret_key, "secret_key", "", "S3 Secret Key")
	flag.StringVar(&config.bucket, "bucket", "", "S3/GS bucket")
	flag.StringVar(&config.key_file, "key_file", "", "GS key file")
	flag.BoolVar(&config.proxy, "proxy", false, "transparent proxy")
	flag.Parse()
}

func main() {
	if config.bucket == "" {
		log.Fatal("Provide a bucket name")
	}
	server := FileServer(config)
	log.Println("listening on...", config.bind)
	log.Fatal(http.ListenAndServe(config.bind, server))
}
