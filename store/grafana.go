package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/raintank/memo"
	log "github.com/sirupsen/logrus"
)

type Grafana struct {
	apiKey string
	apiUrl string // e.g. http://localhost/api/

	bearerHeader      string
	apiUrlAnnotations string
	apiUrlHealth      string
}

func NewGrafana(apiKey, apiUrl string) (Grafana, error) {
	u, err := url.Parse(apiUrl)
	if err != nil {
		return Grafana{}, err
	}

	urlAnnotations := *u
	urlAnnotations.Path = path.Join(u.Path, "annotations")

	urlHealth := *u
	urlHealth.Path = path.Join(u.Path, "health")

	g := Grafana{
		apiKey: apiKey,
		apiUrl: apiUrl,

		bearerHeader:      fmt.Sprintf("Bearer %s", apiKey),
		apiUrlAnnotations: urlAnnotations.String(),
		apiUrlHealth:      urlHealth.String(),
	}
	return g, g.checkHealth()
}

type GrafanaHealthResp struct {
	Commit   string
	Database string
	Version  string
}

func (g Grafana) checkHealth() error {
	req, err := http.NewRequest("GET", g.apiUrlHealth, nil)
	if err != nil {
		return fmt.Errorf("grafana creation of request failed: %s", err)
	}

	req.Header.Set("Authorization", g.bearerHeader)

	resp, err := http.DefaultClient.Do(req)
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

type GrafanaAnnotationReq struct {
	Time     int64    `json:"time"` // unix ts in ms
	IsRegion bool     `json:"isRegion"`
	Tags     []string `json:"tags"`
	Text     string   `json:"text"`
}

type GrafanaAnnotationResp struct {
	Message string `json:"message"`
	Id      int    `json:"id"`
	EndId   int    `json:"endId"`
}

func (g Grafana) Save(memo memo.Memo) error {
	ga := GrafanaAnnotationReq{
		Time:     memo.Date.Unix() * 1000,
		IsRegion: false,
		Tags:     memo.Tags,
		Text:     memo.Desc,
	}
	jsonValue, _ := json.Marshal(ga)

	req, err := http.NewRequest("POST", g.apiUrlAnnotations, bytes.NewBuffer(jsonValue))
	if err != nil {
		return fmt.Errorf("grafana creation of request failed: %s", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", g.bearerHeader)

	resp, err := http.DefaultClient.Do(req)
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
