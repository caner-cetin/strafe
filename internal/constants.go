package internal

const (
	STRAFE_CONFIG_LOC_ENV = "STRAFE_CFG"
	CREDENTIALS_USERNAME  = "credentials.username"
	CREDENTIALS_PASSWORD  = "credentials.password"
	DOCKER_IMAGE_NAME     = "docker.image.name"
	DOCKER_IMAGE_TAG      = "docker.image.tag"
	DOCKER_SOCKET         = "docker.socket"
	DB_URL                = "db.url"
)

const (
	DOCKER_IMAGE_NAME_DEFAULT = "strafe"
	DOCKER_IMAGE_TAG_DEFAULT  = "latest"
)

var (
	TimeoutMS int
	CFGFile   string
	Verbosity int
)

type ContextKey string

const (
	APP_CONTEXT_KEY ContextKey = "strafe_ctx.app"
)
