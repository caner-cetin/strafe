package internal

import (
	"context"
	"fmt"
	"os"
	"github.com/caner-cetin/strafe/pkg/db"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/docker/docker/client"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
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

func (ctx *AppCtx) InitializeDB() error {
	ctx.Context = context.TODO()
	if !viper.IsSet(DB_URL) {
		log.Warn("database url is not set")
		return nil
	}

	conf, err := pgx.ParseConfig(viper.GetString(DB_URL))
	if err != nil {
		return fmt.Errorf("failed to parse database config: %w", err)
	}

	conn, err := pgx.ConnectConfig(ctx.Context, conf)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	ctx.DB = db.New(conn)
	ctx.Conn = conn
	ctx.StdDB = stdlib.OpenDB(*conf)
	return nil
}

func (ctx *AppCtx) Cleanup() {
	if (ctx.Conn != nil && !ctx.Conn.IsClosed()) {
		err := ctx.Conn.Close(ctx.Context)
		if err != nil {
			log.Fatal(err)
		}
		err = ctx.StdDB.Close()
		if err != nil {
			log.Fatal(err)
		}
	}
	if (ctx.Docker != nil) {
		err := ctx.Docker.Close()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (ctx *AppCtx) InitializeDocker() error {
	client, err := NewDockerClient()
	if err != nil {
		return err
	}
	ctx.Docker = client
	return nil
}

func (ctx *AppCtx) InitializeS3() error {
	if !viper.IsSet(S3_ACCESS_KEY_ID) {
		return fmt.Errorf("access key id is not set")
	}
	if !viper.IsSet(S3_ACCESS_KEY_SECRET) {
		return fmt.Errorf("access key secret is not set")
	}
	if !viper.IsSet(S3_ACCOUNT_ID) {
		return fmt.Errorf("account id is not set")
	}
	if !viper.IsSet(S3_BUCKET_NAME) {
		return fmt.Errorf("bucket name is not set")
	}
	cfg, err := config.LoadDefaultConfig(
		context.TODO(), // todo: ehhhhhh
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(viper.GetString(S3_ACCESS_KEY_ID), viper.GetString(S3_ACCESS_KEY_SECRET), "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return err
	}
	ctx.S3.Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", viper.GetString(S3_ACCOUNT_ID)))
		o.UsePathStyle = true
	})
	ctx.S3.Manager = manager.NewUploader(ctx.S3.Client, func(u *manager.Uploader) {
		u.PartSize = 5 * 1024 * 1024
		u.Concurrency = 3
		u.LeavePartsOnError = false
	})
	_, err = ctx.CreateBucketIfNotExists(context.Background(), viper.GetString(S3_BUCKET_NAME))
	if err != nil {
		return err
	}

	return nil
}

func NewDockerClient() (*client.Client, error) {
	if !viper.IsSet(DOCKER_SOCKET) {
		log.Warn("docker socket is not set, defaulting back to unix:///var/run/docker.sock")
		os.Setenv(client.DefaultDockerHost, "unix:///var/run/docker.sock")
	} else {
		os.Setenv(client.DefaultDockerHost, viper.GetString(DOCKER_SOCKET))
	}
	docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return docker, nil
}
