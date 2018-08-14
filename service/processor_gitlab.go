package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/entwico/helm-deployer/domain"
)

const (
	headerWebhookGitlab = "X-Gitlab-Event"

	gitlabEventTypePipeline = "Pipeline Hook"
)

type gitlabWebhookProcessor struct {
	events         chan domain.DeployConfig
	webhookService domain.WebhookService
}

//NewGitlabProcessor returns new instance of Gitlab webhook processor
func NewGitlabProcessor(webhookService domain.WebhookService) domain.WebhookProcessor {
	return &gitlabWebhookProcessor{webhookService: webhookService}
}

//CanProcess returns true if webhook can be processed by this processor
func (p *gitlabWebhookProcessor) CanProcess(headers http.Header, body []byte) bool {
	val := headers.Get(headerWebhookGitlab)
	if val != "" {
		return true
	}
	return false
}

//Process handles webhook
func (p *gitlabWebhookProcessor) Process(headers http.Header, body []byte) error {
	logrus.Info("processing Gitlab webhook")
	event := headers.Get(headerWebhookGitlab)

	switch event {
	case gitlabEventTypePipeline:
		return p.processPipelineEvent(body)
	}
	return fmt.Errorf("event '%s' not supported", event)
}

func (p *gitlabWebhookProcessor) GetDeployConfigEvents() chan domain.DeployConfig {
	return p.events
}

func (p *gitlabWebhookProcessor) processPipelineEvent(body []byte) error {
	logrus.Debug("processing pipeline event")
	payload := new(WebhookGitlabPipeline)
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}

	if payload.ObjectAttributes.Status == "success" {
		cond := domain.GitlabWebhookCondition{
			WebhookType:      payload.ObjectKind,
			ProjectName:      payload.Project.Name,
			ProjectNamespace: payload.Project.Namespace,
			GitRef:           payload.ObjectAttributes.Ref,
			IsTag:            payload.ObjectAttributes.Tag,
		}

		if err := p.processCondition(cond); err != nil {
			logrus.Error(err)
			return err
		}
	} else {
		logrus.Infof(fmt.Sprintf("skipping pipeline status '%s'", payload.ObjectAttributes.Status))
	}
	return nil
}

func (p *gitlabWebhookProcessor) processCondition(cond domain.GitlabWebhookCondition) error {
	dc, err := p.getDeployConfigs(cond)
	if err != nil {
		return err
	}
	for _, cfg := range dc {
		p.events <- cfg
	}
	return nil
}

func (p *gitlabWebhookProcessor) getDeployConfigs(cond domain.GitlabWebhookCondition) ([]domain.DeployConfig, error) {
	webhooks, err := p.webhookService.FindAll()
	if err != nil {
		return nil, err
	}
	var dc []domain.DeployConfig
	for _, w := range webhooks {
		if w.Condition == cond {
			dc = append(dc, w.DeployConfig)
		}
	}

	return dc, nil
}

//WebhookGitlabPipeline struct
type WebhookGitlabPipeline struct {
	Builds []struct {
		ArtifactsFile struct {
			Filename interface{} `json:"filename"`
			Size     int         `json:"size"`
		} `json:"artifacts_file"`
		CreatedAt  string `json:"created_at"`
		FinishedAt string `json:"finished_at"`
		ID         int    `json:"id"`
		Manual     bool   `json:"manual"`
		Name       string `json:"name"`
		Runner     struct {
			Active      bool   `json:"active"`
			Description string `json:"description"`
			ID          int    `json:"id"`
			IsShared    bool   `json:"is_shared"`
		} `json:"runner"`
		Stage     string `json:"stage"`
		StartedAt string `json:"started_at"`
		Status    string `json:"status"`
		User      struct {
			AvatarURL string `json:"avatar_url"`
			Name      string `json:"name"`
			Username  string `json:"username"`
		} `json:"user"`
		When string `json:"when"`
	} `json:"builds"`
	Commit struct {
		Author struct {
			Email string `json:"email"`
			Name  string `json:"name"`
		} `json:"author"`
		ID        string    `json:"id"`
		Message   string    `json:"message"`
		Timestamp time.Time `json:"timestamp"`
		URL       string    `json:"url"`
	} `json:"commit"`
	ObjectAttributes struct {
		BeforeSha  string   `json:"before_sha"`
		CreatedAt  string   `json:"created_at"`
		Duration   int      `json:"duration"`
		FinishedAt string   `json:"finished_at"`
		ID         int      `json:"id"`
		Ref        string   `json:"ref"`
		Sha        string   `json:"sha"`
		Stages     []string `json:"stages"`
		Status     string   `json:"status"`
		Tag        bool     `json:"tag"`
	} `json:"object_attributes"`
	ObjectKind string `json:"object_kind"`
	Project    struct {
		AvatarURL         string      `json:"avatar_url"`
		CiConfigPath      interface{} `json:"ci_config_path"`
		DefaultBranch     string      `json:"default_branch"`
		Description       string      `json:"description"`
		GitHTTPURL        string      `json:"git_http_url"`
		GitSSHURL         string      `json:"git_ssh_url"`
		ID                int         `json:"id"`
		Name              string      `json:"name"`
		Namespace         string      `json:"namespace"`
		PathWithNamespace string      `json:"path_with_namespace"`
		VisibilityLevel   int         `json:"visibility_level"`
		WebURL            string      `json:"web_url"`
	} `json:"project"`
	User struct {
		AvatarURL string `json:"avatar_url"`
		Name      string `json:"name"`
		Username  string `json:"username"`
	} `json:"user"`
}
