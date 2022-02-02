package goget

import (
	"fmt"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"net/http"
	"strings"

	"github.com/antchfx/htmlquery"
)

// Client .
type Client interface {
	GetRepositoryURL(module string) (string, error)
}

// ClientImpl .
type ClientImpl struct {
	HTTPClient piperhttp.Sender
}

// GetRepositoryURL resolves the repository URL for the given go module. Only git is supported.
func (c *ClientImpl) GetRepositoryURL(module string) (string, error) {
	response, err := c.HTTPClient.SendRequest("GET", fmt.Sprintf("https://%s?go-get=1", module), nil, http.Header{}, nil)

	if err != nil {
		return "", err
	} else if response.StatusCode == 404 {
		return "", fmt.Errorf("module '%s' doesn't exist", module)
	} else if response.StatusCode != 200 {
		return "", fmt.Errorf("received unexpected response status code: %d", response.StatusCode)
	}

	html, err := htmlquery.Parse(response.Body)

	if err != nil {
		return "", fmt.Errorf("unable to parse content: %q", err)
	}

	metaNode := htmlquery.FindOne(html, "//meta[@name='go-import']/@content")

	if metaNode == nil {
		return "", fmt.Errorf("couldn't find go-import statement")
	}

	goImportStatement := htmlquery.SelectAttr(metaNode, "content")
	goImport := strings.Split(goImportStatement, " ")

	if len(goImport) != 3 || goImport[1] != "git" {
		return "", fmt.Errorf("unsupported module: '%s'", module)
	}

	return goImport[2], nil
}
