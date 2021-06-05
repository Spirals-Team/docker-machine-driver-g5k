package api

import (
	"net/url"
	gopath "path"

	"github.com/go-resty/resty/v2"
)

const (
	g5kAPIhostname string = "api.grid5000.fr"
	g5kAPIversion  string = "sid"
)

// Client is a client to the Grid'5000 REST API
type Client struct {
	caller  *resty.Client
	baseURL url.URL
}

// NewClient returns a new configured Grid'5000 API client
func NewClient(username, password, site string) *Client {
	caller := resty.New().
		SetHeader("Accept", "application/json").
		SetBasicAuth(username, password)

	baseURL := url.URL{
		Scheme: "https",
		Host:   g5kAPIhostname,
		Path:   gopath.Join(g5kAPIversion, "sites", site),
	}

	return &Client{caller, baseURL}
}

// getEndpoint construct and returns the API endpoint for the given api name and path
func (c *Client) getEndpoint(api string, path string, params url.Values) string {
	endpoint := c.baseURL
	endpoint.Path = gopath.Join(endpoint.Path, api, path)
	endpoint.RawQuery = params.Encode()
	return endpoint.String()
}
