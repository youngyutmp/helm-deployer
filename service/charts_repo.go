package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/entwico/helm-deployer/conf/logging"
	"github.com/entwico/helm-deployer/domain"
	"github.com/pkg/errors"
)

const apiPathCharts = "/api/charts"

//ChartRepositoryServiceImpl is in implementation of ChartRepositoryService
type ChartRepositoryServiceImpl struct {
	RepositoryBaseURL string
	HTTPClient        *http.Client
}

//NewChartRepositoryService returns a new instance of ChartRepositoryService
func NewChartRepositoryService(baseURL string) domain.ChartRepositoryService {
	return &ChartRepositoryServiceImpl{
		RepositoryBaseURL: baseURL,
		HTTPClient:        &http.Client{Timeout: 10 * time.Second},
	}
}

//FindAllCharts returns a list of helm charts
func (c *ChartRepositoryServiceImpl) FindAllCharts(ctx context.Context) (items []domain.ChartRepositoryItem, err error) {
	url := fmt.Sprintf("%s%s", c.RepositoryBaseURL, apiPathCharts)
	logger := logging.FromContext(ctx)
	logger.WithField("url", url).Debug("fetching charts list")
	r, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = r.Body.Close()
	}()

	var res map[string][]domain.ChartRepositoryItem
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		return nil, err
	}

	items = make([]domain.ChartRepositoryItem, 0)
	for _, v := range res {
		for _, i := range v {
			items = append(items, i)
		}
	}

	return items, nil
}

//GetChartData returns helm chart binary data
func (c *ChartRepositoryServiceImpl) GetChartData(ctx context.Context, chartName, chartVersion string) ([]byte, error) {
	logger := logging.FromContext(ctx)
	charts, err := c.FindAllCharts(ctx)
	if err != nil {
		return nil, err
	}
	for _, chart := range charts {
		if chart.Name == chartName && chart.Version == chartVersion {
			url := fmt.Sprintf("%s/%s", c.RepositoryBaseURL, chart.Urls[0])
			logger.WithField("url", url).Debug("downloading chart")
			resp, err := http.Get(url)
			if err != nil {
				return nil, err
			}
			data, err := ioutil.ReadAll(resp.Body)
			if err := resp.Body.Close(); err != nil {
				return nil, err
			}
			return data, err
		}
	}
	return nil, errors.New("chart not found")
}
