package domain

import (
	"time"
)

//ReleaseUpdateRequest struct
type ReleaseUpdateRequest struct {
	Name     string `json:"name"`
	ChartURL string `json:"chartUrl"`
	Values   string `json:"values"`
}

//Release struct
type Release struct {
	Name      string              `json:"name"`
	Namespace string              `json:"namespace"`
	Version   int                 `json:"version"`
	Manifest  *string             `json:"manifest,omitempty"`
	Info      *ReleaseInfo        `json:"info,omitempty"`
	Chart     *ReleaseChart       `json:"chart,omitempty"`
	Config    *ReleaseChartConfig `json:"config,omitempty"`
}

//ReleaseInfo struct
type ReleaseInfo struct {
	Status        *ReleaseStatus `json:"status"`
	FirstDeployed time.Time      `json:"firstDeployed"`
	LastDeployed  time.Time      `json:"lastDeployed"`
	Description   string         `json:"Description"`
}

//ReleaseStatus struct
type ReleaseStatus struct {
	Status    string `json:"status"`
	Resources string `json:"resources"`
	Notes     string `json:"notes"`
}

//ReleaseChart struct
type ReleaseChart struct {
	Metadata *ReleaseChartMetadata `json:"metadata,omitempty"`
	Values   *ReleaseChartValues   `json:"values,omitempty"`
}

//ReleaseChartMetadata struct
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

//ReleaseChartValues struct
type ReleaseChartValues struct {
	Raw string `json:"raw"`
}

//ReleaseChartConfig struct
type ReleaseChartConfig struct {
	Raw string `json:"raw"`
}

//ReleaseService manages Releases
type ReleaseService interface {
	ListReleases() ([]Release, error)
	UpdateRelease(r *ReleaseUpdateRequest) error
}
