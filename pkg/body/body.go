package body

import (
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
)

// ReadResponseBody reads the body of a response and returns it as a byte slice
func ReadResponseBody(response *http.Response) ([]byte, error) {
	if response == nil {
		return nil, errors.Errorf("did not retrieve an HTTP response")
	}
	if response.Body != nil {
		defer response.Body.Close()
	}
	bodyText, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		return nil, errors.Wrap(readErr, "HTTP response body could not be read")
	}
	return bodyText, nil
}
