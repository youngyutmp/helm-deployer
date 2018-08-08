package service

import "time"

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
