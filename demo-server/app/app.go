package app

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	listingCmd "github.com/retro-framework/go-retro/commands/listing"
	"github.com/retro-framework/go-retro/framework/engine"
	"github.com/retro-framework/go-retro/framework/retro"
)

type contextKey string

var (
	ContextKeySessionID = contextKey("sessionID")
)

type renderCtx struct {
	r *http.Request

	messages map[string][]string
}

func NewServer(
	e engine.Engine,
	tplDir string,
	mountPoint string,
	projections map[string]Projection,
) server {
	var s = &server{e: e}
	s.parseTemplates(e, tplDir)
	return *s
}

type Projection interface {
	Get(string) interface{}
}

type server struct {
	e           engine.Engine
	template    *template.Template
	mountpoint  string
	projections map[string]Projection
}

func (s server) NewProfileServer() profile {
	return profile{s}
}

func (s server) NewListingServer() listing {
	return listing{s}
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

type listing struct {
	server
}

func (l listing) NewHandler(w http.ResponseWriter, req *http.Request) {

	// var retroCmd = retro.CommandDesc{
	// 	Name: "create_listing",
	// }

	// type createListingParams struct {
	// 	Name string `json:"name"`
	// 	Desc string `json:"desc"`
	// }

	if req.Method == "GET" {
		err := l.template.ExecuteTemplate(w, "listing/new", renderCtx{req})
		if err != nil {
			fmt.Fprintf(w, err.Error())
		}
	} else if req.Method == "POST" {

		err := req.ParseMultipartForm(1 << 20)
		if err != nil {
			fmt.Fprintf(w, err.Error())
		}

		// var cmd = listing.CreateListing{}
		var cmd = struct {
			Name string          `json:"name"`
			Args listingCmd.Args `json:"args"`
		}{
			Name: "create_listing",
			Args: listingCmd.Args{
				Name: req.Form.Get("name"),
				Desc: req.Form.Get("desc"),
			},
		}

		// Parse the startPrice into an Int16
		if price, err := strconv.ParseInt(req.FormValue("startPrice"), 10, 16); err != nil {
			fmt.Fprintf(w, err.Error())
		} else {
			cmd.Args.StartPrice = uint16(price)
		}

		// Parse the files attached, if any
		for _, img := range req.MultipartForm.File["images"] {
			f, err := img.Open()
			defer f.Close()
			if err != nil {
				fmt.Fprintf(w, err.Error())
				continue
			}
			b, err := ioutil.ReadAll(f)
			if err != nil {
				fmt.Fprintf(w, err.Error())
				continue
			}
			cmd.Args.Images = append(cmd.Args.Images, b)
		}

		cmdB, err := json.Marshal(cmd)
		if err != nil {
			fmt.Fprintf(w, err.Error())
			return
		}

		var sessionIDStr = req.Context().Value(ContextKeySessionID)
		newListingName, err := l.e.Apply(req.Context(), ioutil.Discard, retro.SessionID(sessionIDStr.(string)), cmdB)
		if err != nil {
			fmt.Fprintf(w, err.Error())
			return
		}

		SetFlash(w, "message", []byte(fmt.Sprintf("Listing %q created successfully!", newListingName))
		http.Redirect(w, req, newListingName, http.StatusFound)
	}
}

var (
	hasIdentity = func(rtx renderCtx) bool {
		return false
	}
	sessionID = func(rtx renderCtx) string {
		var sessionID = rtx.r.Context().Value(ContextKeySessionID)
		return sessionID.(string)
	}
	urlFor = func(rtx renderCtx, rest string) string {
		fmt.Println("got", rest, "making", filepath.Join("/demo-app", rest))
		return filepath.Join("/demo-app/", rest)
	}
)

/*	Below this line is just boilerplate */
func (s *server) parseTemplates(e engine.Engine, dir string) {
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
	s.template = templ
}

// taken from https://www.alexedwards.net/blog/simple-flash-messages-in-golang
func SetFlash(w http.ResponseWriter, name string, value []byte) {
	c := &http.Cookie{Name: name, Value: encode(value)}
	http.SetCookie(w, c)
}

func GetFlash(w http.ResponseWriter, r *http.Request, name string) ([]byte, error) {
	c, err := r.Cookie(name)
	if err != nil {
		switch err {
		case http.ErrNoCookie:
			return nil, nil
		default:
			return nil, err
		}
	}
	value, err := decode(c.Value)
	if err != nil {
		return nil, err
	}
	dc := &http.Cookie{Name: name, MaxAge: -1, Expires: time.Unix(1, 0)}
	http.SetCookie(w, dc)
	return value, nil
}

// -------------------------

func encode(src []byte) string {
	return base64.URLEncoding.EncodeToString(src)
}

func decode(src string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(src)
}
