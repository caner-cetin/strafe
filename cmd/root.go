package cmd

import (
	"context"
	"database/sql"
	"os"
	"time"

	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "strafe",
	Short: "Upload utility for dj.cansu.dev",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var (
	TimeoutMS int
)

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.strafe.yaml)")
	rootCmd.PersistentFlags().CountVarP(&verbosity, "verbose", "v", "verbose output (-v: info, -vv: debug, -vvv: trace)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().IntVarP(&TimeoutMS, "timeout", "T", int((time.Minute * 20).Milliseconds()), "default timeout for commands in milliseconds, set to 20 minutes by default")
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(getConfigCmd())
	rootCmd.AddCommand(getDockerRootCmd())
	rootCmd.AddCommand(getAudioRootCmd())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
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
		switch verbosity {
		case 1:
			log.SetLevel(log.InfoLevel)
		case 2:
			log.SetLevel(log.DebugLevel)
		case 3:
			log.SetLevel(log.TraceLevel)
		default:
			log.SetLevel(log.WarnLevel)
		}
		log.Debugf("using config file: %s \n", viper.ConfigFileUsed())
		viper.SetDefault(DOCKER_IMAGE_NAME, DOCKER_IMAGE_NAME_DEFAULT)
		viper.SetDefault(DOCKER_IMAGE_TAG, DOCKER_IMAGE_TAG_DEFAULT)
	} else {
		log.Errorf("cannot load config file: %v \n", err)
		os.Exit(1)
	}

}

type AppCtx struct {
	DB     *sql.DB
	Docker *client.Client
}

func wrapCommandWithContext(fn func(cmd *cobra.Command, args []string)) func(cmd *cobra.Command, args []string) {
	appCtx := AppCtx{
		DB:     openDB(),
		Docker: newDockerClient(),
	}
	defer appCtx.Docker.Close()
	defer appCtx.DB.Close()
	
	return func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(TimeoutMS))
		defer cancel()
		cmd.SetContext(context.WithValue(ctx, APP_CONTEXT_KEY, appCtx))
		fn(cmd, args)
	}
}
