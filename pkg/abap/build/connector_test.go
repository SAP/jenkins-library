//go:build unit
// +build unit

package build

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateUrl(t *testing.T) {
	//arrange global
	conn := new(Connector)
	conn.MaxRuntime = time.Duration(1 * time.Second)
	conn.PollingInterval = time.Duration(1 * time.Microsecond)
	conn.Baseurl = "/BUILD/CORE_SRV"
	t.Run("Zero Parameter", func(t *testing.T) {
		//act
		url := conn.createUrl("/builds('123456789')")
		//assert
		assert.Equal(t, "/BUILD/CORE_SRV/builds('123456789')", url)
	})
	t.Run("One Parameter", func(t *testing.T) {
		//arange
		conn.Parameters = url.Values{}
		abapSourceClient := "001"
		conn.Parameters.Add("sap-client", abapSourceClient)
		//act
		url := conn.createUrl("/builds('123456789')")

		//assert
		assert.Equal(t, "/BUILD/CORE_SRV/builds('123456789')?sap-client=001", url)
	})
	t.Run("Two Parameters", func(t *testing.T) {
		//arrange
		conn.Parameters = url.Values{}
		conn.Parameters.Add("sap-client", "001")
		conn.Parameters.Add("format", "json")
		//act
		url := conn.createUrl("/builds('123456789')")

		//assert
		assert.Equal(t, "/BUILD/CORE_SRV/builds('123456789')?format=json&sap-client=001", url)
	})
	t.Run("Three Parameters", func(t *testing.T) {
		//arrange
		conn.Parameters = url.Values{}
		conn.Parameters.Add("sap-client", "001")
		conn.Parameters.Add("format", "json")
		conn.Parameters.Add("top", "2")

		//act
		url := conn.createUrl("/builds('123456789')")

		//assert
		assert.Equal(t, "/BUILD/CORE_SRV/builds('123456789')?format=json&sap-client=001&top=2", url)
	})
}
