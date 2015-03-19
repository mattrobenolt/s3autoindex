package s3autoindex

import (
	"html/template"
	"io"
	"net/http"

	"github.com/mitchellh/goamz/s3"
)

var indexTemplate = template.Must(template.New("index").Parse(`<!doctype html>
<html>
<head><title>Index of {{ .Path }}</title></head>
<body>
<h1>Index of {{ .Path }}</h1>
<hr>
<table style="font-family:monospace;">
{{if not .Root}}<tr><td><a href="../">../</a></td><td></td><td></td></tr>{{end}}
{{range .Folders}}<tr><td><a href="{{ .Name }}">{{ .Name }}</a></td><td></td><td>-</td></tr>{{end}}
{{range .Keys}}<tr><td><a href="{{ .Key }}">{{ .Key }}</a></td><td>{{ .LastModified }}</td><td>{{ .Size }}</td></tr>{{end}}
</table>
<hr>
</body>
</html>
`))

type Folder struct {
	Name string
}

type Result struct {
	Root    bool
	Path    string
	Folders []Folder
	Keys    []s3.Key
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

	if len(resp.Contents) == 0 && len(resp.CommonPrefixes) == 0 {
		http.NotFound(w, r)
		return
	}

	if len(resp.Contents) == 1 && len(resp.CommonPrefixes) == 0 {
		f.DeliverKey(w, resp.Contents[0])
		return
	}

	if len(resp.Contents) == 0 && len(resp.CommonPrefixes) == 1 && resp.CommonPrefixes[0] == prefix+"/" {
		http.Redirect(w, r, path+"/", 302)
		return
	}

	folders := make([]Folder, 0)
	for _, folder := range resp.CommonPrefixes {
		folders = append(folders, Folder{folder[len(prefix):]})
	}

	keys := make([]s3.Key, 0)
	for _, key := range resp.Contents {
		if key.Key == "" || key.Key == prefix {
			continue
		}

		key.Key = key.Key[len(prefix):]

		keys = append(keys, key)
	}

	result := &Result{
		Root:    path == "/",
		Path:    path,
		Folders: folders,
		Keys:    keys,
	}

	indexTemplate.Execute(w, result)
}

func (f *s3FileServer) DeliverKey(w http.ResponseWriter, key s3.Key) {
	reader, err := f.bucket.GetReader(key.Key)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer reader.Close()

	io.Copy(w, reader)
}
