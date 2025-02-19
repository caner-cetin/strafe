package cli

import (
	"context"
	"strafe/internal"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type ResourceType int

const (
	ResourceDatabase ResourceType = iota
	ResourceDocker
	ResourceS3
)

type ResourceConfig struct {
	Resources []ResourceType
	Timeout   *time.Duration
}

func WrapCommandWithResources(fn func(cmd *cobra.Command, args []string), config ResourceConfig) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		log.Tracef("running command %s", cmd.Name())
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
				if err := internal.InitializeDB(ctx, &appCtx); err != nil {
					log.Errorf("failed to initialize database: %v", err)
					return
				}
			case ResourceDocker:
				if err := internal.InitializeDocker(&appCtx); err != nil {
					log.Errorf("failed to initialize docker: %v", err)
					return
				}
			case ResourceS3:
				if err := internal.InitializeS3(&appCtx); err != nil {
					log.Errorf("failed to initialize s3: %v", err)
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
		cmd.SetContext(context.WithValue(ctx, internal.APP_CONTEXT_KEY, appCtx))
		fn(cmd, args)
	}
}
