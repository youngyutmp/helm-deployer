package domain

//WebhookCallbackService listens to Webhooks and deploys charts
type WebhookCallbackService interface {
	ProcessWebhook(webhookType string, webhookBody []byte) error
	DeployChart(cfg DeployConfig) error
}
