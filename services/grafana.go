package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/dliakhov/pmm/grafana/backup/model"
)

const (
	limit = 5000
)

type grafanaServiceImpl struct {
	httpClient   *http.Client
	grafanaURL   string
	grafanaToken string
}

type GrafanaService interface {
	GetAllDashboards() (model.DashboardSearch, error)
	GetDashboard(uid string) (map[string]any, error)
	CreateOrUpdateDashboards(d map[string]any) error
}

func NewGrafanaService(httpClient *http.Client, grafanaURL, grafanaToken string) *grafanaServiceImpl {
	return &grafanaServiceImpl{
		httpClient:   httpClient,
		grafanaURL:   grafanaURL,
		grafanaToken: grafanaToken,
	}
}

func (g *grafanaServiceImpl) GetAllDashboards() (model.DashboardSearch, error) {
	resp, err := g.getGrafanaData("search")
	if err != nil {
		return nil, err
	}
	ds := model.DashboardSearch{}
	err = json.Unmarshal(resp, &ds)
	if err != nil {
		return nil, err
	}
	return ds, nil
}

func (g *grafanaServiceImpl) GetDashboard(uid string) (map[string]any, error) {
	db, err := g.getGrafanaData(fmt.Sprintf("dashboards/uid/%s", uid))
	if err != nil {
		return nil, err
	}
	var dashJSON map[string]interface{}
	err = json.Unmarshal(db, &dashJSON)
	if err != nil {
		return nil, err
	}
	return dashJSON, nil
}

func (g *grafanaServiceImpl) getGrafanaData(endpoint string) ([]byte, error) {
	req, err := http.NewRequest("GET", g.grafanaURL+endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", g.grafanaToken))
	req.Header.Set("Content-Type", "application/json")
	params := req.URL.Query()
	params.Add("limit", strconv.Itoa(limit))
	req.URL.RawQuery = params.Encode()

	response, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		fmt.Println("response.StatusCode: ", response.StatusCode)
		return nil, fmt.Errorf("[Error] %s", response.Status)
	}

	read, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return read, nil
}

func (g *grafanaServiceImpl) CreateOrUpdateDashboards(d map[string]any) error {
	err := g.postGrafanaData("dashboards/db", d)
	if err != nil {
		return err
	}
	return nil
}

func (g *grafanaServiceImpl) postGrafanaData(endpoint string, body any) error {
	bytesBody, err := json.Marshal(&body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", g.grafanaURL+endpoint, bytes.NewBuffer(bytesBody))
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", g.grafanaToken))
	req.Header.Set("Content-Type", "application/json")
	params := req.URL.Query()
	params.Add("limit", strconv.Itoa(limit))
	req.URL.RawQuery = params.Encode()

	response, err := g.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		_, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("[Error] %s", response.Status)
	}

	_, err = io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	return nil
}
