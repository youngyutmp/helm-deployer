package domain

import (
	"context"
	"time"
)

//ChartRepositoryItem struct
type ChartRepositoryItem struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	APIVersion  string    `json:"apiVersion"`
	Urls        []string  `json:"urls"`
	Created     time.Time `json:"created"`
	Digest      string    `json:"digest"`
}

//ChartRepositoryService interface
type ChartRepositoryService interface {
	FindAllCharts(ctx context.Context) ([]ChartRepositoryItem, error)
	GetChartData(ctx context.Context, chartName, chartVersion string) ([]byte, error)
}
