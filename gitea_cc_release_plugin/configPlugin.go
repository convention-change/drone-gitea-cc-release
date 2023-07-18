package gitea_cc_release_plugin

const (
	EnvPluginResultShareHost = "PLUGIN_RESULT_SHARE_HOST"

	//msgTypeText        = "text"
	//msgTypePost        = "post"
	//msgTypeInteractive = "interactive"
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

		GiteaApiKey       string
		GiteaReleaseFiles []string
	}
)
