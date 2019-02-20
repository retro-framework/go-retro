package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/namsral/flag"
	opentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"

	"github.com/retro-framework/go-retro/aggregates"
	"github.com/retro-framework/go-retro/commands"
	"github.com/retro-framework/go-retro/events"

	"github.com/retro-framework/go-retro/demo-server/app"
	"github.com/retro-framework/go-retro/demo-server/app/projections"

	"github.com/retro-framework/go-retro/framework/depot"
	"github.com/retro-framework/go-retro/framework/engine"
	"github.com/retro-framework/go-retro/framework/repository"
	"github.com/retro-framework/go-retro/framework/resolver"
	"github.com/retro-framework/go-retro/framework/retro"
	"github.com/retro-framework/go-retro/framework/storage/fs"

	_ "github.com/retro-framework/go-retro/commands/identity"
	_ "github.com/retro-framework/go-retro/commands/listing"
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

	var ctx = context.Background()

	flag.StringVar(&storagePath, "storage_path", "/tmp", "storage dir for the depot")
	flag.Parse()

	storagePath, err := filepath.Abs(storagePath)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Using Storage Path:", storagePath)

	templatePath, err := filepath.Abs("./app/tpl/")
	log.Println("Using Template Path:", templatePath)

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
		d        = depot.NewSimple(odb, refdb)
		r        = repository.NewSimpleRepository(odb, refdb, events.DefaultManifest)
		_ps      = make(map[string]app.Projection)
		idFn     = func() (string, error) {
			b := make([]byte, 12)
			_, err := rand.Read(b)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("%x", b), nil
		}
		rFn = resolver.New(aggregates.DefaultManifest, commands.DefaultManifest)
		e   = engine.New(d, r, rFn, idFn, clock{}, aggregates.DefaultManifest, events.DefaultManifest)
	)

	rMux := mux.NewRouter()

	rMux.Handle("/list/aggregates", aggregateManifestServer{aggregates.DefaultManifest}).Methods("GET")
	rMux.Handle("/list/commands", commandManifestServer{commands.DefaultManifest}).Methods("GET")
	rMux.Handle("/list/events", eventManifestServer{events.DefaultManifest}).Methods("GET")
	rMux.Handle("/obj/{hash}", objDBSrv).Methods("GET")
	rMux.Handle("/ref/", refDBSrv).Methods("GET")
	rMux.Handle("/apply", engineServer{e}).Methods("POST")

	var (
		appMount = "/demo-app"
		demoApp  = app.NewServer(e, templatePath, appMount, _ps)
		profile  = demoApp.NewProfileServer()
		listing  = demoApp.NewListingServer()
	)
	var aMux = rMux.PathPrefix(appMount).Subrouter()
	aMux.HandleFunc("", demoApp.IndexHandler).Methods("GET")
	aMux.HandleFunc("/", demoApp.IndexHandler).Methods("GET")
	aMux.HandleFunc("/listing/new", listing.NewHandler).Methods("GET", "POST")
	aMux.HandleFunc("/profiles/{name}", profile.ShowHandler).Methods("GET")
	aMux.HandleFunc("/profile", demoApp.IndexHandler).Methods("POST")
	aMux.Use(retroSessionMiddleware{e}.Middleware)

	// Prints mounted routes
	//
	// rMux.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
	// 	tpl, err1 := route.GetPathTemplate()
	// 	met, err2 := route.GetMethods()
	// 	fmt.Println(tpl, err1, met, err2)
	// 	return nil
	// })

	s := &http.Server{
		Addr:           listenAddr,
		Handler:        handlers.CombinedLoggingHandler(os.Stdout, rMux),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go projections.PublishDataToInfluxDB(ctx, d)
	go projections.PublishListingsToElasticSearch(ctx, d)

	log.Fatal(s.ListenAndServe())
}

// retroSessionMiddleware ensures that a started session is injected
// into all requests.
type retroSessionMiddleware struct {
	e engine.Engine
}

func (rsm retroSessionMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			err error
			sid retro.SessionID
		)

		var ctx = r.Context()

		if sessionCookie, _ := r.Cookie("retroSessionID"); sessionCookie == nil {
			sid, err = rsm.e.StartSession(ctx)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			cookie := http.Cookie{
				Name:    "retroSessionID",
				Value:   string(sid),
				Path:    "/demo-app/",
				Expires: time.Now().Add(6 * time.Hour),
			}
			http.SetCookie(w, &cookie)
		} else {
			sid = retro.SessionID(sessionCookie.Value)
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(ctx, app.ContextKeySessionID, string(sid))))
	})
}
