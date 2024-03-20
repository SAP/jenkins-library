//go:build unit
// +build unit

package build

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"golang.org/x/exp/slices" //in 1.21 will be a standard package "slices"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/stretchr/testify/assert"
)

type HeaderVerifyingMockClient struct {
	Header map[string][]string
}

func (mc *HeaderVerifyingMockClient) SetOptions(opts piperhttp.ClientOptions) {}
func (mc *HeaderVerifyingMockClient) SendRequest(Method, Url string, bdy io.Reader, hdr http.Header, cookies []*http.Cookie) (*http.Response, error) {
	for requiredHeaderKey, requiredHeaderValues := range mc.Header {
		suppliedHeaderValues, existingHeader := hdr[requiredHeaderKey]
		if existingHeader {
			for _, element := range requiredHeaderValues {
				existingValue := slices.Contains(suppliedHeaderValues, element)
				if !existingValue {
					return nil, fmt.Errorf("header %s does not contain expected value %s", requiredHeaderKey, element)
				}
			}
		} else {
			return nil, fmt.Errorf("Expected header %s not part of the http request", requiredHeaderKey)
		}
	}

	return &http.Response{Body: io.NopCloser(bytes.NewReader([]byte("")))}, nil
}

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

func TestInitAAKaaSHeader(t *testing.T) {
	conn := new(Connector)

	client := HeaderVerifyingMockClient{}
	client.Header = make(map[string][]string)
	client.Header["Accept"] = []string{"application/json"}
	client.Header["Content-Type"] = []string{"application/json"}
	client.Header["User-Agent"] = []string{"Piper-abapAddonAssemblyKit/1.0"}
	t.Run("InitAAKaaS success no hash", func(t *testing.T) {
		conn.InitAAKaaS("endpoint", "user", "pw", &client, "", "", "")
		_, err := conn.Get("something")
		assert.NoError(t, err)
	})
	t.Run("InitAAKaaS success with hash", func(t *testing.T) {
		client.Header["build-config-token"] = []string{"hash"}
		conn.InitAAKaaS("endpoint", "user", "pw", &client, "hash", "", "")
		_, err := conn.Get("something")
		assert.NoError(t, err)
	})
	t.Run("InitAAKaaS sanity check Header", func(t *testing.T) {
		client.Header["FAIL"] = []string{"verify HeaderVerifyingMockClient works"}
		conn.InitAAKaaS("endpoint", "user", "pw", &client, "hash", "", "")
		_, err := conn.Get("something")
		assert.Error(t, err)
	})
	t.Run("InitAAKaaS sanity check wrong Value in existing Header", func(t *testing.T) {
		client.Header["Accept"] = []string{"verify HeaderVerifyingMockClient works"}
		conn.InitAAKaaS("endpoint", "user", "pw", &client, "hash", "", "")
		_, err := conn.Get("something")
		assert.Error(t, err)
	})
}
