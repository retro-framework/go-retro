package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/retro-framework/go-retro/framework/engine"
	"github.com/retro-framework/go-retro/framework/object"
	"github.com/retro-framework/go-retro/framework/packing"
	"github.com/retro-framework/go-retro/framework/ref"
	"github.com/retro-framework/go-retro/framework/types"
)

type aggregateManifestServer struct {
	m types.AggregateManifest
}

func (ms aggregateManifestServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	if lAggregate, ok := ms.m.(types.ListingAggregateManifest); ok {
		var enc = json.NewEncoder(w)
		enc.SetIndent("", "    ")
		err := enc.Encode(lAggregate.List())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	} else {
		http.Error(w, http.StatusText(501), 501)
	}
}

type commandManifestServer struct {
	m types.CommandManifest
}

func (ms commandManifestServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	if lCommand, ok := ms.m.(types.ListingCommandManifest); ok {
		var enc = json.NewEncoder(w)
		enc.SetIndent("", "    ")
		err := enc.Encode(lCommand.List())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	} else {
		http.Error(w, http.StatusText(501), 501)
	}
}

type eventManifestServer struct {
	m types.EventManifest
}

func (ms eventManifestServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	if lEvent, ok := ms.m.(types.ListingEventManifest); ok {
		var enc = json.NewEncoder(w)
		enc.SetIndent("", "    ")
		err := enc.Encode(lEvent.List())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	} else {
		http.Error(w, http.StatusText(501), 501)
	}
}

type objectDBServer struct {
	db object.DB
}

func (srv objectDBServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	var (
		parts   []string
		jsonEnc = json.NewEncoder(w)
		jp      = packing.NewJSONPacker()
	)

	for _, part := range strings.Split(r.URL.Path, "/") {
		if len(part) > 0 {
			parts = append(parts, part)
		}
	}

	switch len(parts) {
	case 2:
		hashedObj, err := srv.db.RetrievePacked(parts[1])
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch hashedObj.Type() {
		case packing.ObjectTypeCheckpoint:
			cp, _ := jp.UnpackCheckpoint(hashedObj.Contents())
			jsonEnc.Encode(cp)
		case packing.ObjectTypeAffix:
			af, _ := jp.UnpackAffix(hashedObj.Contents())
			jsonEnc.Encode(af)
		case packing.ObjectTypeEvent:
			var evPlaceholder map[string]interface{}
			evName, evEncodedString, _ := jp.UnpackEvent(hashedObj.Contents())
			err := json.Unmarshal(evEncodedString, &evPlaceholder)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error unmarshalling %s: %s\n", evEncodedString, err)
			}
			jsonEnc.Encode(struct {
				Name    string      `json:"name"`
				Payload interface{} `json:"payload"`
			}{evName, evPlaceholder})
		default:
			http.Error(w, http.StatusText(http.StatusExpectationFailed), http.StatusExpectationFailed)
			return
		}
		break
	default:
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

}

type refDBServer struct {
	db ref.DB
}

func (srv refDBServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	var (
		parts   []string
		jsonEnc = json.NewEncoder(w)
	)
	for _, part := range strings.Split(r.URL.Path, "/") {
		if len(part) > 0 {
			parts = append(parts, part)
		}
	}

	switch len(parts) {
	case 1:
		var ldb, ok = srv.db.(ref.ListableStore)
		if !ok {
			http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
			return
		}
		hashes, err := ldb.Ls()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		err = jsonEnc.Encode(hashes)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		break
	default:
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

}

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

	sid, err := e.e.StartSession(ctx)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
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
