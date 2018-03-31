package service

import (
	"net/http"
	"sort"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/ptypes"
)

type ReleaseUpdateRequest struct {
	Name     string `json:"name"`
	ChartURL string `json:"chartUrl"`
	Values   string `json:"values"`
}

type Release struct {
	Name      string              `json:"name"`
	Namespace string              `json:"namespace"`
	Version   int                 `json:"version"`
	Manifest  *string             `json:"manifest,omitempty"`
	Info      *ReleaseInfo        `json:"info,omitempty"`
	Chart     *ReleaseChart       `json:"chart,omitempty"`
	Config    *ReleaseChartConfig `json:"config,omitempty"`
}

type ReleaseInfo struct {
	Status        *ReleaseStatus `json:"status"`
	FirstDeployed time.Time      `json:"firstDeployed"`
	LastDeployed  time.Time      `json:"lastDeployed"`
	Description   string         `json:"Description"`
}

type ReleaseStatus struct {
	Status    string `json:"status"`
	Resources string `json:"resources"`
	Notes     string `json:"notes"`
}

type ReleaseChart struct {
	Metadata *ReleaseChartMetadata `json:"metadata,omitempty"`
	Values   *ReleaseChartValues   `json:"values,omitempty"`
}

type ReleaseChartMetadata struct {
	// The name of the chart
	Name string `json:"name,omitempty"`
	// The URL to a relevant project page, git repo, or contact person
	Home string `json:"home,omitempty"`
	// A SemVer 2 conformant version string of the chart
	Version string `json:"version,omitempty"`
	// A one-sentence description of the chart
	Description string `json:"description,omitempty"`
	// A list of string keywords
	Keywords []string `json:"keywords,omitempty"`
	// The URL to an icon file.
	Icon string `json:"icon,omitempty"`
	// The API Version of this chart.
	APIVersion string `json:"apiVersion,omitempty"`
	// The tags to check to enable chart
	Tags string `json:"tags,omitempty"`
	// The version of the application enclosed inside of this chart.
	AppVersion string `json:"appVersion,omitempty"`
	// Whether or not this chart is deprecated
	Deprecated bool `json:"deprecated,omitempty"`
	// Annotations are additional mappings uninterpreted by Tiller,
	// made available for inspection by other applications.
	Annotations map[string]string `json:"annotations,omitempty"`
	// KubeVersion is  a SemVer constraints on what version of Kubernetes is required.
	KubeVersion string `json:"kubeVersion,omitempty"`
}

type ReleaseChartValues struct {
	Raw string `json:"raw"`
}

type ReleaseChartConfig struct {
	Raw string `json:"raw"`
}

type ReleaseService interface {
	ListReleases() ([]Release, error)
	UpdateRelease(r *ReleaseUpdateRequest) error
}

type ReleaseServiceImpl struct {
	HelmService HelmService
}

func NewReleaseService(helmService HelmService) *ReleaseServiceImpl {
	return &ReleaseServiceImpl{
		HelmService: helmService,
	}
}

func (c *ReleaseServiceImpl) ListReleases() ([]Release, error) {
	logrus.Debug("Getting all releases")

	r, err := c.HelmService.ListReleases()
	if err != nil {
		return nil, err
	}
	var releases []Release
	for _, item := range r.Releases {
		firstDeployed, _ := ptypes.Timestamp(item.Info.FirstDeployed)
		lastDeployed, _ := ptypes.Timestamp(item.Info.LastDeployed)
		release := Release{
			Name:      item.Name,
			Namespace: item.Namespace,
			Version:   int(item.Version),
			Info: &ReleaseInfo{
				Status: &ReleaseStatus{
					Status:    item.Info.Status.GetCode().String(),
					Resources: item.Info.Status.GetResources(),
					Notes:     item.Info.Status.GetNotes(),
				},
				FirstDeployed: firstDeployed,
				LastDeployed:  lastDeployed,
				Description:   item.Info.Description,
			},
			Chart: &ReleaseChart{
				Metadata: &ReleaseChartMetadata{
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

func (c *ReleaseServiceImpl) UpdateRelease(r *ReleaseUpdateRequest) error {
	logrus.Debugf("Updating release %s", r.Name)
	response, err := http.Get(r.ChartURL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	_, err = c.HelmService.UpdateRelease(r.Name, response.Body, []byte(r.Values))
	return err
}
