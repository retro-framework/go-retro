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

type clock struct{}

func (c clock) Now() time.Time { return time.Now().UTC() }

func main() {

	var (
		storagePath string
		listenAddr  = fmt.Sprintf(":%s", os.Getenv("PORT"))
	)

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
		zipkin.NewRecorder(collector, true, listenAddr, "go-retro-demo-server"),
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
		idFn     = func() (string, error) {
			b := make([]byte, 12)
			_, err := rand.Read(b)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("%x", b), nil
		}
		r = resolver.New(aggregates.DefaultManifest, commands.DefaultManifest)
		e = engine.New(depot, r.Resolve, idFn, clock{}, aggregates.DefaultManifest, events.DefaultManifest)
	)

	mux := http.NewServeMux()

	mux.Handle("/", handlers.CombinedLoggingHandler(os.Stdout, engineServer{e}))

	mux.Handle("/list/aggregates", handlers.CombinedLoggingHandler(os.Stdout, aggregateManifestServer{aggregates.DefaultManifest}))
	mux.Handle("/list/commands", handlers.CombinedLoggingHandler(os.Stdout, commandManifestServer{commands.DefaultManifest}))
	mux.Handle("/list/events", handlers.CombinedLoggingHandler(os.Stdout, eventManifestServer{events.DefaultManifest}))

	mux.Handle("/obj/", handlers.CombinedLoggingHandler(os.Stdout, objDBSrv))
	mux.Handle("/ref/", handlers.CombinedLoggingHandler(os.Stdout, refDBSrv))

	s := &http.Server{
		Addr:           listenAddr,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(s.ListenAndServe())
}
