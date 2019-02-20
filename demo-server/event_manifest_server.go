package main

import (
	"encoding/json"
	"net/http"

	"github.com/retro-framework/go-retro/framework/retro"
)

type eventManifestServer struct {
	m retro.EventManifest
}

func (ms eventManifestServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	if lEvent, ok := ms.m.(retro.ListingEventManifest); ok {
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
