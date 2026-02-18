package influx

import (
	"context"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

// Client handles communication with InfluxDB
type Client struct {
	client       influxdb2.Client
	ctx          context.Context
	organization string
	bucket       string
}

// NewClient instantiates a Client
func NewClient(influxClient influxdb2.Client, organization string, bucket string) *Client {
	ctx := context.Background()
	client := Client{
		client:       influxClient,
		ctx:          ctx,
		organization: organization,
		bucket:       bucket,
	}
	return &client
}

// WriteMetrics writes metrics to InfluxDB
func (c *Client) WriteMetrics(dataMap map[string]map[string]any, dataMapTags map[string]map[string]string) error {
	writeAPI := c.client.WriteAPIBlocking(c.organization, c.bucket)

	for measurement, fields := range dataMap {
		tags := dataMapTags[measurement]
		point := influxdb2.NewPoint(measurement,
			tags,
			fields,
			time.Now())
		if err := writeAPI.WritePoint(c.ctx, point); err != nil {
			return err
		}
	}
	return nil
}
