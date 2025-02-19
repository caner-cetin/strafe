package internal

import (
	"strafe/pkg/db"

	"github.com/docker/docker/client"
	"github.com/jackc/pgx/v5"
)

type AppCtx struct {
	DB     *db.Queries
	Docker *client.Client
	Conn   *pgx.Conn
}

const (
	STRAFE_CONFIG_LOC_ENV     = "STRAFE_CFG"
	CREDENTIALS_USERNAME      = "credentials.username"
	CREDENTIALS_PASSWORD      = "credentials.password"
	DOCKER_IMAGE_NAME         = "docker.image.name"
	DOCKER_IMAGE_TAG          = "docker.image.tag"
	DOCKER_SOCKET             = "docker.socket"
	DB_URL                    = "db.url"
	DISPLAY_ASCII_ART_ON_HELP = "display_ascii_art_on_help"
)

type ConfigDefault string

const (
	DOCKER_IMAGE_NAME_DEFAULT ConfigDefault = "strafe"
	DOCKER_IMAGE_TAG_DEFAULT  ConfigDefault = "latest"
)

type ContextKey string

const (
	APP_CONTEXT_KEY ContextKey = "strafe_ctx.app"
)

var (
	TimeoutMS int
	CFGFile   string
	Verbosity int
)

type ExifInfo struct {
	SourceFile          string  `json:"SourceFile,omitempty"`
	ExifToolVersion     float64 `json:"ExifToolVersion,omitempty"`
	FileName            string  `json:"FileName,omitempty"`
	Directory           string  `json:"Directory,omitempty"`
	FileSize            string  `json:"FileSize,omitempty"`
	FileModifyDate      string  `json:"FileModifyDate,omitempty"`
	FileAccessDate      string  `json:"FileAccessDate,omitempty"`
	FileInodeChangeDate string  `json:"FileInodeChangeDate,omitempty"`
	FilePermissions     string  `json:"FilePermissions,omitempty"`
	FileType            string  `json:"FileType,omitempty"`
	FileTypeExtension   string  `json:"FileTypeExtension,omitempty"`
	MIMEType            string  `json:"MIMEType,omitempty"`
	MPEGAudioVersion    int     `json:"MPEGAudioVersion,omitempty"`
	AudioLayer          int     `json:"AudioLayer,omitempty"`
	AudioBitrate        string  `json:"AudioBitrate,omitempty"`
	SampleRate          int     `json:"SampleRate,omitempty"`
	ChannelMode         string  `json:"ChannelMode,omitempty"`
	MSStereo            string  `json:"MSStereo,omitempty"`
	IntensityStereo     string  `json:"IntensityStereo,omitempty"`
	CopyrightFlag       bool    `json:"CopyrightFlag,omitempty"`
	OriginalMedia       bool    `json:"OriginalMedia,omitempty"`
	Emphasis            string  `json:"Emphasis,omitempty"`
	ID3Size             int     `json:"ID3Size,omitempty"`
	Album               string  `json:"Album,omitempty"`
	Artist              string  `json:"Artist,omitempty"`
	PartOfSet           int     `json:"PartOfSet,omitempty"`
	Title               string  `json:"Title,omitempty"`
	Track               int     `json:"Track,omitempty"`
	Year                int     `json:"Year,omitempty"`
	Comment             string  `json:"Comment,omitempty"`
	Genre               string  `json:"Genre,omitempty"`
	DateTimeOriginal    int     `json:"DateTimeOriginal,omitempty"`
	Duration            string  `json:"Duration,omitempty"`
}
