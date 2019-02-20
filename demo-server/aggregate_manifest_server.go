package main

import (
	"encoding/json"
	"net/http"

	"github.com/retro-framework/go-retro/framework/retro"
)

type aggregateManifestServer struct {
	m retro.AggregateManifest
}

func (ms aggregateManifestServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	if lAggregate, ok := ms.m.(retro.ListingAggregateManifest); ok {
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
