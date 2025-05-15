//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestInfluxIntegration ./integration/...

package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/SAP/jenkins-library/pkg/influx"
)

func TestInfluxIntegrationWriteMetrics(t *testing.T) {
	// t.Parallel()
	ctx := context.Background()
	const authToken = "influx-token"
	const username = "username"
	const password = "password"
	const bucket = "piper"
	const organization = "org"

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			AlwaysPullImage: true,
			Image:           "influxdb:2.0",
			ExposedPorts:    []string{"8086/tcp"},
			Env: map[string]string{
				"DOCKER_INFLUXDB_INIT_MODE":        "setup",
				"DOCKER_INFLUXDB_INIT_USERNAME":    username,
				"DOCKER_INFLUXDB_INIT_PASSWORD":    password,
				"DOCKER_INFLUXDB_INIT_ORG":         organization,
				"DOCKER_INFLUXDB_INIT_BUCKET":      bucket,
				"DOCKER_INFLUXDB_INIT_ADMIN_TOKEN": authToken,
			},
			WaitingFor: wait.ForListeningPort("8086/tcp"),
		},
		Started: true,
	}

	influxContainer, err := testcontainers.GenericContainer(ctx, req)
	require.NoError(t, err)
	defer influxContainer.Terminate(ctx)

	ip, err := influxContainer.Host(ctx)
	require.NoError(t, err)
	port, err := influxContainer.MappedPort(ctx, "8086")
	require.NoError(t, err)
	host := fmt.Sprintf("http://%s:%s", ip, port.Port())
	dataMap := map[string]map[string]interface{}{
		"series_1": {"field_a": 11, "field_b": 12},
		"series_2": {"field_c": 21, "field_d": 22},
	}
	dataMapTags := map[string]map[string]string{
		"series_1": {"tag_a": "a", "tag_b": "b"},
		"series_2": {"tag_c": "c", "tag_d": "d"},
	}

	time.Sleep(30 * time.Second)

	influxClient := influxdb2.NewClient(host, authToken)
	defer influxClient.Close()
	client := influx.NewClient(influxClient, organization, bucket)
	err = client.WriteMetrics(dataMap, dataMapTags)
	assert.NoError(t, err)

	queryAPI := influxClient.QueryAPI(organization)
	result, err := queryAPI.Query(context.Background(),
		`from(bucket:"piper")|> range(start: -1h) |> filter(fn: (r) => r._measurement == "series_1" or r._measurement == "series_2")`)
	assert.NoError(t, err)
	valuesMap := map[string]map[string]interface{}{}
	expectedValuesMap := map[string]map[string]interface{}{
		"series_1_field_a": {"_field": "field_a", "_measurement": "series_1", "_value": int64(11), "tag_a": "a", "tag_b": "b"},
		"series_1_field_b": {"_field": "field_b", "_measurement": "series_1", "_value": int64(12), "tag_a": "a", "tag_b": "b"},
		"series_2_field_c": {"_field": "field_c", "_measurement": "series_2", "_value": int64(21), "tag_c": "c", "tag_d": "d"},
		"series_2_field_d": {"_field": "field_d", "_measurement": "series_2", "_value": int64(22), "tag_c": "c", "tag_d": "d"},
	}
	for result.Next() {
		values := result.Record().Values()
		measurement := values["_measurement"]
		field := values["_field"]
		delete(values, "_start")
		delete(values, "_stop")
		delete(values, "_time")
		delete(values, "result")
		delete(values, "table")
		valuesMap[fmt.Sprintf("%v_%v", measurement, field)] = values
	}
	assert.NoError(t, result.Err())
	assert.Equal(t, len(expectedValuesMap), len(valuesMap))
	assert.Equal(t, expectedValuesMap, valuesMap)
}
