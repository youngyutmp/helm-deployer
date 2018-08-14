package service

import (
	"io"

	"bytes"

	"github.com/Sirupsen/logrus"
	"github.com/entwico/helm-deployer/domain"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/services"
)

//helmServiceImpl is an implementation of HelmService interface
type helmServiceImpl struct {
	client                 *helm.Client
	chartValuesService     domain.ChartValuesService
	chartRepositoryService domain.ChartRepositoryService
}

//NewHelmService returns a new instance of HelmService
func NewHelmService(client *helm.Client, chartValuesService domain.ChartValuesService, chartRepositoryService domain.ChartRepositoryService) domain.HelmService {
	return &helmServiceImpl{
		client:                 client,
		chartValuesService:     chartValuesService,
		chartRepositoryService: chartRepositoryService,
	}
}

//ListReleases returns all Helm releases
func (s *helmServiceImpl) ListReleases() (*services.ListReleasesResponse, error) {
	response, err := s.client.ListReleases()
	if err != nil {
		return nil, err
	}
	return response, nil
}

//UpdateRelease updates helm release
func (s *helmServiceImpl) UpdateRelease(rlsName string, chartData io.Reader, rawVals []byte) (*services.UpdateReleaseResponse, error) {
	chart, err := chartutil.LoadArchive(chartData)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("chart %s loaded", chart.Metadata.Name)
	logrus.Debugf("updating release %s", rlsName)
	return s.client.UpdateReleaseFromChart(rlsName, chart, helm.UpdateValueOverrides(rawVals), helm.UpgradeForce(true), helm.UpgradeRecreate(true))
}

//DeployChart deploys the helm chart
func (s *helmServiceImpl) DeployChart(cfg domain.DeployConfig) error {
	logrus.Debugf("deploying chart %s %s", cfg.ChartName, cfg.ChartVersion)

	var rawVals []byte
	if cfg.ChartValuesID != nil {
		values, err := s.chartValuesService.FindOne(*cfg.ChartValuesID)
		if err != nil {
			return err
		}
		if values != nil {
			rawVals = []byte(values.Data)
		}
	}

	data, err := s.chartRepositoryService.GetChartData(cfg.ChartName, cfg.ChartVersion)
	if err != nil {
		return err
	}

	_, err = s.UpdateRelease(cfg.ReleaseName, bytes.NewReader(data), rawVals)
	return err
}
