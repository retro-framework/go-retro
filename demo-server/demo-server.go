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

	"github.com/go-redis/redis"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/olivere/elastic"

	"github.com/namsral/flag"
	opentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"

	"github.com/retro-framework/go-retro/aggregates"
	"github.com/retro-framework/go-retro/commands"
	"github.com/retro-framework/go-retro/events"

	"github.com/retro-framework/go-retro/demo-server/app"
	"github.com/retro-framework/go-retro/projections"

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

	esClient, err := elastic.NewClient(
		elastic.SetSniff(false),
		elastic.SetURL("http://localhost:9200"),
		elastic.SetErrorLog(log.New(os.Stdout, "es-listings: ", 0)),
		// elastic.SetTraceLog(log.New(os.Stdout, "es-listings: ", 0)),
	)
	if err != nil {
		fmt.Println("es-listings: err dialing ES:", err)
		return
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	var (
		listingsProjection = projections.NewListings(esClient, events.DefaultManifest, d)
		profilesProjection = projections.NewProfiles(redisClient, events.DefaultManifest, d)
		sessionsProjection = projections.NewSessions(profilesProjection, events.DefaultManifest, d)
		tsdbProjection     = projections.NewTSDB(d)
	)

	go listingsProjection.Run(ctx)
	go profilesProjection.Run(ctx)
	go sessionsProjection.Run(ctx)
	go tsdbProjection.Run(ctx)

	rMux := mux.NewRouter()

	rMux.Handle("/list/aggregates", aggregateManifestServer{aggregates.DefaultManifest}).Methods("GET")
	rMux.Handle("/list/commands", commandManifestServer{commands.DefaultManifest}).Methods("GET")
	rMux.Handle("/list/events", eventManifestServer{events.DefaultManifest}).Methods("GET")
	rMux.Handle("/obj/{hash}", objDBSrv).Methods("GET")
	rMux.Handle("/ref/", refDBSrv).Methods("GET")
	rMux.Handle("/apply", engineServer{e}).Methods("POST")

	var (
		appMount = "/demo-app"
		demoApp  = app.NewServer(e, sessionsProjection, templatePath, appMount)
		profile  = demoApp.NewProfileServer(profilesProjection)
		listing  = demoApp.NewListingServer(listingsProjection)
	)
	var aMux = rMux.PathPrefix(appMount).Subrouter()
	aMux.HandleFunc("", demoApp.IndexHandler).Methods("GET")
	aMux.HandleFunc("/", demoApp.IndexHandler).Methods("GET")
	aMux.HandleFunc("/listings", listing.ListHandler).Methods("GET")
	aMux.HandleFunc("/listing/new", listing.NewHandler).Methods("GET", "POST")
	aMux.HandleFunc("/profiles/{name}", profile.ShowHandler).Methods("GET")
	aMux.HandleFunc("/profile/new", profile.NewHandler).Methods("GET", "POST")
	aMux.HandleFunc("/profile", demoApp.IndexHandler).Methods("POST")
	aMux.Use(retroSessionMiddleware{e}.Middleware)

	s := &http.Server{
		Addr:           listenAddr,
		Handler:        handlers.CombinedLoggingHandler(os.Stdout, rMux),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

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
