package api

import (
	"net/url"
	gopath "path"

	"github.com/go-resty/resty"
)

const (
	g5kAPIhostname string = "api.grid5000.fr"
	g5kAPIversion  string = "4.0"
)

// Client is a client to the Grid'5000 REST API
type Client struct {
	Username string
	Password string
	Site     string
}

// NewClient returns a new configured Grid'5000 API client
func NewClient(username, password, site string) *Client {
	return &Client{
		Username: username,
		Password: password,
		Site:     site,
	}
}

// getRequest returns a configured resty request
func (c *Client) getRequest() *resty.Request {
	return resty.R().
		SetHeader("Accept", "application/json").
		SetBasicAuth(c.Username, c.Password)
}

// getBaseURL returns the Grid'5000 API base url
func (c *Client) getBaseURL() *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   g5kAPIhostname,
		Path:   gopath.Join(g5kAPIversion, "sites", c.Site),
	}
}

// getEndpoint construct and returns the API endpoint for the given api name and path
func (c *Client) getEndpoint(api string, path string, params url.Values) string {
	url := c.getBaseURL()
	url.Path = gopath.Join(url.Path, api, path)
	url.RawQuery = params.Encode()
	return url.String()
}
