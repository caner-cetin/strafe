package cli

import (
	"context"
	"time"

	"github.com/caner-cetin/strafe/internal"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// ResourceType defines the type of resource to be initialized
type ResourceType int

// Resource types for initialization
const (
	// ResourceDatabase represents database resource type
	ResourceDatabase ResourceType = iota
	// ResourceDocker represents docker resource type
	ResourceDocker
	// ResourceS3 represents S3 resource type
	ResourceS3
)

// ResourceConfig holds the configuration for resource initialization
type ResourceConfig struct {
	Resources []ResourceType
	Timeout   *time.Duration
}

// WrapCommandWithResources wraps a Cobra command function with resource initialization and cleanup logic based on the provided configuration
func WrapCommandWithResources(fn func(cmd *cobra.Command, args []string), config ResourceConfig) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		var to time.Duration
		if internal.TimeoutMS == 0 {
			to = time.Millisecond * time.Duration(*config.Timeout)
		} else {
			to = time.Millisecond * time.Duration(internal.TimeoutMS)
		}
		ctx, cancel := context.WithTimeout(context.Background(), to)
		defer cancel()

		appCtx := internal.AppCtx{}

		for _, resource := range config.Resources {
			switch resource {
			case ResourceDatabase:
				if err := appCtx.InitializeDB(); err != nil {
					log.Error().Err(err).Msg("failed to initialize database")
					return
				}
			case ResourceDocker:
				if err := appCtx.InitializeDocker(); err != nil {
					log.Error().Err(err).Msg("failed to initialize docker")
					return
				}
			case ResourceS3:
				if err := appCtx.InitializeS3(); err != nil {
					log.Error().Err(err).Msg("failed to initialize s3")
					return
				}
			}

		}
		defer func() {
			if appCtx.Conn != nil {
				if err := appCtx.Conn.Close(ctx); err != nil {
					log.Error().Err(err).Msg("failed to close database connection")
					return
				}
			}
			if appCtx.Docker != nil {
				if err := appCtx.Docker.Close(); err != nil {
					log.Error().Err(err).Msg("failed to close docker client")
				}
			}
		}()
		cmd.SetContext(context.WithValue(ctx, internal.APP_CONTEXT_KEY, appCtx))
		fn(cmd, args)
	}
}
