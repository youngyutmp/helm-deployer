package domain

//K8SReleaseProvider interface
type K8SReleaseProvider interface {
	Start()
	GetDeployConfigsForImagePath(path string) ([]*DeployConfig, error)
}
