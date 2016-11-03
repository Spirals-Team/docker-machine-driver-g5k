package api

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	G5kApiFrontend = "https://api.grid5000.fr/stable"
)

type Api struct {
	client   http.Client
	Username string
	Passwd   string
	Site     string
}

type CollectionHeader struct {
	total  int `json:"total"`
	offset int `json:"offset"`
}

type Link struct {
	Relationship string `json:"rel"`
	Reference    string `json:"href"`
	Type         string `json:"type"`
}

func NewApi(username, password, site string) *Api {
	return &Api{
		Username: username,
		Passwd:   password,
		Site:     site,
	}
}

func (a *Api) get(url string) ([]byte, error) {
	return a.request("GET", url, "")
}

func (a *Api) post(url, args string) ([]byte, error) {
	return a.request("POST", url, args)
}

func (a *Api) del(url string) ([]byte, error) {
	return a.request("DELETE", url, "")
}

// Arguments are in a string. Maybe we could arrange that later
func (a *Api) request(method, url, args string) ([]byte, error) {
	// Create a request
	req, errReq := http.NewRequest(method, url, bytes.NewReader([]byte(args)))
	if errReq != nil {
		return nil, errReq
	}

	// Authentification parameters
	req.SetBasicAuth(a.Username, a.Passwd)
	// Necessary
	req.Header.Add("Accept", "*/*")
	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}

	return response(resp)
}

func response(response *http.Response) ([]byte, error) {
	defer response.Body.Close()

	// Check whether the request succeeded or not
	if response.StatusCode < 200 || response.StatusCode >= 400 {
		errorText := http.StatusText(response.StatusCode)
		return nil, fmt.Errorf("[G5K_api] request failed: %v %s", response.StatusCode, errorText)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
