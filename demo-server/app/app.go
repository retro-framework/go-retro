package app

import (
	"bytes"
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

	"github.com/retro-framework/go-retro/framework/engine"
	"github.com/retro-framework/go-retro/framework/retro"
	"github.com/retro-framework/go-retro/projections"

	identityCmd "github.com/retro-framework/go-retro/commands/identity"
	listingCmd "github.com/retro-framework/go-retro/commands/listing"
)

type contextKey string

var (
	ContextKeySessionID = contextKey("sessionID")
)

type displayListing struct {
	Name string `json:"name"`
	Desc string `json:"desc"`

	Hash string `json:"hash"`
}

type renderCtx struct {
	r *http.Request

	// not used?
	Flash string

	// wow this is bad!
	Listings []projections.Listing
	Profile  projections.Profile

	// General thing
	Sessions projections.Sessions
}

func NewServer(
	e engine.Engine,
	session projections.Sessions,
	tplDir string,
	mountPoint string,
) server {
	var s = &server{e: e, session: session}
	s.parseTemplates(e, tplDir)
	return *s
}

type Projection interface {
	Get(string) interface{}
}

type server struct {
	e          engine.Engine
	template   *template.Template
	mountpoint string
	session    projections.Sessions
}

func (s server) NewProfileServer(projection projections.Profiles) profile {
	return profile{s, projection}
}

func (s server) NewListingServer(projection projections.Listings) listing {
	return listing{s, projection}
}

func (s server) IndexHandler(w http.ResponseWriter, req *http.Request) {
	err := s.template.ExecuteTemplate(w, "index", renderCtx{
		r:        req,
		Sessions: s.session,
		Flash:    GetFlash(w, req, "message"),
	})
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
}

type profile struct {
	server
	projection projections.Profiles
}

func (p profile) ShowHandler(w http.ResponseWriter, req *http.Request) {
	err := p.template.ExecuteTemplate(w, "profile", renderCtx{
		r:        req,
		Sessions: p.session,
		Flash:    GetFlash(w, req, "message"),
	})
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
}

func (p profile) NewHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		err := p.template.ExecuteTemplate(w, "profile/new", renderCtx{
			r:        req,
			Sessions: p.session,
			Flash:    GetFlash(w, req, "message"),
		})
		if err != nil {
			fmt.Fprintf(w, err.Error())
		}
	} else if req.Method == "POST" {

		err := req.ParseMultipartForm(1 << 20)
		if err != nil {
			fmt.Fprintf(w, err.Error())
		}

		var cmd = struct {
			Name string                 `json:"name"`
			Args identityCmd.CreateArgs `json:"args"`
		}{
			Name: "create_identity",
			Args: identityCmd.CreateArgs{Name: req.Form.Get("name")},
		}

		// Parse the publishNow into a bool
		if s, err := strconv.ParseBool(req.Form.Get("visibilityPublic")); err != nil {
			fmt.Fprintf(w, err.Error())
		} else {
			cmd.Args.PubliclyVisible = s
		}

		// Parse the files attached, if any
		avatar, _, err := req.FormFile("avatar")
		if err != nil {
			fmt.Fprintf(w, err.Error())
			return
		}
		defer avatar.Close()
		avatarBuf, err := ioutil.ReadAll(avatar)
		if err != nil {
			fmt.Fprintf(w, err.Error())
			return
		}
		cmd.Args.Avatar = avatarBuf

		cmdB, err := json.Marshal(cmd)
		if err != nil {
			fmt.Fprintf(w, err.Error())
			return
		}

		var (
			b            bytes.Buffer
			sessionIDStr = req.Context().Value(ContextKeySessionID)
		)

		_, err = p.e.Apply(req.Context(), &b, retro.SessionID(sessionIDStr.(string)), cmdB)
		if err != nil {
			fmt.Fprintf(w, err.Error())
			return
		}

		SetFlash(w, "message", []byte(fmt.Sprintf("Profile %q created successfully!", b.String())))
		http.Redirect(w, req, "/demo-app/", http.StatusFound)
	}
}

func (p profile) UpdateHandler(w http.ResponseWriter, req *http.Request) {

}

type listing struct {
	server
	projection projections.Listings
}

func (l listing) ListHandler(w http.ResponseWriter, req *http.Request) {

	var listings, err = l.projection.PublishedListings(req.Context())
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	var rtx = renderCtx{
		r:        req,
		Listings: listings,
		Sessions: l.session,
		Flash:    GetFlash(w, req, "message"),
	}

	err = l.template.ExecuteTemplate(w, "listing/list", rtx)
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

}

func (l listing) NewHandler(w http.ResponseWriter, req *http.Request) {

	if req.Method == "GET" {
		err := l.template.ExecuteTemplate(w, "listing/new", renderCtx{
			r:        req,
			Sessions: l.session,
			Flash:    GetFlash(w, req, "message"),
		})
		if err != nil {
			fmt.Fprintf(w, err.Error())
		}
	} else if req.Method == "POST" {

		err := req.ParseMultipartForm(1 << 20)
		if err != nil {
			fmt.Fprintf(w, err.Error())
		}

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

		// Parse the publishNow into a bool
		if s, err := strconv.ParseBool(req.Form.Get("publishNow")); err != nil {
			fmt.Fprintf(w, err.Error())
		} else {
			cmd.Args.PublishNow = s
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

		var (
			b            bytes.Buffer
			sessionIDStr = req.Context().Value(ContextKeySessionID)
		)
		_, err = l.e.Apply(req.Context(), &b, retro.SessionID(sessionIDStr.(string)), cmdB)
		if err != nil {
			fmt.Fprintf(w, err.Error())
			return
		}

		SetFlash(w, "message", []byte(fmt.Sprintf("Listing %q created successfully!", b.String())))
		http.Redirect(w, req, "/demo-app/listings", http.StatusFound)
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
	session = func(rtx renderCtx) projections.Session {
		var sessionID = rtx.r.Context().Value(ContextKeySessionID)
		return rtx.Sessions.Get(
			rtx.r.Context(),
			retro.PartitionName(filepath.Join("session", sessionID.(string))), // TODO: plural?
		)
	}
	urlFor = func(rtx renderCtx, rest string) string {
		return "/demo-app/" + rest
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
				"session":     session,
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

func GetFlash(w http.ResponseWriter, r *http.Request, name string) string {
	c, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	value, err := decode(c.Value)
	if err != nil {
		return ""
	}
	dc := &http.Cookie{Name: name, MaxAge: -1, Expires: time.Unix(1, 0)}
	http.SetCookie(w, dc)
	return string(value)
}

// -------------------------

func encode(src []byte) string {
	return base64.URLEncoding.EncodeToString(src)
}

func decode(src string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(src)
}
