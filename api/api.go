package api

import (
	"github.com/go-resty/resty"
)

const (
	// G5kAPIFrontend is the link to the Grid'5000 API frontend
	G5kAPIFrontend = "https://api.grid5000.fr/stable"
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

// Request returns a configured resty request
func (c *Client) Request() *resty.Request {
	return resty.R().
		SetHeader("Accept", "application/json").
		SetBasicAuth(c.Username, c.Password)
}
