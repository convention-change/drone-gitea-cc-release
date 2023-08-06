package gitea_cc_release_plugin

const (
	EnvPluginResultShareHost = "PLUGIN_RESULT_SHARE_HOST"

	NamePluginDebug   = "config.debug"
	EnvPluginTimeOut  = "PLUGIN_TIMEOUT_SECOND"
	NamePluginTimeOut = "config.timeout_second"

	//msgTypeText        = "text"
	//msgTypePost        = "post"
	//msgTypeInteractive = "interactive"

	EnvDraft  = "PLUGIN_DRAFT"
	NameDraft = "config.draft"

	EnvPrerelease  = "PLUGIN_PRERELEASE"
	NamePrerelease = "config.prerelease"

	EnvRootFolderPath  = "PLUGIN_RELEASE_GITEA_ROOT_FOLDER_PATH"
	NameRootFolderPath = "config.release_gitea_root_folder_path"

	EnvApiBaseUrl  = "PLUGIN_RELEASE_GITEA_BASE_URL"
	NameApiBaseUrl = "config.release_gitea_base_url"

	EnvGiteaInsecure  = "PLUGIN_RELEASE_GITEA_INSECURE"
	NameGiteaInsecure = "config.release_gitea_insecure"

	EnvGiteaApiKey  = "PLUGIN_RELEASE_GITEA_API_KEY"
	NameGiteaApiKey = "config.release_gitea_api_key"

	EnvReleaseFiles  = "PLUGIN_RELEASE_GITEA_FILES"
	NameReleaseFiles = "config.release_gitea_files"

	EnvFilesChecksum  = "PLUGIN_RELEASE_GITEA_FILES_CHECKSUM"
	NameFilesChecksum = "config.release_gitea_files_checksum"

	EnvFileExistsDo  = "PLUGIN_RELEASE_GITEA_FILE_EXISTS_DO"
	NameFileExistsDo = "config.release_gitea_file_exists_do"

	FileExistsDoOverwrite = "overwrite"
	FileExistsDoFail      = "fail"
	FileExistsDoSkip      = "skip"

	EnvTitle  = "PLUGIN_RELEASE_GITEA_TITLE"
	NameTitle = "config.release_gitea_title"

	EnvNote  = "PLUGIN_RELEASE_GITEA_NOTE"
	NameNote = "config.release_gitea_note"

	EnvNoteByConventionChange  = "PLUGIN_RELEASE_GITEA_NOTE_BY_CONVENTION_CHANGE"
	NameNoteByConventionChange = "config.release_gitea_note_by_convention_change"

	EnvReadChangeLogFile  = "PLUGIN_RELEASE_READ_CHANGE_LOG_FILE"
	NameReadChangeLogFile = "config.release_read_change_log_file"

	EnvGitRemote  = "PLUGIN_RELEASE_GIT_REMOTE"
	NameGitRemote = "config.release_git_remote"
)

var (
	// supportFileExistsDoList
	supportFileExistsDoList = []string{
		FileExistsDoOverwrite,
		FileExistsDoFail,
		FileExistsDoSkip,
	}

	cleanResultEnvList = []string{
		EnvPluginResultShareHost,
	}
)

type (
	// Config plugin private config
	Config struct {
		Debug         bool
		TimeoutSecond uint

		RootFolderPath string

		GiteaBaseUrl  string
		GiteaInsecure bool
		GiteaApiKey   string

		GiteaReleaseFileGlobs        []string
		GiteaReleaseFileGlobRootPath string
		FilesChecksum                []string

		GiteaFileExistsDo string
		GiteaDraft        bool
		GiteaPrerelease   bool
		GiteaTitle        string
		GiteaNote         string

		NoteByConventionChange bool
		ReadChangeLogFile      string

		GitRemote string
	}
)
