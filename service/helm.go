package service

import (
	"io"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/services"
)

type HelmService interface {
	ListReleases() (*services.ListReleasesResponse, error)
	UpdateRelease(rlsName string, chartData io.Reader, rawVals []byte) (*services.UpdateReleaseResponse, error)
}

type HelmServiceImpl struct {
	Client *helm.Client
}

func NewHelmService(client *helm.Client) *HelmServiceImpl {
	return &HelmServiceImpl{
		Client: client,
	}
}

func (c *HelmServiceImpl) ListReleases() (*services.ListReleasesResponse, error) {
	response, err := c.Client.ListReleases()
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *HelmServiceImpl) UpdateRelease(rlsName string, chartData io.Reader, rawVals []byte) (*services.UpdateReleaseResponse, error) {
	chart, err := chartutil.LoadArchive(chartData)
	if err != nil {
		return nil, err
	}
	return c.Client.UpdateReleaseFromChart(rlsName, chart, helm.UpdateValueOverrides(rawVals), helm.UpgradeForce(true))
}