package service

import (
	"bytes"
	"context"
	"io"

	"github.com/entwico/helm-deployer/conf/logging"
	"github.com/entwico/helm-deployer/domain"
	log "github.com/sirupsen/logrus"
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
func (s *helmServiceImpl) ListReleases(ctx context.Context) (*services.ListReleasesResponse, error) {
	response, err := s.client.ListReleases()
	if err != nil {
		return nil, err
	}
	return response, nil
}

//UpdateRelease updates helm release
func (s *helmServiceImpl) UpdateRelease(ctx context.Context, rlsName string, chartData io.Reader, rawVals []byte) (*services.UpdateReleaseResponse, error) {
	logger := logging.FromContext(ctx)
	chart, err := chartutil.LoadArchive(chartData)
	if err != nil {
		return nil, err
	}
	logger.WithField("chart_name", chart.Metadata.Name).Debug("chart loaded")
	logger.WithField("release", rlsName).Debug("updating release")
	return s.client.UpdateReleaseFromChart(rlsName, chart, helm.UpdateValueOverrides(rawVals), helm.UpgradeForce(true), helm.UpgradeRecreate(true))
}

//DeployChart deploys the helm chart
func (s *helmServiceImpl) DeployChart(ctx context.Context, cfg domain.DeployConfig) error {
	logger := logging.FromContext(ctx)
	logger.WithFields(log.Fields{
		"chart_name":    cfg.ChartName,
		"chart_version": cfg.ChartVersion,
	}).Debug("deploying chart")

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

	data, err := s.chartRepositoryService.GetChartData(ctx, cfg.ChartName, cfg.ChartVersion)
	if err != nil {
		return err
	}

	_, err = s.UpdateRelease(ctx, cfg.ReleaseName, bytes.NewReader(data), rawVals)
	return err
}
