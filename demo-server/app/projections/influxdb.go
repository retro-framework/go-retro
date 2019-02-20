package projections

import (
	"context"
	"fmt"

	"github.com/influxdata/influxdb/client/v2"

	"github.com/retro-framework/go-retro/framework/depot"
	"github.com/retro-framework/go-retro/framework/retro"
)

var PublishDataToInfluxDB = func(ctx context.Context, d retro.Depot) {
	var (
		influxDBName        = "retrov1"
		everything          = d.Watch(ctx, "*")
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
			go func(evIter retro.EventIterator) {
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
						fmt.Println("projection(influx):", pt)
					}
				}
			}(everythingEvents)
		}
	}
}
