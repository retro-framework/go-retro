package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"testing"

	"github.com/leehambley/ls-cms/events"
)

func Test_Event_Envelope_Marshal(t *testing.T) {

	var evs []events.Event = []events.Event{
		&events.AllowCreateIdentities{},
		&events.SetAvatar{
			"image/gif",
			[]byte("DEADBEEF"),
		},
		&events.SetDisplayName{"mr. meeseeks"},
	}

	var out bytes.Buffer

	for _, ev := range evs {
		b, err := json.Marshal(ev)
		if err != nil {
			log.Fatal("err enveloping the ev", err)
		}
		out.Write(b)
		if ev != evs[len(evs)-1] {
			out.WriteByte('\n')
		}
	}

	fmt.Println(out.String())

}

func Test_Event_Envelope_Unmarshal(t *testing.T) {

	// var es = []string{
	// 	`{"type":"SetDisplayName","ev":{"Name":"mr. meeseeks"}}`,
	// }
	//
	// _, err := events.UnpackJSON([]byte(es[0]))
	// if err != nil {
	// 	log.Fatalf("err unpacking json event envelope %s", err)
	// }

	fmt.Println(events.Manifest)

}
