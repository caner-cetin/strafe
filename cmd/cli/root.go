package cli

import (
	"os"
	"strafe/internal"
	"strafe/pkg/server"
	"time"

	"github.com/spf13/cobra"
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

var ()

func init() {
	cobra.OnInitialize(internal.InitConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.strafe.yaml)")
	rootCmd.PersistentFlags().CountVarP(&internal.Verbosity, "verbose", "v", "verbose output (-v: info, -vv: debug, -vvv: trace)")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().IntVarP(&internal.TimeoutMS, "timeout", "T", int((time.Minute * 20).Milliseconds()), "default timeout for commands in milliseconds, set to 20 minutes by default")
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(getConfigCmd())
	rootCmd.AddCommand(getDockerRootCmd())
	rootCmd.AddCommand(getAudioRootCmd())
	rootCmd.AddCommand(server.GetRunCmd())
}
