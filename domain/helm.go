package domain

import (
	"context"
	"io"

	"k8s.io/helm/pkg/proto/hapi/services"
)

//HelmService interface
type HelmService interface {
	ListReleases(ctx context.Context) (*services.ListReleasesResponse, error)
	UpdateRelease(ctx context.Context, rlsName string, chartData io.Reader, rawVals []byte) (*services.UpdateReleaseResponse, error)
	DeployChart(ctx context.Context, cfg DeployConfig) error
}
