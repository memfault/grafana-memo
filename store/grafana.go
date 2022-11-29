package store

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/grafana/memo"
	log "github.com/sirupsen/logrus"
)

// Grafana
type Grafana struct {
	// apiKey generated in Grafana's settings, service accounts work too
	apiKey string
	// apiUrl is your grafana instance URI with /api appended
	// e.g. http://localhost/api/
	apiUrl string

	// tlsKey for when you are using self signed certificates
	tlsKey string
	// tlsCert
	tlsCert string

	// bearerHeader is an internal cache of the header with the apiKey present
	bearerHeader string
	// apiUrlAnnotations is an internal cache of the url for the API
	apiUrlAnnotations string
	// apiUrlHealth is an internal cache of the url for the API health page
	apiUrlHealth string
}

// NewGrafana returns a new grafana instance
func NewGrafana(apiKey, apiUrl, tlsKey, tlsCert string) (Grafana, error) {
	u, err := url.Parse(apiUrl)
	if err != nil {
		return Grafana{}, err
	}

	urlAnnotations := *u
	urlAnnotations.Path = path.Join(u.Path, "annotations")

	urlHealth := *u
	urlHealth.Path = path.Join(u.Path, "health")

	g := Grafana{
		apiKey:  apiKey,
		apiUrl:  apiUrl,
		tlsKey:  tlsKey,
		tlsCert: tlsCert,

		bearerHeader:      fmt.Sprintf("Bearer %s", apiKey),
		apiUrlAnnotations: urlAnnotations.String(),
		apiUrlHealth:      urlHealth.String(),
	}
	return g, nil
}

// GrafanaHealthResp
type GrafanaHealthResp struct {
	// Commit
	Commit string
	// Database
	Database string
	// Version
	Version string
}

// httpClient returns the client for communicating with Grafana
func (g Grafana) httpClient() (*http.Client, error) {
	client := &http.Client{}

	if g.tlsKey != "" || g.tlsCert != "" {
		// Load client cert
		cert, err := tls.LoadX509KeyPair(g.tlsCert, g.tlsKey)
		if err != nil {
			return nil, err
		}
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		transport := &http.Transport{TLSClientConfig: tlsConfig}
		client.Transport = transport
	}
	return client, nil
}

// Check ensures the API is healthy
func (g Grafana) Check() error {
	client, err := g.httpClient()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("GET", g.apiUrlHealth, nil)
	if err != nil {
		return fmt.Errorf("grafana creation of request failed: %s", err)
	}

	req.Header.Set("Authorization", g.bearerHeader)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("grafana health check fail: %s", err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("grafana failed to read body: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Grafana replied with http %d and body %s", resp.StatusCode, string(data))
	}

	var gaResp GrafanaHealthResp
	err = json.Unmarshal(data, &gaResp)
	if err != nil {
		return fmt.Errorf("grafana failed to unmarshal grafana response: %s. The body was: %s", err, string(data))
	}
	log.Infof("Can talk to Grafana version %s - its database is %s", gaResp.Version, gaResp.Database)
	return nil
}

// GrafanaAnnotationReq
type GrafanaAnnotationReq struct {
	// Time unix ts in ms
	Time int64 `json:"time"`
	// IsRegion
	IsRegion bool `json:"isRegion"`
	// Tags
	Tags []string `json:"tags"`
	// Text
	Text string `json:"text"`
}

// GrafanaAnnotationResp
type GrafanaAnnotationResp struct {
	// Message
	Message string `json:"message"`
	// Id
	Id int `json:"id"`
	// EndId
	EndId int `json:"endId"`
}

// Save stores the memo in the API
func (g Grafana) Save(memo memo.Memo) error {
	ga := GrafanaAnnotationReq{
		Time:     memo.Date.Unix() * 1000,
		IsRegion: false,
		Tags:     memo.Tags,
		Text:     memo.Desc,
	}
	jsonValue, _ := json.Marshal(ga)

	client, err := g.httpClient()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", g.apiUrlAnnotations, bytes.NewBuffer(jsonValue))
	if err != nil {
		return fmt.Errorf("grafana creation of request failed: %s", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", g.bearerHeader)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("grafana post fail: %s", err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("grafana failed to read body: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Grafana replied with http %d and body %s", resp.StatusCode, string(data))
	}

	var gaResp GrafanaAnnotationResp
	err = json.Unmarshal(data, &gaResp)
	if err != nil {
		return fmt.Errorf("grafana failed to unmarshal grafana response: %s. The body was: %s", err, string(data))
	}
	if gaResp.Message != "Annotation added" {
		return fmt.Errorf("Grafana replied with http %d and unexpected message %q", resp.StatusCode, gaResp.Message)
	}
	return nil
}
