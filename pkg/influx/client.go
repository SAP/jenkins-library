package influx

import (
	"context"
	"errors"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

var (
	errFieldsMapAssertion = errors.New("fields map assertion failed")
	errTagsMapAssertion   = errors.New("tags map assertion failed")
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
func (c *Client) WriteMetrics(dataMap map[string]interface{}, dataMapTags map[string]interface{}) error {
	writeAPI := c.client.WriteAPIBlocking(c.organization, c.bucket)

	for measurement, fields := range dataMap {
		fieldsMap, ok := fields.(map[string]interface{})
		if !ok {
			return errFieldsMapAssertion
		}
		tagsMapString := map[string]string{}
		if tags, ok := dataMapTags[measurement]; ok {
			if tagsMap, ok := tags.(map[string]interface{}); !ok {
				return errTagsMapAssertion
			} else {
				for key, value := range tagsMap {
					tagsMapString[key] = fmt.Sprintf("%s", value)
				}
			}
		}
		point := influxdb2.NewPoint(measurement,
			tagsMapString,
			fieldsMap,
			time.Now())
		if err := writeAPI.WritePoint(c.ctx, point); err != nil {
			return err
		}
	}
	return nil
}
