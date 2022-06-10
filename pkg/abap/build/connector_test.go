package build

import (
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
		abapSourceClient := "001"
		conn.Parameters = []string{"sap-client=" + abapSourceClient}
		//act
		url := conn.createUrl("/builds('123456789')")

		//assert
		assert.Equal(t, "/BUILD/CORE_SRV/builds('123456789')?sap-client=001", url)
	})
	t.Run("Two Parameters", func(t *testing.T) {
		//arrange
		configParameters := make([]string, 2)
		configParameters[0] = "sap-client=001"
		configParameters[1] = "$format=json"
		conn.Parameters = configParameters
		//act
		url := conn.createUrl("/builds('123456789')")

		//assert
		assert.Equal(t, "/BUILD/CORE_SRV/builds('123456789')?sap-client=001&$format=json", url)
	})
	t.Run("Three Parameters", func(t *testing.T) {
		//arrange
		configParameters := make([]string, 3)
		configParameters[0] = "sap-client=001"
		configParameters[1] = "$format=json"
		configParameters[2] = "$top=2"
		conn.Parameters = configParameters
		//act
		url := conn.createUrl("/builds('123456789')")

		//assert
		assert.Equal(t, "/BUILD/CORE_SRV/builds('123456789')?sap-client=001&$format=json&$top=2", url)
	})
}
