package internal

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"strafe/pkg/db"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/docker/docker/client"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type AppCtx struct {
	DB     *db.Queries
	StdDB  *sql.DB
	Docker *client.Client
	Conn   *pgx.Conn
	S3     struct {
		Client  *s3.Client
		Config  aws.Config
		Manager *manager.Uploader
	}
	Context context.Context
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

func (a *AppCtx) UploadObject(ctx context.Context, bucket string, key string, contents []byte) (string, error) {
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(contents),
	}

	_, err := a.S3.Manager.Upload(ctx, input)
	if err != nil {
		var noBucket *types.NoSuchBucket
		if errors.As(err, &noBucket) {
			return "", fmt.Errorf("bucket %s does not exist: %w", bucket, err)
		}
		return "", fmt.Errorf("failed to upload object: %w", err)
	}

	return key, nil
}
func (a *AppCtx) ListObjects(ctx context.Context, bucketName string) ([]types.Object, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	}

	var objects []types.Object
	paginator := s3.NewListObjectsV2Paginator(a.S3.Client, input)

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			var noBucket *types.NoSuchBucket
			if errors.As(err, &noBucket) {
				return nil, fmt.Errorf("bucket %s does not exist: %w", bucketName, err)
			}
			return nil, fmt.Errorf("failed to list objects in bucket %s: %w", bucketName, err)
		}

		if output.Contents != nil {
			objects = append(objects, output.Contents...)
		}
	}

	return objects, nil
}

func (a *AppCtx) DownloadFile(ctx context.Context, bucketName string, objectKey string) ([]byte, error) {
	result, err := a.S3.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		var noKey *types.NoSuchKey
		if errors.As(err, &noKey) {
			log.Printf("Can't get object %s from bucket %s. No such key exists.\n", objectKey, bucketName)
			err = noKey
		} else {
			log.Printf("Couldn't get object %v:%v. Here's why: %v\n", bucketName, objectKey, err)
		}
		return nil, err
	}
	defer result.Body.Close()
	body, err := io.ReadAll(result.Body)
	if err != nil {
		log.Printf("Couldn't read object body from %v. Here's why: %v\n", objectKey, err)
	}
	return body, err
}
