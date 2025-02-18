package internal

import (
	"context"
	"fmt"
	"os"
	"strafe/pkg/db"
	"time"

	"github.com/docker/docker/client"
	"github.com/fatih/color"
	"github.com/jackc/pgx/v5"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ResourceType int

const (
	ResourceDatabase ResourceType = iota
	ResourceDocker
)

type ResourceConfig struct {
	Resources []ResourceType
	Timeout   *time.Duration
}

type AppCtx struct {
	DB     *db.Queries
	Docker *client.Client
	Conn   *pgx.Conn
}

func WrapCommandWithResources(fn func(cmd *cobra.Command, args []string), config ResourceConfig) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		log.Tracef("running command %s", cmd.Name())
		var to time.Duration
		if TimeoutMS == 0 {
			to = time.Millisecond * time.Duration(*config.Timeout)
		} else {
			to = time.Millisecond * time.Duration(TimeoutMS)
		}
		ctx, cancel := context.WithTimeout(context.Background(), to)
		defer cancel()

		appCtx := AppCtx{}

		for _, resource := range config.Resources {
			switch resource {
			case ResourceDatabase:
				if err := initializeDB(ctx, &appCtx); err != nil {
					log.Errorf("failed to initialize database: %v", err)
					return
				}
			case ResourceDocker:
				if err := initializeDocker(&appCtx); err != nil {
					log.Errorf("failed to initialize docker: %v", err)
					return
				}
			}
		}
		defer func() {
			if appCtx.Conn != nil {
				if err := appCtx.Conn.Close(ctx); err != nil {
					log.Errorf("failed to close database connection: %v", err)
				}
			}
			if appCtx.Docker != nil {
				appCtx.Docker.Close()
			}
		}()
		cmd.SetContext(context.WithValue(ctx, APP_CONTEXT_KEY, appCtx))
		fn(cmd, args)
	}
}

func check(err error) {
	if err != nil {
		log.Fatal(color.RedString(err.Error()))
	}
}

func NewDockerClient() *client.Client {
	os.Setenv(client.DefaultDockerHost, viper.GetString(DOCKER_SOCKET))
	docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	cobra.CheckErr(err)
	return docker
}

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
		InitLogging()
		log.Debugf("using config file: %s", viper.ConfigFileUsed())
		setDefaultConfigs()
	} else {
		fmt.Printf("Error: cannot load config file: %v\n", err)
		os.Exit(1)
	}
}
func InitLogging() {
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
}

func setDefaultConfigs() {
	viper.SetDefault(DOCKER_IMAGE_NAME, DOCKER_IMAGE_NAME_DEFAULT)
	viper.SetDefault(DOCKER_IMAGE_TAG, DOCKER_IMAGE_TAG_DEFAULT)
}

func initializeDB(ctx context.Context, appCtx *AppCtx) error {
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

func initializeDocker(appCtx *AppCtx) error {
	client := NewDockerClient()
	appCtx.Docker = client
	return nil
}
