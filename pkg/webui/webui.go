package webui

import (
	"html/template"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

type WebUI struct {
	static  http.Handler
	dynamic *template.Template
	context Context
}

type (
	Context map[string]interface{}
	Values map[string]interface{}
)

func New(dir string, context Context) (*WebUI, error) {
	files := make([]string, 0, len(context))
	for file := range context {
		files = append(files, filepath.Join(dir, file))
	}

	dynamic, err := template.ParseFiles(files...)
	if err != nil {
		return nil, err
	}

	webUI := &WebUI{
		static:  http.FileServer(http.Dir(dir)),
		dynamic: dynamic,
		context: context,
	}

	return webUI, nil
}

func canonicalizePath(path string) string {
	if len(path) == 0 || strings.HasSuffix(path, "/") {
		path += "index.html"
	}

	relative, err := filepath.Rel("/", path)
	if err != nil {
		return path
	}

	return relative
}

func (ui *WebUI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := canonicalizePath(r.URL.Path)
	if data, ok := ui.context[name]; ok {
		ui.renderDynamic(w, name, data)
	} else {
		ui.static.ServeHTTP(w, r)
	}
}

func (ui *WebUI) renderDynamic(w http.ResponseWriter, file string, data interface{}) {
	contentType := mime.TypeByExtension(filepath.Ext(file))
	if len(contentType) != 0 {
		w.Header().Add("Content-Type", contentType)
	}

	err := ui.dynamic.ExecuteTemplate(w, file, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}