package main

import (
	"html/template"
	"net/http"
	"time"

	"github.com/mattrobenolt/size"
)

type Backend interface {
	ServeHTTP(http.ResponseWriter, *http.Request) *Result
}

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

type fileServer struct {
	backend Backend
}

func FileServer(config *Config) http.Handler {
	if len(config.bucket) > 5 && config.bucket[:5] == "gs://" {
		return &fileServer{GSBackend(config)}
	}
	return &fileServer{S3Backend(config)}
}

func (f *fileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	result := f.backend.ServeHTTP(w, r)
	if result != nil {
		indexTemplate.Execute(w, result)
	}
}
