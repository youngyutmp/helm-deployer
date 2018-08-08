package domain

// Services used by the API
type Services struct {
	ChartValuesService     ChartValuesService
	ChartRepositoryService ChartRepositoryService
	ReleaseService         ReleaseService
	WebhookService         WebhookService
	WebhookCallbackService WebhookCallbackService
}
