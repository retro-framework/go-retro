package projections

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/olivere/elastic"

	"github.com/retro-framework/go-retro/framework/depot"
	"github.com/retro-framework/go-retro/framework/retro"
)

var mapping = `
{
	"settings": {
		"number_of_shards": 1,
		"number_of_replicas": 0
	},
	"mappings": {
		"listing": {
			"properties": {
				"created_at": {
					"type": "date"
				},
				"updated_at": {
					"type": "date"
				},
				"name": {
					"type": "text",
					"store": true
				},
				"description": {
					"type": "text",
					"store": true
				},
				"published": {
					"type": "boolean"
				}
			}
		}
	}
}`

var PublishListingsToElasticSearch = func(ctx context.Context, d retro.Depot) {

	var (
		indexName = "listings"
		listing   = d.Watch(ctx, "listing/*")
	)

	client, err := elastic.NewClient(
		elastic.SetSniff(false),
		elastic.SetURL("http://localhost:9200"),
		elastic.SetErrorLog(log.New(os.Stdout, "es-listings: ", 0)),
		// elastic.SetTraceLog(log.New(os.Stdout, "", 0)),
	)
	if err != nil {
		fmt.Println("es-listings: err dialing ES:", err)
		return
	}

	exists, err := client.IndexExists(indexName).Do(context.Background())
	if err != nil {
		fmt.Println("es-listings: err checking for existence of ES index", err)
		return
	}
	if !exists {
		createIndex, err := client.CreateIndex(indexName).Body(mapping).Do(context.Background())
		if err != nil {
			fmt.Println("es-listings: err creating ES index", err)
			return
		}
		if !createIndex.Acknowledged {
			fmt.Println("es-listings: ES did not acknowledge creation of index")
			return
		}
	}

	for {
		listingEvents, err := listing.Next(ctx)
		if err == depot.Done {
			continue
		}
		if err != nil {
			fmt.Println("err on pi", err)
			return
		}
		if listingEvents != nil {
			go func(evIter retro.EventIterator) {
				for {
					var ev, err = evIter.Next(ctx)
					if err == depot.Done {
						continue
					}
					if err != nil {
						fmt.Println("es-listings: err", err)
						return
					}
					if ev != nil {
						_ = ev
						// fmt.Println("es-listings: ev", ev)
					}
				}
			}(listingEvents)
		}
	}
}
