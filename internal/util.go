package internal

import (
	"context"
	"fmt"
	"os"
	"strafe/pkg/db"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/docker/docker/client"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
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
	client, err := NewDockerClient()
	if err != nil {
		return err
	}
	appCtx.Docker = client
	return nil
}

func InitializeS3(appCtx *AppCtx) error {
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
	appCtx.S3.Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", viper.GetString(S3_ACCOUNT_ID)))
	})
	return nil
}

func (a *AppCtx) CreateBucket(ctx context.Context, name string, region string) error {
	_, err := a.S3.Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(name),
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(region),
		},
	})
	if err != nil {
		var owned *types.BucketAlreadyOwnedByYou
		var exists *types.BucketAlreadyExists
		if errors.As(err, &owned) {
			log.Printf("You already own bucket %s.\n", name)
			err = owned
		} else if errors.As(err, &exists) {
			log.Printf("Bucket %s already exists.\n", name)
			err = exists
		}
	} else {
		err = s3.NewBucketExistsWaiter(a.S3.Client).Wait(
			ctx, &s3.HeadBucketInput{Bucket: aws.String(name)}, time.Duration(TimeoutMS)*time.Millisecond)
		if err != nil {
			log.Printf("Failed attempt to wait for bucket %s to exist.\n", name)
		}
	}
	return err
}

func (a *AppCtx) CreateBucketIfNotExists(ctx context.Context, bucketName string) (bool, error) {
	_, err := a.S3.Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	exists := true
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			switch apiError.(type) {
			case *types.NotFound:
				log.Printf("Bucket %v is available.\n", bucketName)
				exists = false
				err = nil
			default:
				log.Printf("Either you don't have access to bucket %v or another error occurred. "+
					"Here's what happened: %v\n", bucketName, err)
			}
		}
	}
	if !exists {
		if a.S3.Config.Region == "" {
			a.S3.Config.Region = "auto"
		}
		err = a.CreateBucket(ctx, viper.GetString(S3_BUCKET_NAME), a.S3.Config.Region)
		if err != nil {
			log.Fatal(err)
		} else {
			log.Println("Bucket created.")
		}
	}

	return exists, err
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
