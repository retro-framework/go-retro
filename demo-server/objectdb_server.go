package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/retro-framework/go-retro/framework/object"
	"github.com/retro-framework/go-retro/framework/packing"
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
		jsonEnc = json.NewEncoder(w)
		jp      = packing.NewJSONPacker()
	)

	var vars = mux.Vars(r)
	hashedObj, err := srv.db.RetrievePacked(vars["hash"])
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
}
