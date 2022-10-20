package main

import (
	"html/template"
	"log"
	"net/http"
)

// goVCSHandler serves HTML files with a <meta name=go-import> header,
// based on a module repository resolver.
type goVCSHandler struct {
	r          *Resolver
	hostHeader string
}

func (h *goVCSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rec, err := h.r.Resolve(r.Context(), h.host(r)+r.URL.Path)
	if err != nil {
		code := http.StatusInternalServerError
		if err == ErrNotFound {
			code = http.StatusNotFound
		}
		w.Header().Set("Content-Type", "text/plain")
		http.Error(w, err.Error(), code)
		log.Printf("%d [%s] %s %s", code, r.RemoteAddr, r.Method, r.URL.Path)
		return
	}

	type tmplData struct {
		Record  *Record
		RepoURL template.URL
	}

	w.Header().Set("Content-Type", "text/html")
	if err := htmlTmpl.Execute(w, tmplData{rec, template.URL(rec.RepoURL)}); err != nil {
		w.Header().Set("Content-Type", "text/plain")
		log.Printf("Template rendering failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	log.Printf("%d [%s] %s %s", http.StatusOK, r.RemoteAddr, r.Method, r.URL.Path)
}

var htmlTmpl = template.Must(template.New("").Parse(`<!DOCTYPE html>
<html>
  <meta name="go-import" content="{{.Record.Root}} {{.Record.VCS}} {{.Record.RepoURL}}">
  <title>Go Repository Resolver</title>
  See <a href="{{.RepoURL}}">{{.Record.Root}}</a>.
</html>
`))

// host returns the requested host.
func (h *goVCSHandler) host(r *http.Request) string {
	if h.hostHeader == "host" {
		return r.Host
	}

	return r.Header.Get(h.hostHeader)
}
