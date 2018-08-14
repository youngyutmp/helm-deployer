package domain

// Services used by the API
type Services struct {
	ChartValuesService     ChartValuesService
	ChartRepositoryService ChartRepositoryService
	HelmService            HelmService
	K8SReleaseProvider     K8SReleaseProvider
	ReleaseService         ReleaseService
	WebhookService         WebhookService
	WebhookDispatcher      WebhookDispatcher
}
