package protecode

import (
	"io"
)

const (
	statusBusy   = "B"
	statusReady  = "R"
	statusFailed = "F"
)

func (pc *Protecode) send(method string, url string, headers map[string][]string) (*io.ReadCloser, error) {
	r, err := pc.client.SendRequest(method, url, nil, headers, nil)
	if err != nil {
		return nil, err
	}
	return &r.Body, nil
}
