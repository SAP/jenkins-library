package influx

import (
	"context"
	"errors"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"golang.org/x/mod/semver"
)

var (
	errVersionNotValid     = errors.New("Influx version is invalid")
	errVersionNotSupported = errors.New("version 1.8 and higher is supported")
	errFieldsMapAssertion  = errors.New("fields map assertion failed")
	errTagsMapAssertion    = errors.New("tags map assertion failed")
)

// Client handles communication with InfluxDB
type Client struct {
	client        influxdb2.Client
	ctx           context.Context
	influxVersion string
	organozation  string
	bucket        string
}

// NewClient instantiates a Client
func NewClient(influxVersion string, serverUrl string, authToken string, organization string, bucket string) (*Client, error) {
	fmt.Println(influxVersion)
	// check InfluxDB version
	version := "v" + influxVersion
	if !semver.IsValid(version) {
		return nil, errVersionNotValid
	}
	if result := semver.Compare(version, "v1.8"); result < 0 {
		return nil, errVersionNotSupported
	}
	// In 1.8 version of InfluxDB the organization parameter is not used. It must be empty.
	if semver.MajorMinor(version) == "v1.8" {
		organization = ""
	}
	// Create a Client
	influxClient := influxdb2.NewClient(serverUrl, authToken)
	ctx := context.Background()
	client := Client{
		client:        influxClient,
		ctx:           ctx,
		influxVersion: influxVersion,
		organozation:  organization,
		bucket:        bucket,
	}

	return &client, nil
}

// WriteMetrics writes metrics to InfluxDB
func (c *Client) WriteMetrics(dataMap map[string]interface{}, dataMapTags map[string]interface{}) error {
	writeAPI := c.client.WriteAPIBlocking(c.organozation, c.bucket)

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
