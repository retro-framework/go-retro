package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/namsral/flag"
	opentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"

	"github.com/retro-framework/go-retro/aggregates"
	"github.com/retro-framework/go-retro/commands"
	"github.com/retro-framework/go-retro/events"

	"github.com/retro-framework/go-retro/framework/depot"
	"github.com/retro-framework/go-retro/framework/engine"
	"github.com/retro-framework/go-retro/framework/resolver"
	"github.com/retro-framework/go-retro/framework/storage/fs"
)

import (
	_ "github.com/retro-framework/go-retro/commands/session"
	_ "github.com/retro-framework/go-retro/commands/widgets_app"
)

func main() {

	var storagePath string

	flag.StringVar(&storagePath, "storage_path", "/tmp", "storage dir for the depot")
	flag.Parse()

	log.Println("Using Storage Path:", storagePath)

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
		odb   = &fs.ObjectStore{BasePath: storagePath}
		refdb = &fs.RefStore{BasePath: storagePath}

		objDBSrv = objectDBServer{odb}
		refDBSrv = refDBServer{refdb}
		depot    = depot.NewSimple(odb, refdb)
		idFn     = func() (string, error) { return fmt.Sprintf("%x", rand.Uint64()), nil }
		r        = resolver.New(aggregates.DefaultManifest, commands.DefaultManifest)
		e        = engine.New(depot, r.Resolve, idFn, aggregates.DefaultManifest, events.DefaultManifest)
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

		resStr, err := e.Apply(req.Context(), sid, []byte(`{"path":"foo bar", "name":"dummyCmd", "params": {...}}`))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Fprintf(w, string(sid))
		fmt.Fprintf(w, string(resStr))
	})

	mux.Handle("/list/aggregates", handlers.CombinedLoggingHandler(os.Stdout, aggregateManifestServer{aggregates.DefaultManifest}))
	mux.Handle("/list/commands", handlers.CombinedLoggingHandler(os.Stdout, commandManifestServer{commands.DefaultManifest}))
	mux.Handle("/list/events", handlers.CombinedLoggingHandler(os.Stdout, eventManifestServer{events.DefaultManifest}))

	mux.Handle("/obj/", handlers.CombinedLoggingHandler(os.Stdout, objDBSrv))
	mux.Handle("/ref/", handlers.CombinedLoggingHandler(os.Stdout, refDBSrv))

	s := &http.Server{
		Addr:           fmt.Sprintf(":%s", os.Getenv("PORT")),
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(s.ListenAndServe())
}
