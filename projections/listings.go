package projections

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/olivere/elastic"

	"github.com/retro-framework/go-retro/events"
	"github.com/retro-framework/go-retro/framework/depot"
	"github.com/retro-framework/go-retro/framework/retro"
)

type Listing struct {
	Name string `json:"name"`
	Desc string `json:"desc"`

	Created time.Time `json:"created_at"`
	Updated time.Time `json:"updated_at"`

	Published bool `json:"published"`

	Hash string    `json:"hash"`
	URN  retro.URN `json:"urn"`
}

type Listings interface {
	PublishedListings(context.Context) ([]Listing, error)
}

func NewListings(c *elastic.Client, evManifest retro.EventManifest, d retro.Depot) esListings {
	return esListings{c, evManifest, d, "listings"}
}

type esListings struct {
	client     *elastic.Client
	evManifest retro.EventManifest
	d          retro.Depot

	indexName string
}

func (esl esListings) PublishedListings(ctx context.Context) ([]Listing, error) {
	var listings []Listing
	searchResult, err := esl.client.Search().
		Index(esl.indexName).                           // search in index "twitter"
		Query(elastic.NewTermQuery("published", true)). // specify the query
		From(0).Size(10).                               // take documents 0-9
		Do(ctx)                                         // execute
	if err != nil {
		return listings, err
	}
	if searchResult.Hits.TotalHits > 0 {
		for _, hit := range searchResult.Hits.Hits {
			var l Listing
			err := json.Unmarshal(*hit.Source, &l)
			if err != nil {
				return listings, err
			}
			listings = append(listings, l)
		}
	}
	return listings, nil
}

func (esl esListings) Run(ctx context.Context) {

	var (
		listing = esl.d.Watch(ctx, "listing/*")
		mapping = `{
			"settings": {
				"number_of_shards": 1,
				"number_of_replicas": 0
			},
			"mappings": {
				"listing": {
					"properties": {
						"created_at": { "type": "date" },
						"updated_at": { "type": "date" },
						"name": { "type": "text" },
						"desc": { "type": "text" },
						"published": { "type": "boolean" }
					}
				}
			}
		}`
	)

	exists, err := esl.client.IndexExists(esl.indexName).Do(ctx)
	if err != nil {
		fmt.Println("es-listings: err checking for existence of ES index", err)
		return
	}
	if !exists {
		createIndex, err := esl.client.CreateIndex(esl.indexName).Body(mapping).Do(ctx)
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
				var listing = Listing{
					URN: retro.URN("urn:" + evIter.Pattern()),
				}
				for {
					var pEv, err = evIter.Next(ctx)
					if err == depot.Done {
						continue
					}
					if err != nil {
						fmt.Println("es-listings: err", err)
						return
					}
					if pEv != nil {

						if listing.Created.IsZero() {
							listing.Created = pEv.Time()
						}

						listing.Updated = pEv.Time()

						listing.Hash = pEv.CheckpointHash().String()

						ev, err := esl.evManifest.ForName(pEv.Name())
						if err != nil {
							fmt.Println("err looking up event", err)
							continue
						}

						err = json.Unmarshal(pEv.Bytes(), &ev)
						if err != nil {
							fmt.Println("err unmarshalling event", err)
							continue
						}

						switch tEv := ev.(type) {
						case *events.CreateListingImage:
							// fmt.Println("skipping image event, for now")
						case *events.SetDisplayName:
							listing.Name = tEv.Name
						case *events.SetDescription:
							listing.Desc = tEv.Desc
						case *events.SetVisibility:
							listing.Published = (tEv.Radius == "public")
						default:
							// fmt.Printf("not sure what to do with %#v", tEv)
						}

						_, err = esl.client.Index().
							Index(esl.indexName).
							Type(pEv.PartitionName().Dirname()).
							Id(pEv.PartitionName().ID()).
							BodyJson(listing).
							Do(ctx)
						if err != nil {
							// Handle error
							panic(err)
						}
					}
				}
			}(listingEvents)
		}
	}
}
