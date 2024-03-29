package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/storage"
	"github.com/mattrobenolt/size"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type gsBackend struct {
	config *gsFileServerConfig
}

type gsFileServerConfig struct {
	Context          context.Context
	BucketName       string
	Bucket           *storage.BucketHandle
	TransparentProxy bool
}

func GSBackend(config *Config) Backend {
	ctx := context.Background()
	key_file := config.key_file
	if key_file == "" {
		key_file = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	}

	var client *storage.Client
	var err error
	if key_file == "" {
		log.Print("GOOGLE_APPLICATION_CREDENTIALS not found in environment. Using default credentials.")
		client, err = storage.NewClient(ctx)
	} else {
		opt := option.WithServiceAccountFile(key_file)
		client, err = storage.NewClient(ctx, opt)
	}

	if err != nil {
		log.Fatal(err)
	}
	return &gsBackend{&gsFileServerConfig{
		Context:          ctx,
		BucketName:       config.bucket[5:],
		Bucket:           client.Bucket(config.bucket[5:]),
		TransparentProxy: config.proxy,
	}}
}

func (f *gsBackend) ServeHTTP(w http.ResponseWriter, r *http.Request) *Result {
	path := r.URL.Path
	prefix := path[1:]

	objects := f.config.Bucket.Objects(f.config.Context, &storage.Query{
		Delimiter: "/",
		Prefix:    prefix,
	})
	folders := make([]Folder, 0)
	keys := make([]Key, 0)
	for {
		attrs, err := objects.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			http.Error(w, err.Error(), 500)
			return nil
		}
		if attrs.Name == "" && attrs.Prefix != "" {
			folders = append(folders, Folder{attrs.Prefix[len(prefix):]})
		} else if attrs.Name != "" && attrs.Prefix == "" {
			name := attrs.Name[len(prefix):]
			keys = append(keys, Key{
				Name:         name,
				LastModified: attrs.Updated,
				Size:         size.Capacity(attrs.Size) * size.Byte,
			})
		}
	}

	// No keys or folders at this path, so 404
	if len(folders) == 0 && len(keys) == 0 {
		http.NotFound(w, r)
		return nil
	}

	// There exists a key that matches exactly: return as file
	for _, key := range keys {
		if key.Name != "" {
			continue;
		}

		key.Name = prefix;
		if f.config.TransparentProxy {
			reader, err := f.config.Bucket.Object(key.Name).NewReader(f.config.Context)
			if err != nil {
				http.Error(w, err.Error(), 500)
			} else {
				if reader.CacheControl() != "" {
					w.Header().Add("Cache-Control", reader.CacheControl())
				}
				if reader.ContentEncoding() != "" {
					w.Header().Add("Content-Encoding", reader.ContentEncoding())
				}
				if reader.ContentType() != "" {
					w.Header().Add("Content-Type", reader.ContentType())
				}
				io.Copy(w, reader)
			}
		} else {
			http.Error(w, "Signed urls not implemented for Google", 500)
		}
		return nil
	}

	// No keys, but 1 subdirectory match with a trailing slash.
	// Append trailing slash and redirect
	if len(keys) == 0 && len(folders) == 1 && folders[0].Name == "/" {
		http.Redirect(w, r, path+"/", 302)
		return nil
	}

	// No trailing slash => must be a file, but there was no exact match, so 404
	if path[len(path)-1:] != "/" {
		http.NotFound(w, r)
		return nil
	}

	return &Result{
		Root:    path == "/",
		Path:    path,
		Folders: folders,
		Keys:    keys,
		Bucket:  f.config.BucketName,
	}
}
