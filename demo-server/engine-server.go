package main

import (
	"io/ioutil"
	"net/http"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/retro-framework/go-retro/framework/engine"
	"github.com/retro-framework/go-retro/framework/retro"
)

type engineServer struct {
	e engine.Engine
}

func (e engineServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	if req.URL.Path != "/apply" {
		http.NotFound(w, req)
		return
	}

	var ctx = req.Context()

	spnApply, ctx := opentracing.StartSpanFromContext(ctx, "/apply")
	defer spnApply.Finish()

	// Check if we have a session cookie, if not we'll get one and
	// set it into the response
	var (
		err error
		sid retro.SessionID
	)
	if sessionCookie, _ := req.Cookie("retroSessionID"); sessionCookie == nil {
		sid, err = e.e.StartSession(ctx)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		cookie := http.Cookie{
			Name:    "retroSessionID",
			Value:   string(sid),
			Expires: time.Now().Add(6 * time.Hour),
		}
		http.SetCookie(w, &cookie)
	} else {
		sid = retro.SessionID(sessionCookie.Value)
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "error reading request body", 500)
		return
	}
	spnApply.SetTag("payload", string(body))

	_, err = e.e.Apply(ctx, w, sid, body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

}
