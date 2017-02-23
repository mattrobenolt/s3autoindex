package main

import (
	"html/template"
	"net/http"
	"time"

	"github.com/mattrobenolt/size"
	"github.com/mitchellh/goamz/s3"
)

var indexTemplate = template.Must(template.New("index").Parse(`<!doctype html>
<html>
<head>
<title>Index of {{ .Bucket }}{{ .Path }}</title>
<style type="text/css">table{ font-family: monospace; } td{ padding-right: 25px; }</style>
</head>
<body>
<h1>Index of {{ .Bucket }}{{ .Path }}</h1>
<hr>
<table>
{{if not .Root}}<tr><td><a href="../">../</a></td><td></td><td></td></tr>{{end}}
{{range .Folders}}<tr><td><a href="{{ .Name }}">{{ .Name }}</a></td><td></td><td>-</td></tr>{{end}}
{{range .Keys}}<tr><td><a href="{{ .Name }}">{{ .Name }}</a></td><td>{{ .LastModified.Format "_2-Jan-2006 15:04" }}</td><td>{{ .Size }}</td></tr>{{end}}
</table>
<hr>
</body>
</html>
`))

type Folder struct {
	Name string
}

type Key struct {
	Name         string
	LastModified time.Time
	Size         size.Capacity
}

type Result struct {
	Root    bool
	Path    string
	Folders []Folder
	Keys    []Key
	Bucket  string
}

type s3FileServer struct {
	bucket *s3.Bucket
}

func S3FileServer(bucket *s3.Bucket) http.Handler {
	return &s3FileServer{bucket}
}

func (f *s3FileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	prefix := path[1:]

	resp, err := f.bucket.List(prefix, "/", "", 0)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// No keys or folders at this path, so 404
	if len(resp.Contents) == 0 && len(resp.CommonPrefixes) == 0 {
		http.NotFound(w, r)
		return
	}

	// 1 key, no paths, and key matches what we're looking for,
	// so this must be a file we've requested to download.
	// Redirect to a signed URL
	if len(resp.Contents) == 1 && len(resp.CommonPrefixes) == 0 && resp.Contents[0].Key == prefix {
		url := f.bucket.SignedURL(resp.Contents[0].Key, time.Now().Add(5*time.Minute))
		http.Redirect(w, r, url, 302)
		return
	}

	// No keys, but 1 subdirectory match with a trailing slash.
	// Append trailing slash and redirect
	if len(resp.Contents) == 0 && len(resp.CommonPrefixes) == 1 && resp.CommonPrefixes[0] == prefix+"/" {
		http.Redirect(w, r, path+"/", 302)
		return
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

	result := &Result{
		Root:    path == "/",
		Path:    path,
		Folders: folders,
		Keys:    keys,
		Bucket:  f.bucket.Name,
	}

	indexTemplate.Execute(w, result)
}
