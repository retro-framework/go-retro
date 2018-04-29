package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	sessionsYo "github.com/retro-framework/go-retro/commands/session"
	widgetsAppYo "github.com/retro-framework/go-retro/commands/widgets_app"

	opentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"github.com/retro-framework/go-retro/aggregates"
	"github.com/retro-framework/go-retro/commands"
	"github.com/retro-framework/go-retro/events"
	"github.com/retro-framework/go-retro/framework/depot/redis"
	"github.com/retro-framework/go-retro/framework/engine"
	"github.com/retro-framework/go-retro/framework/resolver"
)

func main() {

	type foo struct {
		a widgetsAppYo.AllowCreationOfNewIdentities
		b sessionsYo.Start
	}

	collector, err := zipkin.NewHTTPCollector("http://localhost:9411/api/v1/spans")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer collector.Close()

	tracer, err := zipkin.NewTracer(
		zipkin.NewRecorder(collector, true, "0.0.0.0:8080", "example"),
	)
	if err != nil {
		log.Fatal(err)
	}
	opentracing.SetGlobalTracer(tracer)

	var (
		emd  = redis.NewDepot(events.DefaultManifest, time.Now)
		idFn = func() (string, error) { return fmt.Sprintf("%x", rand.Uint64()), nil }
		r    = resolver.New(aggregates.DefaultManifest, commands.DefaultManifest)
		e    = engine.New(emd, r.Resolve, idFn, aggregates.DefaultManifest, events.DefaultManifest)
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
		}

		sid, err := e.StartSession(req.Context())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Fprintf(w, string(sid))
	})

	s := &http.Server{
		Addr:           ":8080",
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(s.ListenAndServe())
}
