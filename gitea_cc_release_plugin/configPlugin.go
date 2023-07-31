package gitea_cc_release_plugin

const (
	EnvPluginResultShareHost = "PLUGIN_RESULT_SHARE_HOST"

	//msgTypeText        = "text"
	//msgTypePost        = "post"
	//msgTypeInteractive = "interactive"

	EnvApiKey  = "PLUGIN_RELEASE_GITEA_API_KEY"
	NameApiKey = "config.release_gitea_api_key"

	EnvApiBaseUrl  = "PLUGIN_RELEASE_GITEA_BASE_URL"
	NameApiBaseUrl = "config.release_gitea_base_url"

	EnvFileKey  = "PLUGIN_RELEASE_GITEA_FILES"
	NameFileKey = "config.release_gitea_files"

	EnvGitRemote  = "PLUGIN_RELEASE_GIT_REMOTE"
	NameGitRemote = "config.release_git_remote"
)

var (
	//// supportMsgType
	//supportMsgType = []string{
	//	msgTypeText,
	//	msgTypePost,
	//	msgTypeInteractive,
	//}

	cleanResultEnvList = []string{
		EnvPluginResultShareHost,
	}
)

type (
	// Config plugin private config
	Config struct {
		Debug         bool
		TimeoutSecond uint

		GiteaBaseUrl      string
		GiteaApiKey       string
		GiteaReleaseFiles []string

		GitRemote string
	}
)
