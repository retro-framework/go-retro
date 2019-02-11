package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/retro-framework/go-retro/framework/object"
	"github.com/retro-framework/go-retro/framework/packing"
	"github.com/retro-framework/go-retro/framework/ref"
)

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

	time.Sleep(time.Duration(rand.Intn(750)) * time.Millisecond)

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
		case "checkpoint":
			cp, _ := jp.UnpackCheckpoint(hashedObj.Contents())
			jsonEnc.Encode(cp)
		case "affix":
			af, _ := jp.UnpackAffix(hashedObj.Contents())
			jsonEnc.Encode(af)
		case "event":
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
