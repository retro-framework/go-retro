package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/retro-framework/go-retro/framework/ref"
)

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
