package app

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/retro-framework/go-retro/framework/engine"
)

type contextKey string

type renderCtx struct {
	r *http.Request
}

var (
	ContextKeySessionID = contextKey("sessionID")
)

func NewServer(
	e engine.Engine,
	tplDir string,
	projections map[string]Projection,
) server {
	return server{e: e, template: parseTemplates(e, tplDir)}
}

type Projection interface {
	Get(string) interface{}
}

type server struct {
	e           engine.Engine
	template    *template.Template
	projections map[string]Projection
}

func (s server) NewProfileServer() profile {
	return profile{s}
}

func (s server) IndexHandler(w http.ResponseWriter, req *http.Request) {
	err := s.template.ExecuteTemplate(w, "index", renderCtx{req})
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
}

type profile struct {
	server
}

func (p profile) ShowHandler(w http.ResponseWriter, req *http.Request) {
	// err := p.template.ExecuteTemplate(w, "profile", renderCtx{req})
	// if err != nil {
	// 	fmt.Fprintf(w, err.Error())
	// }
}

func (p profile) CreateHandler(w http.ResponseWriter, req *http.Request) {

}

func (p profile) UpdateHandler(w http.ResponseWriter, req *http.Request) {

}

var (
	hasIdentity = func(rtx renderCtx) bool {
		return false
	}
	sessionID = func(rtx renderCtx) string {
		var sessionID = rtx.r.Context().Value(ContextKeySessionID)
		return sessionID.(string)
	}
	urlFor = func(rtx renderCtx, _ string) string {
		fmt.Println(rtx.r.URL.String())
		return rtx.r.URL.String()
	}
)

/*	Below this line is just boilerplate */
func parseTemplates(e engine.Engine, dir string) *template.Template {
	templ := template.New("")
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".tpl.html") {
			_, err := templ.ParseFiles(path)
			templ.Funcs(template.FuncMap{
				"hasIdentity": hasIdentity,
				"sessionID":   sessionID,
				"urlFor":      urlFor,
			})
			if err != nil {
				log.Println(err)
			}
		}
		return err
	})
	if err != nil {
		panic(err)
	}
	return templ
}
