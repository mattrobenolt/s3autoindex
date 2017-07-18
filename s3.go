package main

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/mattrobenolt/size"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

type s3Backend struct {
	config *s3FileServerConfig
}

type s3FileServerConfig struct {
	Bucket           *s3.Bucket
	TransparentProxy bool
}

func S3Backend(config *Config) Backend {
	var auth aws.Auth
	var err error
	if config.access_key != "" && config.secret_key != "" {
		auth, err = aws.GetAuth(config.access_key, config.secret_key)
	} else {
		auth, err = aws.EnvAuth()
	}
	if err != nil {
		log.Fatal(err)
	}
	client := s3.New(auth, aws.USEast)
	return &s3Backend{&s3FileServerConfig{
		Bucket:           client.Bucket(config.bucket),
		TransparentProxy: config.proxy,
	}}
}

func (f *s3Backend) ServeHTTP(w http.ResponseWriter, r *http.Request) *Result {
	path := r.URL.Path
	prefix := path[1:]

	resp, err := f.config.Bucket.List(prefix, "/", "", 0)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return nil
	}

	// No keys or folders at this path, so 404
	if len(resp.Contents) == 0 && len(resp.CommonPrefixes) == 0 {
		http.NotFound(w, r)
		return nil
	}

	// 1 key, no paths, and key matches what we're looking for,
	// so this must be a file we've requested to download.
	if len(resp.Contents) == 1 && len(resp.CommonPrefixes) == 0 && resp.Contents[0].Key == prefix {
		key := resp.Contents[0]
		if f.config.TransparentProxy {
			httpResponse, err := f.config.Bucket.GetResponse(key.Key)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return nil
			}
			for k, v := range httpResponse.Header {
				for _, v2 := range v {
					w.Header().Add(k, v2)
				}
			}
			io.Copy(w, httpResponse.Body)
		} else {
			// Redirect to a signed URL
			url := f.config.Bucket.SignedURL(key.Key, time.Now().Add(5*time.Minute))
			http.Redirect(w, r, url, 302)
		}
		return nil
	}

	// No keys, but 1 subdirectory match with a trailing slash.
	// Append trailing slash and redirect
	if len(resp.Contents) == 0 && len(resp.CommonPrefixes) == 1 && resp.CommonPrefixes[0] == prefix+"/" {
		http.Redirect(w, r, path+"/", 302)
		return nil
	}

	folders := make([]Folder, 0)
	for _, folder := range resp.CommonPrefixes {
		folders = append(folders, Folder{folder[len(prefix):]})
	}

	keys := make([]Key, 0)
	for _, key := range resp.Contents {
		if key.Key == "" || key.Key == prefix {
			continue
		}

		lastModified, _ := time.Parse(time.RFC3339, key.LastModified)
		keys = append(keys, Key{
			Name:         key.Key[len(prefix):],
			LastModified: lastModified,
			Size:         size.Capacity(key.Size) * size.Byte,
		})
	}

	return &Result{
		Root:    path == "/",
		Path:    path,
		Folders: folders,
		Keys:    keys,
		Bucket:  f.config.Bucket.Name,
	}
}
