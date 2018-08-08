package service

import (
	"io"

	"github.com/entwico/helm-deployer/domain"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/services"
)

//HelmServiceImpl is an implementation of HelmService interface
type HelmServiceImpl struct {
	Client *helm.Client
}

//NewHelmService returns a new instance of HelmService
func NewHelmService(client *helm.Client) domain.HelmService {
	return &HelmServiceImpl{
		Client: client,
	}
}

//ListReleases returns all Helm releases
func (c *HelmServiceImpl) ListReleases() (*services.ListReleasesResponse, error) {
	response, err := c.Client.ListReleases()
	if err != nil {
		return nil, err
	}
	return response, nil
}

//UpdateRelease updates helm release
func (c *HelmServiceImpl) UpdateRelease(rlsName string, chartData io.Reader, rawVals []byte) (*services.UpdateReleaseResponse, error) {
	chart, err := chartutil.LoadArchive(chartData)
	if err != nil {
		return nil, err
	}
	return c.Client.UpdateReleaseFromChart(rlsName, chart, helm.UpdateValueOverrides(rawVals), helm.UpgradeForce(true), helm.UpgradeRecreate(true))
}
