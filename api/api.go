package api

import (
    "net/http"
    "io/ioutil"
    "fmt"
    "encoding/json"
    "bytes"
)

const (
    G5kApiFrontend = "https://api.grid5000.fr/stable"
    defaultImage = "jessie-x64-min"
)

type Api struct {
    client      http.Client
    Username    string
    passwd      string
    Site        string
}

type CollectionHeader struct {
    total   int     `json:"total"`
    offset  int     `json:"offset"`
}

type Link struct {
    Relationship    string  `json:"rel"`
    Reference       string  `json:"href"`
    Type            string  `json:"type"`
}

func NewApi(username, passwd, site string) *Api{
    return &Api{
        Username: username,
        passwd:   passwd,
        Site:     site,
    }
}

func (a *Api) GetSiteClusters() (interface{}, error) {
    var jsonClustersInfo struct {
        CollectionHeader
        Items   []struct {
            Uid     string  `json:"uid"`
        }                   `json:"items"`
    }

    // Looking for the clusters available on the site
    var urlClusters string = G5kApiFrontend + "/sites/" + a.Site + "/clusters"
    if resp, errResp := a.get(urlClusters); errResp != nil {
        return nil, errResp
    } else {
        if errJson := json.Unmarshal(resp, &jsonClustersInfo); errJson != nil {
            return nil, errJson
        }
    }

    for _, cluster := range jsonClustersInfo.Items {
        var urlNodes string = urlClusters + "/" + cluster.Uid + "/nodes"
        if resp, errResp := a.get(urlNodes); errResp != nil {
            return nil, errResp
        } else {
            var data interface{}
            if errJson := json.Unmarshal(resp, &data); errJson != nil {
                return nil, errJson
            }
        }
    }
    return "prout", nil
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
    req.SetBasicAuth(a.Username, a.passwd)
    // Necessary
    req.Header.Add("Accept", "*/*")
    if method == "POST" {
        req.Header.Set("Content-Type", "application/json")
    }

    if resp, err := a.client.Do(req); err != nil {
        return nil, err
    } else {
        return response(resp)
    }
}

func response(response *http.Response) ([]byte, error) {
    defer response.Body.Close()

    // Check whether the request succeeded or not
    if response.StatusCode < 200 || response.StatusCode >= 400 {
        errorText := http.StatusText(response.StatusCode)
        return nil, fmt.Errorf("[G5K_api] request failed: %v %s.", response.StatusCode, errorText)
    }

    body, err := ioutil.ReadAll(response.Body)
    if err != nil {
        return nil, err
    } else {
        return body, nil
    }
}
