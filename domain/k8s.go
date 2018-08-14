package domain

//K8SReleaseProvider interface
type K8SReleaseProvider interface {
	Start()
	GetDeployConfigForImagePath(path string) (*DeployConfig, error)
}
