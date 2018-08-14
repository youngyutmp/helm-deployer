package domain

import (
	"io"

	"k8s.io/helm/pkg/proto/hapi/services"
)

//HelmService interface
type HelmService interface {
	ListReleases() (*services.ListReleasesResponse, error)
	UpdateRelease(rlsName string, chartData io.Reader, rawVals []byte) (*services.UpdateReleaseResponse, error)
	DeployChart(cfg DeployConfig) error
}
