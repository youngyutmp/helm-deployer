package service

import (
	"net/http"
	"sort"

	"github.com/Sirupsen/logrus"
	"github.com/entwico/helm-deployer/domain"
	"github.com/golang/protobuf/ptypes"
)

//ReleaseServiceImpl is an implementation of ReleaseService
type ReleaseServiceImpl struct {
	HelmService domain.HelmService
}

//NewReleaseService returns a new instance of ReleaseService
func NewReleaseService(helmService domain.HelmService) domain.ReleaseService {
	return &ReleaseServiceImpl{
		HelmService: helmService,
	}
}

//ListReleases returns a list of all installed releases
func (c *ReleaseServiceImpl) ListReleases() ([]domain.Release, error) {
	logrus.Debug("Getting all releases")

	r, err := c.HelmService.ListReleases()
	if err != nil {
		return nil, err
	}
	var releases []domain.Release
	for _, item := range r.Releases {
		firstDeployed, _ := ptypes.Timestamp(item.Info.FirstDeployed)
		lastDeployed, _ := ptypes.Timestamp(item.Info.LastDeployed)
		release := domain.Release{
			Name:      item.Name,
			Namespace: item.Namespace,
			Version:   int(item.Version),
			Info: &domain.ReleaseInfo{
				Status: &domain.ReleaseStatus{
					Status:    item.Info.Status.GetCode().String(),
					Resources: item.Info.Status.GetResources(),
					Notes:     item.Info.Status.GetNotes(),
				},
				FirstDeployed: firstDeployed,
				LastDeployed:  lastDeployed,
				Description:   item.Info.Description,
			},
			Chart: &domain.ReleaseChart{
				Metadata: &domain.ReleaseChartMetadata{
					Name:        item.Chart.Metadata.Name,
					Home:        item.Chart.Metadata.Home,
					Version:     item.Chart.Metadata.Version,
					Description: item.Chart.Metadata.Description,
					Keywords:    item.Chart.Metadata.Keywords,
					Icon:        item.Chart.Metadata.Icon,
					APIVersion:  item.Chart.Metadata.ApiVersion,
					Tags:        item.Chart.Metadata.Tags,
					AppVersion:  item.Chart.Metadata.AppVersion,
					Deprecated:  item.Chart.Metadata.Deprecated,
					Annotations: item.Chart.Metadata.Annotations,
					KubeVersion: item.Chart.Metadata.KubeVersion,
				},
			},
		}
		releases = append(releases, release)
	}
	sort.Slice(releases, func(i, j int) bool {
		return releases[i].Info.LastDeployed.After(releases[j].Info.LastDeployed)
	})
	return releases, err
}

//UpdateRelease updates helm release
func (c *ReleaseServiceImpl) UpdateRelease(r *domain.ReleaseUpdateRequest) error {
	logrus.Debugf("Updating release %s", r.Name)
	response, err := http.Get(r.ChartURL)
	if err != nil {
		return err
	}

	_ = response.Body.Close()
	_, err = c.HelmService.UpdateRelease(r.Name, response.Body, []byte(r.Values))
	return err
}
