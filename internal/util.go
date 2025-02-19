package internal

import (
	"context"
	"fmt"
	"os"
	"strafe/pkg/db"

	"github.com/docker/docker/client"
	"github.com/jackc/pgx/v5"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// initConfig reads in config file and ENV variables if set.
func InitConfig() {
	if CFGFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(CFGFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		if os.Getenv(STRAFE_CONFIG_LOC_ENV) != "" {
			viper.AddConfigPath(os.Getenv(STRAFE_CONFIG_LOC_ENV))
		}
		viper.SetConfigType("yaml")
		viper.SetConfigName(".strafe")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Debugf("using config file: %s", viper.ConfigFileUsed())

		viper.SetDefault(DOCKER_IMAGE_NAME, DOCKER_IMAGE_NAME_DEFAULT)
		viper.SetDefault(DOCKER_IMAGE_TAG, DOCKER_IMAGE_TAG_DEFAULT)
		viper.SetDefault(DISPLAY_ASCII_ART_ON_HELP, true)

		switch Verbosity {
		case 1:
			log.SetLevel(log.InfoLevel)
		case 2:
			log.SetLevel(log.DebugLevel)
		case 3:
			log.SetLevel(log.TraceLevel)
		default:
			log.SetLevel(log.WarnLevel)
		}
	} else {
		fmt.Printf("Error: cannot load config file: %v\n", err)
		os.Exit(1)
	}
}

func InitializeDB(ctx context.Context, appCtx *AppCtx) error {
	if !viper.IsSet(DB_URL) {
		log.Warn("database url is not set")
		return nil
	}

	conf, err := pgx.ParseConfig(viper.GetString(DB_URL))
	if err != nil {
		return fmt.Errorf("failed to parse database config: %w", err)
	}

	conn, err := pgx.ConnectConfig(ctx, conf)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	appCtx.DB = db.New(conn)
	appCtx.Conn = conn
	return nil
}

func InitializeDocker(appCtx *AppCtx) error {
	client := NewDockerClient()
	appCtx.Docker = client
	return nil
}

func NewDockerClient() *client.Client {
	os.Setenv(client.DefaultDockerHost, viper.GetString(DOCKER_SOCKET))
	docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	cobra.CheckErr(err)
	return docker
}
