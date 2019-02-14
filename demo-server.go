package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/influxdata/influxdb/client/v2"
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
	"github.com/retro-framework/go-retro/framework/types"

	_ "github.com/retro-framework/go-retro/commands/identity"
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
		d        = depot.NewSimple(odb, refdb)
		idFn     = func() (string, error) {
			b := make([]byte, 12)
			_, err := rand.Read(b)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("%x", b), nil
		}
		r = resolver.New(aggregates.DefaultManifest, commands.DefaultManifest)
		e = engine.New(d, r.Resolve, idFn, clock{}, aggregates.DefaultManifest, events.DefaultManifest)
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

	go func() {
		var (
			influxDBName        = "retrov1"
			ctx                 = context.Background()
			everything          = d.Glob(ctx, "*")
			influxHTTPClient, _ = client.NewHTTPClient(client.HTTPConfig{
				Addr: "http://localhost:8086",
			})
		)
		res, err := influxHTTPClient.Query(client.NewQuery("CREATE DATABASE "+influxDBName, "", ""))
		if err != nil {
			fmt.Println("err creating database", err)
		}
		if res.Error() != nil {
			fmt.Println("err creating database", res.Error())
		}
		for {
			everythingEvents, err := everything.Next(ctx)
			if err == depot.Done {
				continue
			}
			if err != nil {
				fmt.Println("err", err)
				return
			}
			if everythingEvents != nil {
				go func(evIter types.EventIterator) {
					for {
						var ev, err = evIter.Next(ctx)
						if err == depot.Done {
							continue
						}
						if err != nil {
							fmt.Println("err", err)
							return
						}
						if ev != nil {
							bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
								Database:  influxDBName,
								Precision: "s",
							})
							var tags = map[string]string{
								"name":      ev.Name(),
								"partition": evIter.Pattern(),
							}
							var fields = map[string]interface{}{"count": 1}
							pt, err := client.NewPoint("events", tags, fields, ev.Time())
							if err != nil {
								fmt.Println("Error: ", err.Error())
							}
							bp.AddPoint(pt)
							err = influxHTTPClient.Write(bp)
							if err != nil {
								fmt.Println("err writing to influxdb", err)
							}
						}
					}
				}(everythingEvents)
			}
		}
	}()

	log.Fatal(s.ListenAndServe())
}
