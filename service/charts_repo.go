package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

const apiPathCharts = "/api/charts"

type ChartRepositoryItem struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	APIVersion  string    `json:"apiVersion"`
	Urls        []string  `json:"urls"`
	Created     time.Time `json:"created"`
	Digest      string    `json:"digest"`
}

type ChartRepositoryService interface {
	FindAllCharts() ([]ChartRepositoryItem, error)
	GetChartData(chartName, chartVersion string) ([]byte, error)
}

type ChartRepositoryServiceImpl struct {
	RepositoryBaseUrl string
	HttpClient        *http.Client
}

func NewChartRepositoryService(baseUrl string) *ChartRepositoryServiceImpl {
	return &ChartRepositoryServiceImpl{
		RepositoryBaseUrl: baseUrl,
		HttpClient:        &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *ChartRepositoryServiceImpl) FindAllCharts() ([]ChartRepositoryItem, error) {
	url := fmt.Sprintf("%s%s", c.RepositoryBaseUrl, apiPathCharts)
	logrus.Debugf("Fetching charts list from %s", url)
	r, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var res map[string][]ChartRepositoryItem
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		return nil, err
	}

	items := []ChartRepositoryItem{}
	for _, v := range res {
		for _, i := range v {
			items = append(items, i)
		}
	}

	return items, nil
}

func (c *ChartRepositoryServiceImpl) GetChartData(chartName, chartVersion string) ([]byte, error) {
	charts, err := c.FindAllCharts()
	if err != nil {
		return nil, err
	}
	for _, chart := range charts {
		if chart.Name == chartName && chart.Version == chartVersion {
			url := fmt.Sprintf("%s/%s", c.RepositoryBaseUrl, chart.Urls[0])
			logrus.Debugf("Downloading chart from %s", url)
			resp, err := http.Get(url)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
			return ioutil.ReadAll(resp.Body)
		}
	}
	return nil, errors.New("chart not found")
}
