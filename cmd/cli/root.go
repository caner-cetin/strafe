package cli

import (
	"fmt"
	"math/rand/v2"
	"os"
	"time"

	"github.com/caner-cetin/strafe/internal"
	"github.com/caner-cetin/strafe/pkg/server"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "strafe",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	internal.InitConfig()
	cobra.OnInitialize(internal.InitConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.SetHelpFunc(modifyHelp(rootCmd.HelpFunc()))
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
	rootCmd.AddCommand(getDBRootCmd())
}

func modifyHelp(fn func(cmd *cobra.Command, args []string)) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		if viper.GetBool(internal.DISPLAY_ASCII_ART_ON_HELP) {
			fmt.Println(arts[rand.IntN(len(arts))])
		}
		fn(cmd, args)
	}
}

var arts = []string{
	`
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠿⠛⢛⣛⣛⣛⠛⠻⠿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⢛⣡⣴⣾⣿⣿⣿⣿⣿⣿⣿⣶⣤⡙⠛⣉⣙⠛⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⡿⠋⣡⡴⢂⣴⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣼⣿⣿⣦⡉⠻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⠏⣠⣾⣿⣧⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡏⢿⣿⣿⣶⣬⣙⠛⠛⠛⠛⠛⠛⠻⠿⣿⣿⣿
⣿⣿⣿⣿⠟⣡⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠋⠈⢻⣿⣿⡄⢻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣄⠹⣿
⣿⣿⠟⣡⣾⣿⣿⠿⠟⠰⢿⣿⡟⠉⠻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣆⣀⢼⣿⣿⣻⠘⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣇⢹
⡿⢡⣾⣿⠟⣉⣤⣶⣶⣶⣦⡙⢇⠀⢀⣿⣿⣿⢻⡿⠿⠿⢉⣿⣿⣿⡿⢻⠿⣿⠝⣠⡙⠿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠇⣼
⢡⣿⣿⣇⣾⣿⣿⣿⣿⣿⣿⣿⡌⡿⠟⣿⣿⣿⣶⣶⣿⣿⠿⣿⣿⣿⡿⠟⢋⣡⣾⣿⣿⣷⣬⣙⠻⠿⣿⣿⣿⣿⠿⠛⣡⣾⣿
⠘⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠃⣿⣾⣿⡿⠿⠿⢋⣴⡄⢶⣦⣁⢠⡀⠺⣿⣿⣿⣿⣿⣿⣿⣿⣿⣶⣶⣦⣤⣶⣶⣿⣿⣿⣿
⣧⠙⢿⣿⣿⣿⣿⣿⣿⠿⠛⣡⡶⠆⠐⠲⠀⣶⣠⣿⣿⠇⣌⠻⢿⣿⣿⣦⡘⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣷⣦⣭⣭⣭⣭⣥⣴⣾⣿⢋⣴⡿⢿⡟⢸⣿⣿⣿⣿⣼⣿⣿⣿⣿⣿⡿⢁⣴⡄⢻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡇⢸⣿⡇⢿⣇⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⣿⣿⡟⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣧⠸⣿⣿⣶⣴⣄⠻⣿⣿⣿⣿⣿⠿⠟⠋⣡⣴⣌⣉⣡⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣌⣙⡛⣛⣉⣴⣦⣌⠹⠿⠷⢀⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣶⣶⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
			`,
	`
⠀⠀⠀⠀⠀⠀⠀⠀⠉⠛⠻⢿⣿⡀⢀⠾⣿⡹⢧⣛⢧⣟⡏⣼⢻⣽⣻⣿⣿⣿⣿⢿⢹⣳⡻⣿⣿⡟⢸⣽⢳⣻⣿⣿⣿⣿⣿⣿⣞⠩⣿⡼⣇⣿⣇⠠⢻⣧⢂⢣⢻⣿⣿⣿⣷⣙⣿⣿⣿⣇
⠀⠀⠀⠀⠀⠀⢀⣾⣷⡄⢢⠠⠈⢀⡾⣛⡿⣝⡳⣏⢿⣽⢠⣟⢻⣷⣿⣿⠿⣛⣯⠏⣳⢷⡻⣽⢞⠅⣻⢸⣏⣷⢻⡼⣏⡿⣏⡿⣽⡂⣽⡃⣯⢟⣻⠀⢃⢿⣎⡆⢏⣿⣿⣿⣿⣿⠹⣿⣿⣿
⠀⠀⠀⠀⠀⠀⣾⡻⣟⡿⠆⠡⢀⡾⡵⣏⡿⣱⢯⡽⣾⡌⣴⣋⣾⣿⠟⣡⢟⡾⠅⢸⣛⡾⣝⣧⡟⢀⡏⢰⣻⡼⢯⡷⣏⡷⣏⡇⣳⠇⣳⡅⢯⣛⡷⠂⠘⡖⢻⡶⡘⣞⣶⢲⣭⣛⠇⢿⣿⣿
⠀⠀⠀⠀⠀⠀⣖⢯⣳⡀⢐⠀⣸⢳⣝⡞⣷⢫⣷⣽⣼⡁⡷⣭⣿⢋⡾⢁⣮⠇⡄⡿⣭⡟⣽⢶⡃⢸⡇⢸⣳⣻⢻⡼⣏⡷⣏⣇⢽⡃⡷⡆⣹⣏⡇⡀⣇⢹⡈⡷⣇⠻⣼⣛⡶⢯⣗⠘⣿⢿
⠀⠀⠀⠀⠀⠀⢠⠀⠄⡁⢈⠀⣯⢻⡼⣹⣷⣯⣾⣿⣿⠁⣇⣿⡿⢸⠃⣸⡞⢰⠸⣝⣧⢿⣹⡞⢱⢸⡅⠸⣧⡟⣯⣗⢯⡽⣝⡞⢸⣁⡷⡇⢼⢧⡃⡇⣿⡄⣷⠸⣳⠈⡷⢯⣽⣛⡎⡀⣿⣫
⠀⠀⠀⠀⠀⠀⠐⢥⡘⢄⠂⢰⣏⢷⡻⣷⣿⢿⣿⣿⣿⠀⡇⡿⡽⡇⢠⢷⡃⡌⣾⡽⢾⣹⢮⡅⡎⢸⠃⡀⣷⣻⢷⣞⡯⣟⡾⡅⢻⢠⢿⡅⡾⢋⢀⡇⡻⣷⢸⡅⢹⡇⢹⣛⡶⣏⡯⠇⣷⣫
⠀⠀⠀⠀⠀⠀⠀⠈⣾⡀⠆⢸⣞⣳⣿⣿⣿⣿⣿⣿⡿⢀⠇⣿⡽⠀⡼⣯⢡⢃⡷⠻⠯⠛⢋⢐⡃⣋⢀⡃⣁⣉⣉⣉⠛⣽⣳⠆⡏⢸⢯⢰⡇⡞⢸⢂⢿⣮⡌⣯⠈⣿⡀⢸⣳⢯⣽⢂⡷⢯
⠀⠀⠀⠀⠀⠀⠀⢰⣟⠻⠀⣻⣿⣿⣿⣿⣿⢺⣿⣟⡟⢠⠘⣷⠏⠀⣿⠃⣋⢨⡶⣶⢯⡟⡈⣼⡇⡟⣼⡇⠀⣿⡿⣽⡇⣿⡽⢰⠃⣾⠇⣾⢰⢣⢸⢘⣓⠋⠃⠿⡀⡸⡇⠆⣿⣿⢾⣸⢼⣻
⠀⠀⠀⠀⠀⠀⠀⠸⢈⣵⡇⣿⣿⣿⡟⠙⠻⣸⢷⡾⣹⠀⠘⣯⠂⢸⣯⠁⡏⢸⡽⣞⡷⢰⢡⣿⢃⠇⣿⣿⠀⣿⣿⣽⠰⣯⡇⡞⢠⡿⢰⠃⡎⣾⠘⣾⣿⣿⣷⢠⠄⣅⠛⠀⢹⣯⣿⢸⣻⢷
⠀⠀⠀⠀⠀⠀⠀⠀⠻⣽⠀⣿⣿⡟⣼⡀⠜⣯⣟⣳⣟⠀⠀⣿⡀⣼⡳⠀⣇⢸⡿⡽⢃⢃⣿⡿⡘⢐⣻⢿⠀⣿⣳⡏⢸⣷⠃⢀⣾⢡⠇⡸⣸⡏⢠⣿⣿⣿⣷⢸⢢⣿⡸⢠⢘⠻⣽⢸⣿⡏
⠀⠀⠀⠀⠀⢀⡒⡀⠀⠉⠄⢹⣿⢰⣟⣇⢂⠰⣯⢷⢯⡀⠄⢟⠀⣷⡇⢲⠹⠘⠁⠁⠀⠀⠉⠁⠁⠹⣻⣿⠀⢹⣿⠇⣾⠇⢀⡾⢡⢪⢆⢳⣿⠃⡾⣛⣛⡿⠿⢸⢸⣿⣇⠃⢸⣿⣮⠈⣷⠇
⠀⠀⠀⠀⠀⢆⡱⠨⢄⠀⠈⠸⣧⠸⣻⢾⡄⠂⢹⣞⣟⡆⠀⢬⠀⠋⠀⠀⠀⠀⢀⠠⠀⡔⠠⠰⣶⣦⣤⣉⠀⡀⣿⢡⡿⢠⡞⠁⣵⠏⣠⣿⠃⣼⠯⠛⠋⠙⠛⠘⠸⠟⣿⠀⢸⣿⢾⢰⣿⢀
⠀⠀⠀⢠⠘⠤⢒⡉⢦⠀⡄⠀⢻⡄⢫⣟⢷⡘⡀⢹⡞⣷⡀⠸⠈⠀⢀⣤⠐⡂⢌⡘⢡⣿⣿⠂⢿⣿⣿⣿⣤⣧⣄⡜⠡⠋⣠⣾⣟⣰⣿⣯⣾⡇⠀⡀⠠⣀⠀⢀⠀⠈⠛⠇⢽⣿⡇⣼⡏⣸
⠀⠀⢀⠆⠘⢬⡁⡈⢠⡇⠐⡆⠈⢳⡄⠻⣎⠳⣌⠀⠙⣍⣁⠀⠀⠀⣾⣷⠠⣵⡀⢘⡀⣼⢣⠇⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠰⠠⢮⡿⢄⢋⡈⣷⣄⠀⠀⠠⣴⢠⡿⢱⣻
⠀⠀⣌⠂⣍⠢⢀⡇⢸⠇⣦⠈⡅⠀⠘⢄⠈⠳⡝⢦⡂⠈⠛⣧⡘⣦⡙⢿⡄⢳⣿⡟⡾⣭⡐⢿⣿⣿⣿⣿⣿⢿⣿⡿⣿⡿⣿⣿⣿⣿⣿⣿⣿⢰⠈⠆⢰⣊⢦⠅⣿⣿⣇⠀⠀⠁⡾⢁⣿⠍
⠀⡔⢢⠁⡆⠁⡼⡆⢸⡃⣽⠀⠆⣴⡶⢀⠁⠀⠈⠳⣍⠂⡀⠀⠑⠈⠿⣷⣿⣦⣨⣭⣥⢴⣶⣿⣿⣿⣿⠅⣽⣿⣿⡿⣿⣟⣯⠿⢿⣽⣿⣿⣿⡈⡷⣖⣿⢞⣍⣰⣿⣿⠇⡠⠂⠔⢠⡾⠋⡰
⠰⡈⠦⠑⣠⢯⣝⡃⢸⡇⢀⡘⠠⣿⢸⢻⠛⡖⡀⠤⡈⠛⠒⠤⡀⠑⣶⣌⣙⠛⣾⣾⢞⡷⡻⣞⣺⢹⣽⣿⣿⢷⣿⡷⢟⣹⣗⡿⣿⣿⣿⣻⣯⣷⣦⢍⡈⢃⢹⣿⣻⢯⣾⠀⠀⠴⠋⠁⡴⠁
⠡⡅⠃⣰⢯⣳⢾⠀⢸⡇⢸⣃⠀⣿⡈⣿⡆⢹⠀⢶⣥⣉⠒⠄⠀⢠⣝⡻⡛⡸⢿⢺⣙⣟⢋⣿⣿⣯⡯⣿⣹⣼⣻⣩⣶⣽⣺⣵⣾⡭⣻⣯⠿⣿⣽⣶⠿⣾⢻⣾⣷⣯⡇⠀⡔⠀⢌⠊⢀⡀
⠱⢀⡃⣽⢺⡵⣫⠀⠊⡅⢰⣻⠀⠘⢷⣬⡳⣈⡁⢸⡾⣝⡿⡆⣮⠈⣷⢟⣰⣛⣗⠾⡼⣪⣴⢔⡨⣶⣷⢽⣾⣿⡽⣯⡾⣜⣿⣾⣿⣧⣭⣚⣋⡻⡿⣯⣾⣿⣯⣿⣿⣷⢀⠓⢀⡜⢬⠇⠸⡄
⠁⠼⢐⡃⠿⣜⠷⠀⢂⠅⣤⡟⣆⢄⠀⠙⢿⣮⣕⠸⣽⣫⢷⢁⣵⠀⣱⣊⣼⢬⣿⢿⣶⣻⣹⣾⣶⣿⣷⣞⣽⣿⣿⢿⡷⠿⣵⠞⣽⣿⢯⣤⣵⡾⣏⣩⣼⣿⣾⡿⢿⡏⣠⢡⣞⡟⡏⠀⣇⠇
⠀⠀⠌⡐⠀⠀⠀⠈⠂⠆⢸⣳⠈⡌⡴⣄⡀⠙⠿⢠⢿⣵⡏⢸⣽⠀⢸⣟⣾⣿⣾⣿⢯⣿⣿⣿⣿⣿⣿⣿⣿⣿⢟⢫⣾⣿⣶⣝⡿⣿⣿⣿⣿⣷⣷⣾⡿⣟⣵⣿⣷⡅⢡⣟⡾⡽⠐⢰⣊⢧
⠀⠀⢣⠀⠀⠀⠀⠀⡁⠄⠘⢽⡆⠰⠀⢻⣳⡄⠀⠘⣿⣞⠇⣟⣾⠁⢸⣿⣿⣿⣿⣿⣷⣿⣿⣿⣿⣿⡿⢟⣯⣾⣿⣿⣿⣿⣿⣿⣿⣶⣔⠙⣿⣿⡿⣫⣾⣿⣿⣿⢿⣿⣦⡙⡾⢡⠃⢦⡙⣮
⠀⠀⢂⠆⠀⠀⠀⠀⢐⠈⠐⡀⠻⣄⠣⠈⢷⣻⡄⠈⣷⡟⢨⣟⣾⠀⡇⠻⣿⣿⣿⣿⣿⣿⡿⢫⣵⣶⣿⣿⣿⣿⣿⣿⢏⣽⣛⠿⣿⣿⡿⠼⣿⠏⣾⣿⣿⢋⡋⠘⣿⢿⣿⣿⣦⡙⢰⢢⡝⣳
⢀⠀⠀⠂⠀⡀⠀⠀⢈⠜⡀⠐⠄⡈⠂⡡⠀⢻⣽⠀⣿⠇⣼⣟⡾⠈⣇⢸⣮⡛⢿⣿⠟⣡⣾⣿⣿⣿⣿⣿⣿⠿⣫⣿⣿⣿⣿⣷⡘⢿⡀⠀⠁⢀⣾⢟⡵⠋⠀⡇⢈⠻⠯⣿⡿⣻⣦⣬⣘⠳
⠀⠈⠀⠀⠀⠀⠀⠀⢀⠊⡔⠈⠰⠠⠁⠄⠁⠀⠹⡀⣿⢀⣿⣿⣽⠃⣯⠀⢿⡿⢃⣴⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣭⡻⢿⣿⣿⣿⣿⣮⣑⣀⣤⣭⡷⠋⠀⠀⢰⢁⡾⣷⠀⣩⣾⣿⣿⣿⣿⣿
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠌⠀⠀⠀⠐⠠⠈⡐⠀⠀⠀⡏⣼⣿⣿⣯⡇⣿⠄⣠⣴⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣍⣻⢿⣿⣿⣿⡿⠟⠁⠀⠀⠀⠀⠼⣸⡽⡞⣰⣿⣿⣿⣿⡿⣿⣿
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠡⡁⢄⡈⠐⠀⠃⣿⣿⣿⡿⢃⣡⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣍⠛⠿⢿⣿⣿⣿⣿⣷⣝⠛⠁⠀⠀⠀⠀⠀⠀⠀⡇⣟⢎⣼⣿⣿⣿⠟⠉⣴⣿⣿
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⡁⠂⠐⠡⠂⢠⣿⠟⣫⣴⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣦⣀⠈⠛⢿⣿⣿⣿⣿⣦⡀⠀⠀⠀⠀⠈⠀⡿⢣⣾⣿⣿⠟⢡⢎⣾⣿⣿⣿
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⡐⡀⠀⠀⠀⠘⣡⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣯⣟⡿⣿⣿⣿⣿⣿⣿⣦⡀⠈⠻⢿⣿⣿⣿⣆⡀⠀⠀⠀⠀⢃⣿⣿⡿⠁⠀⢠⣿⣿⣿⣿⣿
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢄⠡⡁⢆⢀⣿⣿⣿⡿⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣶⣽⡻⢿⣿⣿⠟⠀⠀⠀⠀⠙⠻⣿⣿⣿⣄⠀⠀⢠⡿⣿⠞⣽⡅⠀⢜⡿⡿⣻⣿⣿
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠐⠢⠁⠀⠘⡉⢉⠀⡀⠀⠙⠛⠿⠿⠿⠛⠿⠋⠉⠛⠛⠻⢿⣿⣿⡿⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠻⣿⡟⠁⠀⠁⢈⠏⣶⢹⠆⠀⠀⠉⠀⠹⢿⡿
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢡⠐⡈⠰⢀⠂⠄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠉⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠙⠂⠀⢀⣪⣶⡹⡎⡇⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠠⢂⠁⠆⢡⠈⢂⠡⢂⠐⡀⠆⡐⠀⠀⠀⠀⠐⠀⢀⠀⠀⠀⠀⠀⠄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢿⣿⣿⡕⡀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠐⢂⠉⡐⢂⠡⠌⡐⢂⠡⠐⠂⡔⠀⢠⠀⠄⡀⢂⠀⠀⠀⠀⠄⠂⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⢷⣿⣿⣆⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠐⡈⠤⢁⠂⠆⢂⡁⠆⣈⠡⠡⣀⠃⠀⢧⠐⠁⠀⡀⠆⡐⠠⢀⠀⡀⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠻⣿⣿⣷⣄⠀⠀⠀⠀⠀
`,
	`
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠿⠛⠛⠋⣉⣉⣉⣉⣙⠛⠛⠿⠿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠟⠉⣁⣠⠀⠀⣁⣤⣶⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣶⣦⣄⡉⠛⠟⠛⠛⠻⠿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟⢁⣠⣶⣿⣿⣇⣴⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣤⡈⠿⣷⣦⣄⠙⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟⢁⣴⣿⣿⣿⡿⠻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣄⠘⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡟⠁⣰⣿⣿⣿⣿⡟⢀⣼⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡟⢿⣿⣿⣿⣷⣄⠙⠻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⡿⠋⢠⣾⣿⣿⣿⣿⡿⢀⣾⣿⣿⠏⠉⠙⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟⠻⣿⣿⣿⣿⣿⡀⢿⣿⣿⣿⣿⣿⣦⣈⠛⠿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⠿⠋⣠⣴⣿⣿⣿⣿⣿⣿⠃⣸⣿⣿⣯⠀⠀⢀⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠋⠀⠀⢸⣿⣿⣿⣿⡇⠸⣿⣿⣿⣿⣿⣿⣿⣷⣦⣌⠙⠻⢿⣿⣿⣿⣿⣿
⣿⣿⣿⠿⠋⣁⣴⣾⣿⣿⣿⣿⣿⣿⣿⣿⡀⢸⣶⣾⣶⣷⣶⣿⣿⣿⣿⣿⣿⠟⡛⠿⣿⣿⣿⣿⣿⣶⣀⣀⣾⢿⣿⣿⣿⡇⠀⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣶⣄⡉⠻⣿⣿⣿
⣿⠟⢁⣴⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣧⠈⠻⣛⣟⣯⡿⢿⣿⣿⣿⣿⣿⣴⣾⣶⣾⣿⣿⣿⣿⣿⣻⣿⣿⣿⡾⣿⣿⠃⡀⠙⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣦⠈⢻⣿
⡟⢀⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡟⢠⣄⡉⠛⠋⢁⣈⠙⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠋⠙⠛⣿⣣⠿⠋⣠⣷⡀⠹⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣦⠈⣿
⠀⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟⢀⣾⣿⣿⠀⣼⣿⣿⡇⢠⣌⣉⣉⣉⡙⠉⠋⠉⠋⠉⠉⠀⣶⣿⣆⠈⠀⠀⠚⠻⢿⣿⡄⠙⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡇⢸
⠀⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠋⣠⣾⣿⣿⣿⡀⠹⠟⢻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣦⣌⣿⡿⠀⣠⣶⣶⣶⣤⠈⢻⣦⡈⠻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠃⣸
⣧⡈⠻⢿⣿⣿⣿⣿⣿⣿⡿⠟⠋⣠⣼⣿⣿⣿⣿⣿⣿⠃⣰⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣏⣉⣀⠈⢿⣿⣿⠿⣿⣷⠀⢻⣿⣦⡈⠻⢿⣿⣿⣿⣿⣿⣿⠿⠃⣰⣿
⣿⣿⣶⣤⣈⣉⣉⣉⣉⣁⣤⣴⣾⣿⣿⣿⣿⣿⣿⣿⠇⢠⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡆⠘⠛⠋⣰⣿⣿⠂⢸⣿⣿⣿⣷⣤⣀⣉⣉⣉⣉⣀⣴⣾⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠀⢸⣿⠟⠛⠻⣿⣿⣿⣿⣿⣿⣿⠿⠿⢿⣿⣿⣿⣿⣿⠃⢸⣿⣿⣿⣿⠋⢠⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣆⠈⢁⣴⣿⣾⣿⣿⣿⣿⣿⡿⠁⣤⣶⣦⣬⣿⣿⡿⠋⣠⣤⣀⣁⣀⣤⣶⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⠈⣿⣿⣯⡉⠉⠉⠉⠉⠁⠀⢿⣿⣿⡉⢉⣤⣴⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣆⠘⠿⠟⢀⣾⣿⣿⣿⣿⡆⠸⢿⠟⢁⣼⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣶⣾⣿⣿⣿⣿⣿⣿⣿⣦⣤⣶⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
`,
	`
⠀⠁⠀⠀⢀⠀⠀⢀⢠⣀⠀⠈⠀⠀⠰⢎⣶⣿⣿⣯⠟⣯⠗⢃⣾⣿⣿⠯⠋⢆⡌⣑⣁⢆⡊⢘⡩⠳⠮⣏⠟⣶⣄⡀⠀⠀⠀⡀⠀⠀⢀⠀⠀⠀⡀⠀⠀⡀⢀⠀⠀⡀⠄⠀⠀
⣶⡂⠀⠠⢁⠤⠢⠜⠧⠼⢧⣷⠠⠀⠀⠐⠃⠗⠛⢡⠞⣡⣲⣟⣽⣟⡃⠒⢣⠎⡘⢜⢿⣦⡹⣿⣗⠦⣭⣮⢕⠻⡝⢿⣦⣄⠁⠀⠄⠂⢀⠐⠈⠀⠄⠂⠁⠀⠄⢀⠂⠠⠐⠈⠀
⠀⠈⠁⠂⠁⠀⢀⠀⢀⠰⢀⠀⠀⠀⢶⠀⠀⠀⠘⣦⡙⡴⣹⢿⡻⢷⠉⣰⡟⠰⢸⡌⣎⢻⣷⡌⢿⣿⣾⡻⣿⣷⣯⣉⢻⣿⢷⣄⠀⠠⠀⡀⠂⠈⢀⠐⠈⠀⠌⠀⢀⠂⠐⠈⠀
⢀⠂⡀⠀⠈⠐⢦⣄⠀⠨⠀⠀⠀⠀⠈⠁⠀⠀⠀⠖⡡⢼⣿⣿⠻⡆⢢⣿⠇⡆⣿⣇⡘⡆⢿⣿⣱⡽⣿⣿⣮⡯⣻⣿⣦⡙⣿⣿⣷⣄⠀⠄⠀⢈⠀⠄⠀⠡⢀⠈⢀⠠⠁⠐⠀
⠈⠀⢿⠀⠐⠠⠀⢉⠻⠖⠀⠤⢄⠀⠀⠀⠀⠀⠀⣘⢹⣾⣿⣽⣿⢠⣿⣿⣀⣧⣿⣿⣿⣿⢸⣿⣷⣿⡽⣮⣻⣿⣧⡻⣿⣿⣮⣿⣿⣿⣷⣄⠔⠀⠀⠀⠌⠀⠠⠀⡀⠄⠂⠈⠀
⠠⠀⡘⠀⠀⢀⠃⠄⠀⠘⠘⠣⣼⡄⠀⠀⠀⠀⠸⡃⣿⣿⣼⣿⡿⣼⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣼⣿⣼⣧⣿⣿⣿⡜⣿⣿⣧⢿⣿⣿⣧⣣⡀⠀⠀⠀⡘⠀⡀⠀⠠⠀⠃⠀
⠀⢂⠐⢀⠀⠀⠀⡀⠄⠀⠈⠈⠉⠙⠀⠀⢀⠀⠴⢡⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡇⣿⣿⣿⣿⣿⣿⣿⣾⢿⣿⣮⢿⣿⣧⠹⣿⣿⣝⢿⣦⣄⠀⢀⠐⠀⢀⠁⠄⠂⠀
⠀⠂⢐⡈⣈⠐⢄⡀⡄⠀⠀⠀⠀⠀⠀⠀⠀⢬⡏⣸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣧⢻⣿⣟⣿⣿⣽⣿⣿⣿⣻⣿⣿⣿⣿⣿⣿⣿⣧⠜⣿⣿⣷⣾⣝⠛⠦⠀⠀⠠⠀⠂⠀⠄
⠀⠀⠐⣠⠂⠉⠀⠉⠉⠁⠈⠀⠀⠀⠀⠐⣎⣿⡇⣿⣿⣿⣟⡿⣼⣿⣿⣿⡟⣿⣿⣿⡏⣿⣋⡿⣿⢸⣿⣿⣞⣧⢟⣿⣧⠻⣿⣿⢿⣿⣯⣌⢻⣿⣷⡻⣷⠄⢀⠀⠀⡐⠀⠁⠀
⠀⠁⠀⠉⠪⢆⢳⠈⣀⠁⣀⠀⠠⢖⡀⣾⣧⣿⣁⣿⣿⣿⣿⢃⣿⣿⣿⣿⡇⣿⣿⣿⡇⢻⠄⣇⣇⣻⣿⣿⣿⢿⡄⣻⣿⣷⡹⣿⣯⢻⣿⡽⣦⠙⢿⣿⣽⣿⣶⣤⡀⠠⠐⠈⠀
⠀⠀⠀⠀⠀⠈⠁⠁⠀⠒⠉⠉⠉⠑⣸⣿⣷⣿⣿⣿⣿⣿⣿⢸⣿⣿⣿⣿⠃⣿⣿⣿⢃⡾⣃⢿⣎⣿⣿⣿⣿⣸⡇⣷⢞⣿⣷⣻⣿⣖⣷⢳⡹⣷⡈⠻⣿⡜⠻⠿⣿⣶⣤⣀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢰⢿⣿⣿⣿⣿⣿⣿⣿⡟⣿⣿⣿⣿⣿⡀⣿⣿⡾⣼⣷⣹⡿⣼⣿⣿⣿⣿⡇⡿⣋⣥⣺⢾⣵⣻⣿⡼⡟⣧⢜⢿⣆⠙⣷⡀⠀⠀⠀⠉⠁⠀
⢸⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢠⢠⢯⣿⣿⣿⣿⣹⣿⡟⣿⡇⣿⠯⣿⣿⠿⠇⣿⠿⢧⣿⢿⡽⢱⣿⣿⣿⣿⡿⣭⡶⣿⣯⢷⣻⣷⢣⢿⣷⣻⡸⣾⡄⠙⠷⣌⠳⡀⠀⠠⠀⠀⠄
⠀⠀⠀⠀⠀⠀⠀⡄⡀⠀⠀⢠⠏⣎⣿⣿⡿⡭⣿⣿⡟⣿⠡⣿⢛⣭⣶⡶⣶⣿⣿⣿⣿⢷⡍⡟⣿⣿⣿⣿⣳⡟⣧⣿⣟⣯⠥⢿⣼⡾⣧⢏⡇⣱⡽⡄⠀⠈⠑⠀⠀⠈⠀⠐⠒
⠀⠀⠀⠀⠀⢀⠾⣠⠱⢤⣀⣏⡸⣼⣿⡿⣵⠗⣿⡿⣧⣿⡰⣟⢸⣧⣿⢳⣹⣿⣳⣿⣏⣣⣷⣇⣿⣿⣿⢯⢿⡏⣸⠟⢁⣠⡈⠹⠇⣿⣿⡌⣿⡠⣿⣧⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⡰⢌⡓⣠⢃⢆⠊⣄⣣⣿⣿⠉⣼⢱⣿⢧⡏⠹⣇⡟⢾⣿⠋⡟⠭⠱⠿⠏⣴⡟⣾⡿⣿⣿⡯⣿⣿⡾⠁⠀⠈⢿⣿⡄⣦⣿⡽⣼⣸⡇⢿⣿⣧⡄⠀⠀⠀⠀⠀⠀⠀
⣀⡤⡴⣠⢃⡈⢇⣹⠌⢈⡆⠙⣿⡿⡉⢆⢏⢸⡿⣼⢃⠀⣽⠇⡚⠉⠀⠁⠀⠀⢰⣤⢭⢥⣿⡇⣿⢟⣵⣿⣿⡇⣀⡀⢦⢸⣿⡇⢹⣿⡑⢯⣿⣷⢘⢿⣯⢵⡄⠀⠀⠀⠀⠀⠀
⢢⠹⡁⢦⠧⢌⣎⡟⠠⣵⢈⢿⣿⠍⠌⢈⡌⣿⣷⢫⣇⢠⠶⠃⣀⠀⠀⠀⠀⢄⠀⣿⣿⣿⣿⢇⣿⣿⣿⣿⣿⣏⢫⣤⣾⢸⣿⣇⢸⣿⢀⢼⣿⣯⣧⠘⣿⣟⢾⣆⠀⠀⠂⠀⠀
⢡⠢⣍⡏⣜⣞⣾⣵⢻⡟⡾⢸⢏⣘⠂⢸⢻⣟⢂⡹⣾⡤⠆⡀⢵⡞⢻⣈⠖⣸⣆⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣦⣟⣫⣷⣿⣿⣸⢇⢨⣔⡫⣿⢸⣧⣊⢿⣮⢿⣆⠀⠀⠀⠀
⢢⣼⡿⢠⢲⣿⣿⠏⣾⣿⣷⠘⣠⠂⡄⣏⡿⠂⣼⢠⠽⡆⠹⣿⣭⣿⣮⣿⣿⣟⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡯⢣⢆⡎⣷⡜⣬⢎⢻⣧⠘⠽⡟⢿⣆⠀⠀⠀
⣼⢳⠃⢡⣿⢿⠏⣬⣽⡿⢹⣠⠧⣙⠀⢾⡄⢰⠸⠈⡞⣥⠡⢺⣷⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣾⢫⢊⡁⠹⡇⠉⢡⡁⠽⣧⠀⠈⠳⣽⣆⠀⢠
⠯⠁⢈⣾⢿⢻⢰⡸⣟⢡⣻⡟⢂⡇⡔⣭⡃⣸⢘⣃⡚⢆⡛⠴⣽⣿⡻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡯⡜⢸⢁⠀⠑⠀⠀⠑⠀⠙⣧⠀⠀⠈⠙⠄⠀
⠁⠀⣺⡝⡎⡅⢁⣷⠃⣬⣿⣇⡿⡘⢠⢹⡟⣿⣆⣿⠈⣼⣷⠳⢋⣥⣝⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⢏⠴⡉⡟⢨⡄⠀⠀⠀⠁⠀⢀⠈⠀⠀⠀⠀⠀⠀
⠀⢲⡽⠸⠌⠀⢸⡞⢱⣹⡟⡼⠐⢀⠧⢆⢡⡿⣼⡟⣼⣿⡟⢀⣾⡿⠙⠿⣗⢮⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟⠁⣿⠒⡇⣷⢻⠐⡀⠀⠀⠠⠁⠌⠀⠀⠀⠀⠀⠀⠀
⢀⠿⡜⠃⠀⠀⣾⠁⣽⡟⡘⠀⠀⡜⡼⡌⣠⣇⠛⠃⠿⠿⡱⠿⠿⠁⠂⠐⠿⠛⠿⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟⠃⠀⠘⠹⡇⣷⣽⠘⠀⠀⠀⠠⢁⠈⡐⠀⣀⣀⠀⠀⠀⠀
⣌⠛⠀⠀⠀⠀⠃⢠⡿⠉⠀⠀⢠⢁⣿⣷⠃⣏⠀⢣⠀⠲⡔⡀⠺⣧⣙⣤⣦⣤⡀⠀⠫⢺⠿⢿⣿⣿⣿⣿⠟⠁⠀⠀⠀⠀⠃⢿⣸⣿⠀⠀⠀⢀⠂⠄⠂⠄⠁⠠⠉⠻⢷⣶⣦
⡎⠀⠀⠀⠀⠀⠀⢸⠁⠀⠀⠀⢸⢸⣿⣷⡎⣿⠀⢻⣾⣿⣿⣶⣶⣾⣿⣿⣿⣿⣿⡗⡄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠘⣿⣿⠀⠀⠀⠂⠀⠂⡌⠐⠀⢡⠈⢠⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⣿⣿⡇⢱⠀⠀⠈⠙⠿⣿⡿⠿⠿⠿⣿⢿⣿⣿⣮⡄⠀⠀⠀⡴⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣿⡯⠀⠀⠠⠈⡐⠠⠐⠀⠐⠠⠀⠂⢈⠐⡀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢻⢻⡇⠈⢃⠀⠀⠀⠀⡀⠀⡜⠿⠷⠦⠀⠶⢩⠙⡞⠀⢀⢄⣷⡄⠀⠀⠀⠀⠀⠀⠀⠀⠀⢹⡇⠀⠀⢂⠁⠠⠐⠠⠁⠀⠀⠐⠀⠠⠐⠀
⠀⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⣯⢷⠂⠀⠀⠀⠀⠀⠀⠀⢸⠜⢿⡻⠄⠐⣲⠀⣸⠀⠠⢸⣿⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠘⠀⠀⡀⠂⢈⠐⠠⠁⠀⠀⠀⠀⠐⠀⠀⠄
⠀⠠⠁⠄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⢎⢷⠀⠀⠀⠀⠀⠀⢠⣶⣾⣭⢭⣤⣤⣴⣾⡟⢈⠐⣾⣿⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⠐⠠⢀⠁⢂⠠⢁⠀⠀⡐⠀⠡⢀⠂⠄⡀
⠀⠀⠐⠈⠠⢀⠀⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣠⣯⣻⣽⣿⣾⡿⢿⣿⣿⠣⠌⢀⠿⠋⠀⠀⠀⠀⢀⠀⠀⠀⠀⠐⠠⠈⡀⠂⢈⠠⢀⠂⠀⡀⠄⠀⡁⠄⠐⣠⣶
`,
	`
⣿⣿⣿⡏⣸⣿⣿⣿⣿⣿⣿⢃⣾⢋⠃⡻⣿⣿⣿⡇⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣇⣹⡟⢻⠿⢻⢹⡜⡰⠹⣼⣿⣿⣿⣿⣿⣿⣿⡆⢿⣿⣿⣆
⣿⣿⣿⢱⣿⣿⣿⣿⣿⣿⡟⣼⢏⡎⠰⢀⣿⣿⣿⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡇⠇⠰⡄⡄⢷⢡⣧⣿⣿⣿⣿⣿⣿⣿⣿⣿⡸⣿⣿⣿
⣿⣿⡇⣼⣿⣿⣿⣿⣿⣿⢣⡏⣼⣿⣷⣾⣿⣿⡏⣼⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡟⣿⣿⣿⣿⣿⡟⣿⣿⣿⣿⣶⣾⣾⣿⣿⣿⡜⣿⣿⣿⣿⣿⣿⣿⡇⢿⣿⣿
⠻⠿⢁⣿⣿⣿⣿⣿⣿⡏⠘⣸⣿⣿⣿⣿⣿⣿⡇⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡇⣿⣿⣿⣿⣿⡇⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡇⣿⣿⣿⣿⣿⣿⣿⣿⢸⣿⣿
⣴⡆⢸⣿⣿⣿⣿⣿⣿⠃⢡⣿⣿⣿⣿⣿⣿⣿⢷⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⢸⣿⣿⣿⣿⣿⢃⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⢻⣿⣿⣿⣿⣿⣿⣿⡄⣿⣿
⣿⠇⣿⣿⣿⣿⣿⣿⣿⠀⣾⣿⣿⣿⣿⣿⣿⣿⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡇⣾⣿⣿⣿⣿⡟⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⢸⣿⣿⣿⣿⣿⣿⣿⡇⣿⣿
⣿⢸⣿⣿⣿⣿⣿⣿⡇⢸⣿⣿⣿⣿⣿⣿⣿⡏⣼⣿⣿⣿⣿⣿⣿⡿⢱⣿⣿⣿⣿⣿⣿⣿⡟⣸⣿⣿⣿⣿⣿⠇⠸⡿⠿⠿⠿⠿⠿⣿⣿⣿⣿⣿⡇⣿⣿⣿⣿⣿⣿⣿⣿⢿⣿
⣿⣼⣿⣿⣿⣿⣿⣿⡇⣾⣿⣿⣿⠏⣿⣿⡟⢠⣿⣿⣿⣿⣿⣿⠟⣰⣿⣿⣿⣿⣿⣿⣿⡿⢡⣿⣿⣿⠟⣋⡍⢰⣦⡘⣿⣿⡘⣿⣿⣷⣾⣿⣿⣿⡇⣿⣿⣿⣿⣿⣿⣿⣿⣺⣿
⡇⣿⣿⣿⣿⣿⣿⣿⡇⣿⣿⡿⠋⣠⣥⠂⠀⣶⣶⣦⣭⣙⠻⢋⣼⣿⣿⣿⣿⢡⣿⣿⡿⢡⣿⣿⣟⣱⣾⣿⢣⣿⣿⣷⡌⢿⣿⣌⢿⣿⣿⣿⣿⣿⡇⢻⣿⣿⣿⣿⣿⣿⣿⣿⣿
⠇⣿⣿⣿⣿⣿⣿⣿⠀⠿⠋⢀⣼⠟⢡⠆⣼⣿⣿⣿⣿⠏⣰⣶⣿⣿⣿⢫⢇⣾⣿⠟⢡⣿⣿⣿⣿⣿⣿⡏⣾⣿⣿⣿⣿⣆⠻⣿⣦⡙⣌⢿⣿⣿⡇⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿
⡆⣿⣿⣿⣿⣿⣿⣿⢀⡴⢁⡾⢋⣴⠃⣼⣿⣿⣿⠟⡡⢰⣿⣿⣿⣿⢇⢎⣾⡿⢃⡞⣸⣿⣿⣿⣿⣿⠟⣼⣿⣿⠿⠿⢿⣿⣷⡈⢿⣿⣮⠢⠙⢿⡇⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿
⡷⣿⣿⣿⣿⣿⣿⣿⠈⡠⠋⣰⣿⠃⣼⣿⡿⠛⠡⠾⢡⣿⣿⣿⣿⠏⢀⠾⢋⣴⡿⢡⣿⣿⣿⣿⣿⠏⣴⠟⣫⠵⠾⠛⠋⠉⠁⠀⠀⠀⠉⠀⠀⠀⠁⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣧⣿⣿⣿⣿⣿⣿⣿⠀⣠⣾⠈⠁⠚⠋⠉⠀⠈⠉⠀⠀⠊⠙⠿⠃⠀⣠⣶⣿⣿⢣⣿⣿⣿⣿⡿⢁⣾⡧⠂⠀⠀⠀⠀⠀⠀⠀⠀⢀⣀⣦⣤⡀⠀⠀⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣇⣿⣿⣿⣿⣿⣿⣿⡆⡿⠛⠀⠀⢀⣠⣴⡶⠀⠀⠀⠀⠀⠀⠀⠡⣾⣿⣿⣿⢃⣾⣿⣿⡿⠋⣰⣿⣿⣷⣀⣴⡆⠀⠄⠀⣤⣄⢀⡙⠉⣿⣿⠇⠀⢀⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣇⠠⡀⠀⠰⣿⣿⣿⡇⠀⠀⠀⣴⣦⠀⠋⢰⣿⣿⣿⢃⣾⣿⣿⠋⣠⣾⣿⣿⣿⣿⣿⣿⣧⠀⣠⣦⠻⠟⢠⣽⠃⣿⡟⠀⠀⣼⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⢛⣿⢻⣿⣿⣿⣿⣿⣿⡄⢹⡄⠀⢻⣿⣿⣇⠠⣾⣧⣙⣋⣼⡇⢸⣿⣿⢃⣾⠟⠛⣡⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣆⠻⣿⠸⠿⠸⠟⣸⣿⢡⡠⠠⢣⣿⣿⣿⣿⣿⡟⣹⣿⢿⣿
⡄⢹⡘⣿⣿⣿⣿⣿⣿⣷⠀⢻⣦⠀⠻⣿⣿⣦⠹⠇⠾⠿⢀⣃⣿⡿⢁⣈⠀⣠⣾⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣶⣶⣾⣿⣿⣦⣤⡿⠁⣠⣿⣿⣿⣿⣿⠟⣠⣿⣿⢸⣿
⣼⢸⣇⢻⣿⣿⣿⣿⣿⣿⣧⡈⢿⣿⣷⣉⣿⣷⣶⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣏⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣯⣿⣿⡟⣻⠝⢀⣴⣿⣿⣿⣿⠟⣡⣾⣿⣿⡇⣬⣉
⣿⡄⣿⡜⣿⣿⠻⣿⣿⣿⣿⣷⣌⠻⣿⣿⣿⣿⣿⣿⣿⣟⣿⣿⣿⣿⣿⣿⣿⣿⣮⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣯⡾⢋⣴⣿⣿⣿⠟⢋⣤⣾⣿⣿⣿⢸⠃⣿⣿
⣿⣷⢹⣷⡘⢿⣷⣦⣝⡛⠿⣿⣿⣷⣌⠻⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⢋⣴⡿⠿⢛⣩⣴⠆⢸⣿⣿⣿⣿⡿⠸⢸⣿⣿
⣉⡉⠘⣿⣿⣮⡻⢿⣿⣿⣷⡄⢨⠉⢛⠛⠂⠌⠙⣻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠿⠿⠿⢿⣿⣿⣿⣿⣏⣀⣬⣤⣶⣾⣿⣿⡟⢀⣿⣿⣿⣿⣿⡇⡁⣿⣿⣿
⣿⣿⣧⢹⣿⣿⣿⣦⣍⠉⠛⠻⠦⠀⠘⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟⠋⢉⣽⣿⣾⣿⣿⣿⣿⣿⣿⣷⡄⢰⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠁⣾⣿⣿⣿⣿⣿⠀⢸⣿⣿⣿
⣿⣿⣿⡌⣿⣿⣿⣿⣿⡟⣷⠀⢸⣿⡆⢹⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠀⢀⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠁⢠⣿⣿⣿⣿⣿⡏⢀⣿⣿⣿⣿
⣿⣿⣿⡇⢹⣿⣿⣿⣿⣇⢻⡄⢸⣿⣿⡄⠻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣧⡘⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⢣⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠋⣴⠇⣾⣿⣿⣿⣿⡿⠀⣼⣿⣿⣿⣿
⣿⣿⣿⡇⡄⢻⣿⣿⣿⣿⡘⡇⠸⣿⣿⣿⣦⡙⢿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣌⠻⢿⣿⣿⣿⣿⣿⣿⠟⣡⣾⣿⣿⣿⣿⣿⣿⣿⠟⢉⣴⣾⡟⣸⣿⣿⣿⣿⡿⢁⣾⣿⣿⣿⣿⣿
⣿⣿⣿⡟⣿⠈⢻⣿⣿⣿⣷⠱⠀⣿⣿⣿⣿⣿⣦⣉⠻⢿⣿⣿⣿⣿⣿⣿⣿⣿⣾⣿⣿⣿⣿⣛⣥⣾⣿⣿⣿⣿⣿⣿⠿⠋⠁⠄⠻⣿⡟⣰⣿⣿⣿⣿⡿⢁⣾⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣷⡏⣸⣆⠻⣿⣿⣿⣧⠀⠹⣿⣿⣿⣿⣿⣿⣿⣶⣬⣙⠻⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠟⢋⡀⡰⢂⡀⣶⣄⠋⣰⣿⣿⣿⣿⠟⣴⣿⣿⣿⣿⣿⣿⣿⠿
⣿⣿⣿⠿⢃⣿⣿⣧⡘⢿⣿⣿⣧⠀⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣶⣬⡙⠻⢿⣿⣿⣿⣿⣿⣿⣿⠿⠋⠁⠀⠀⠀⣠⣶⣾⢡⡿⢃⣼⣿⣿⣿⠟⣡⣾⣿⣿⣿⣿⣿⣿⡟⠁⢀
⣿⣿⡟⠀⣼⣿⣿⣿⣿⣦⡙⢿⣿⣷⡌⢻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡷⠆⠈⠙⠿⠿⠛⠉⡠⠄⠀⣠⣴⣶⣾⣿⣿⡿⠈⣠⣿⣿⡿⢋⣡⣾⣿⣿⣿⣿⣿⣿⡿⠃⠀⠐⢋
⡿⠜⠀⣰⣿⣿⣿⣿⣿⣿⣿⣦⣝⠻⢿⣦⠹⣿⣿⣿⡿⢩⣍⡛⣿⣿⣿⣿⡇⢚⡀⠀⠀⠀⠁⠔⢠⣾⣿⣿⣿⣿⣿⣿⣉⣐⠛⠛⢉⣥⣶⣿⣿⣿⣿⣿⣿⣿⡿⠋⠀⠀⢠⣾⣿
⠁⣠⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣶⣬⣁⣈⣻⡟⠠⢿⡟⣰⣿⣿⣿⣿⣿⣆⢳⣤⣄⠻⠁⣴⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⠇⠀⠀⠉⠻⠛⠻⢿⣿⣿⠿⠋⠀⠀⠀⢀⣾⣿⣿
⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⡟⢿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣶⣶⣿⣿⣿⣿⣿⣿⣿⡀⠿⠿⠁⠄⠿⠿⣿⣿⣿⣿⣿⣿⣿⡿⠃⠀⠀⠀⠀⠀⠀⠀⠈⠠⠄⠀⠀⠀⠀⠀⣼⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⡿⠿⠿⠇⠀⠙⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠟⠁⠀⢶⡆⠀⢰⠸⠖⢈⣿⣿⣿⠿⠛⠁⣠⣾⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣼⣿⣿⣿⣿
`,
	`
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣝⣛⣛⣓⣒⡢⠀⣶⣿⡖⢀⣾⣿⣿⠟⡕⠁⠀⣶⣄⠀⠀⠀⢂⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⠀⠀⠀⠀⠹⠿⠿⢛⠉⠀⢠⣿⠟⣠⡿⣿⡟⠡⠂⠁⠀⣠⣿⣿⡄⡀⠀⠈⢆⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⡌⢆⠀⠀⠀⣴⢠⡀⠀⠀⠀⡀⣠⣾⡆⠦⠀⠄⠃⢠⠄⢬⠤⣤⣤⣶⣮⣭⠶⣋⣤⣌⠀⣀⣼⣿⣿⣿⡇⣿⣶⡀⡜⡇⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⡌⢧⠀⠚⣛⠘⠛⠓⢸⠂⢆⠙⢶⡆⠀⠀⠀⡼⢸⢘⡃⠁⠀⠒⠓⠒⠂⠀⢿⣿⠟⠀⢻⣿⣿⣿⣿⡇⢻⣿⣧⠀⠘⢹⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣄⠀⠘⠀⠀⠈⠃⠀⠐⠁⠀⠀⠁⠈⠁⠀⠀⣼⡇⡟⣼⠇⠀⠀⠀⣠⣤⣶⣾⣦⣤⠰⠀⣸⣿⣿⣿⣿⠇⠈⣧⡻⠸⢰⡄⢻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣷⣄⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣠⢺⣿⢡⢃⣿⠀⠀⣼⣷⣌⡻⠿⠿⠟⣡⠀⠀⣿⣿⢛⠿⢃⠀⢰⡘⣿⡆⠈⢱⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⡝⣿⣦⣀⠀⠀⠀⠀⠀⠀⢀⣴⣿⡿⣸⡟⡼⣸⡇⠀⢠⣿⣿⣿⣿⣿⣿⣿⣿⠀⠀⣿⢹⠘⣛⢻⠀⢸⡿⠌⠁⠀⣇⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⡇⣿⣿⡟⣰⣄⡀⠀⣀⣴⣿⡿⢫⣾⣿⢣⢃⣿⠁⠀⢸⡿⢛⣩⣥⣌⠻⠛⠟⠀⠀⣿⢸⣿⣿⣦⠀⢀⣴⡆⠀⢰⣿⠘⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⠇⣿⣿⠇⣿⣿⣿⣾⣿⠟⣫⣾⣿⣿⡏⡜⣼⡏⠀⠀⣷⣾⣿⣿⣿⣿⣾⣿⡆⠀⠀⢻⢸⣿⣿⡏⠀⠈⣿⡇⠐⠀⣿⠀⢹⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⠿⣫⣾⣿⡿⢸⣿⣿⣿⡿⣡⣾⣿⣿⣿⣿⢱⢣⡿⠀⠀⢸⣿⣿⣿⣿⣿⣿⣿⣿⡇⢠⠀⢸⢸⣿⣿⠃⠆⢸⣿⠁⠀⠀⢛⣀⣾⣿⣿⢫⣶⢹⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⡇⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⢃⢏⣾⠃⠀⠀⣿⣿⣿⣿⣿⣿⣿⣿⣿⡇⢸⣀⣈⣈⣛⣩⣴⣷⣬⠛⠀⠀⠀⢋⣍⢿⣿⡇⣿⣿⢸⣿⣿⣿⣿⣿⠿⢿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿⠏⡎⣼⠃⠀⠀⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿⡇⢸⣿⣿⣿⣿⣿⣿⣿⡏⡄⠀⠆⣸⡸⣿⣷⡙⠇⢿⣿⣧⠻⣿⣿⢏⣴⡾⢡⣿⣿⣿⣿⣿⣿
⣿⣿⣿⡇⣾⣿⣿⣿⣿⣿⣿⣿⣿⡟⡜⣸⠏⠀⠀⠀⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⡇⢸⣯⣭⣭⡝⢿⣿⣿⡇⣡⣴⣡⣿⣷⡘⢿⣿⣷⣬⡛⢿⣷⡝⢋⣾⣿⢣⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣸⣿⣿⣿⣿⣿⣿⣿⢏⡜⣰⠏⠀⢀⠂⣸⣿⣿⣿⣿⡏⣿⣿⣿⣿⣿⡇⠈⠘⠻⠿⠿⣾⣿⣿⢸⣿⣿⣿⣿⣿⢱⣤⣙⠿⣿⣿⣷⣮⢻⣮⢻⡟⣼⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⢣⢎⣼⠏⠀⢀⠆⢠⣿⣿⣿⣿⣿⡇⣿⣿⣿⣿⣿⣿⠀⣶⣤⣥⣴⣾⣿⣿⡎⢿⣿⣿⣿⣿⣦⡙⢿⣿⣶⣮⣭⣛⢧⢻⡜⡇⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟⡵⣡⡾⠃⠀⣠⠋⢀⣾⣿⣿⣿⣿⣿⡇⣿⣿⣿⣿⣿⣿⡄⣿⣿⣿⣿⣿⣿⣿⣿⣎⠻⣿⣿⣿⣿⣿⣶⣍⡛⠿⣿⣿⣎⢧⠁⢣⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⡿⣣⢊⣴⠟⠁⢀⣴⠃⢠⣾⣿⡿⣿⣿⣿⣿⡇⣿⣿⣿⣿⣿⣿⡇⢻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣦⣹⣿⣿⣿⣿⣿⣿⣧⠹⣿⣿⣎⢷⡜⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⡿⣫⠜⣡⡿⠋⠀⣠⡾⠁⠀⢸⣿⣿⠇⣿⣿⣿⣿⡇⢿⣿⢸⣿⣿⣿⣇⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⢹⣿⣿⡟⣻⣛⢿⣿⠀⢹⣿⣿⣮⣷⡸⣿⣿⣿⣿⣿⣿⣿
⣿⣿⡿⢟⡩⢞⣡⡾⠋⡄⠀⣰⡟⠁⠀⠀⣿⣿⣿⢠⣿⣿⣿⣿⡇⢸⡇⠀⣿⣿⣿⣿⠘⣿⣿⣿⣿⣿⣿⣿⣿⣿⢱⣿⣿⣿⡇⣿⣿⣆⠃⣾⣿⣿⣿⣿⣿⣇⢿⣿⣿⣿⣿⣿⣿
⣿⡿⢗⣩⣶⠟⣋⣴⣾⠃⣴⡏⠀⠀⠀⢰⣿⣿⡏⣸⣿⣿⣿⣿⡇⣸⡇⠀⠸⣿⣿⣿⡆⢻⣿⣿⣿⣿⣿⣿⣿⡿⢸⣿⣿⣿⣷⢹⣿⣿⡄⢹⣿⣿⣿⣿⣿⣿⢸⣿⣿⣿⣿⣿⣿
⣿⣿⠟⢋⣤⣾⣿⣿⡇⢰⡟⠀⠀⠀⠀⣾⣿⡿⠀⣿⣿⣿⣿⣿⡇⣿⡇⠀⠀⢻⣿⣿⣷⠘⣿⣿⣿⣿⣿⣄⠀⣦⣈⡛⠿⠿⠿⠘⣿⣿⣧⢸⣿⣿⣿⣿⣿⣿⠀⢿⣿⣿⣿⣿⣿
⣉⣴⣾⣿⣿⣿⣿⣿⠁⢸⠀⠀⠀⠀⢰⣿⣿⠃⢸⣿⣿⣿⣿⣿⡇⣿⣧⠀⠀⠀⢻⣿⣿⣧⠘⣿⣿⣿⣿⡿⢃⣿⣿⣿⣶⣶⣿⡇⣿⣿⣿⢸⣿⣿⣿⣿⣿⣿⡇⡘⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⡇⠈⠀⠀⠀⢀⡿⢹⡟⠀⣾⣿⣿⣿⣿⣿⡇⣿⣿⠀⠀⠀⠀⠻⣿⣿⣧⡘⢿⣿⣿⢡⣿⣿⣿⢿⣿⡿⠟⠋⠘⣿⣧⣿⣿⣿⣿⣿⣿⣿⡇⠃⢻⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⡇⠀⠀⠀⠀⣼⠁⢸⠃⠀⣿⣿⣿⣿⣿⣿⠀⣿⣿⠀⠀⠀⠀⠀⠈⠻⣿⣿⣦⣙⠋⠼⠿⠟⠁⢸⠵⠖⣒⣩⣤⠹⣿⣿⣿⣿⣿⣿⣿⣿⠀⠀⠘⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⠁⠀⠀⠀⣰⠆⠀⡜⠀⢸⣿⣿⣿⣿⣿⣿⠀⣿⣿⡇⠀⠀⠀⢀⠀⠀⠈⠛⠿⠿⠿⠗⠂⠀⠀⠰⣾⣿⣿⣿⣿⣧⠻⣿⣿⣿⣿⣿⣿⣿⣇⠀⠀⠹⣿⣿⣿
⣿⣿⣿⣿⣿⡿⢹⢿⠀⠀⠀⢠⡏⠀⠀⠀⠀⣸⣿⣿⣿⣿⣿⣿⠀⢻⣿⣿⠀⠀⠀⠘⢳⣦⣄⣀⡀⠀⠀⣀⣀⣴⣿⣦⡈⠻⠿⣿⡿⠿⣃⠙⣿⣿⣿⣿⣿⣿⣿⢳⡄⠀⣥⣤⣿
⡙⣿⣿⣿⡿⠡⢡⡆⣤⠄⠀⣿⠁⠀⠀⠀⠀⣿⣿⡏⢿⣿⣿⣿⠀⢸⣿⣿⡆⠀⠀⠀⠀⢻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣶⡾⠟⣡⣾⣌⠻⣿⣻⠿⣛⠁⠈⣠⠞⣡⡿⢃
⣷⣌⠿⠿⠡⠡⠟⠀⠀⠀⣸⡏⠀⠀⠀⠀⠀⣿⣿⠃⢸⣿⣿⣿⠀⠈⣿⣿⣿⡀⠀⠀⠀⠀⠙⢿⣿⣿⣿⣿⣿⣿⡏⢛⣛⣋⠉⠥⣴⣾⣿⣿⣿⣷⣌⡻⠋⠁⢠⠞⢁⣾⠟⢁⣾
⣿⠻⣿⣷⣾⣿⣿⣿⡇⠀⣿⠁⠀⠀⠀⠀⠀⣿⣿⠀⠘⣿⣿⣿⡆⠀⢻⣿⣿⣷⡀⣿⠶⣦⣤⣤⣽⣿⣿⣿⣿⣿⣷⡌⢿⣿⣷⣦⡀⠉⠻⠿⣿⣿⣿⣿⠀⢂⣠⠐⠛⠁⣴⣿⣿
⣿⣷⣌⡻⢿⣿⣿⣿⡇⢠⣿⠀⠀⠀⠀⠀⠀⢻⣿⠀⠀⢿⣿⣿⣧⠀⠈⣿⣿⣿⣇⢠⣀⠀⢀⡀⣭⠙⠛⠿⢿⣿⣿⣿⣎⢻⣿⣿⣿⣷⣶⣦⡈⠩⢄⠃⠠⠄⠀⠀⢾⣿⣦⡻⣿
⣿⣿⣿⣿⣷⣦⣭⣙⡃⠘⣷⠀⠀⠀⠀⠀⠀⠘⣿⠀⠀⠸⣿⣿⣿⡀⠀⠸⣿⣿⡟⠀⠹⢿⣶⣦⣍⣁⣘⡲⠦⠌⠻⣿⣿⣦⡻⣿⣿⣿⣿⣿⣿⣷⠐⡀⠢⠤⠀⢀⠀⠙⢿⣿⣾
⣿⣿⣿⣿⣿⣿⣿⣿⡇⠀⣿⡆⡆⠀⢠⠀⠀⠀⠘⠀⠀⠀⢻⣿⣿⣇⠀⠀⢹⣿⣿⡀⣀⣀⠀⢌⣙⠛⠛⠛⠛⠋⠳⠈⠻⠿⣿⣆⠉⢛⠻⢟⣿⠇⡰⣿⣶⠖⢠⣿⠿⣓⣶⣝⢿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⡄⠘⣧⠱⡀⠀⢆⠀⠀⠀⠀⠀⠀⠈⢿⣿⣿⡄⠀⠀⢿⣿⡇⢹⣿⣿⣶⣌⠳⢦⡈⠻⢿⣤⣝⡒⠆⠀⠠⠉⢀⠉⡒⢈⣀⠁⣿⡟⠀⢞⣵⣿⣿⣿⣿⣷
`,
	`
⣿⣿⣷⣮⡁⢎⠷⣏⡲⣄⣎⠵⣩⢿⣿⣿⣿⣿⢳⡒⢆⠰⢀⢂⠇⡒⢄⠢⡐⢢⠘⠴⣈⠦⢱⢈⠴⡐⡜⠒⡀⠂⠐⠠⢀⠀⣠⠆⡰⢔⡀⠀⢂⡉⠟⡹⠛⡹⠛⠿⢛⡍⣒⠠⢀
⢛⠿⣿⣿⣿⣶⡝⢣⠿⢚⣋⣧⣥⣯⣬⣍⣛⣛⠻⢽⣮⣗⢃⢎⡘⢄⢊⠔⣡⠂⣍⠲⡱⢘⡰⢈⠆⡱⢈⠥⡐⠠⠁⠄⣠⣾⠏⡔⠘⠠⣩⠄⢃⠐⠠⠐⠠⠁⠈⠀⡁⠠⠄⢣⠀
⠠⢍⠸⣉⠛⢉⣠⣴⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣦⡙⢿⣦⣵⣌⢦⣑⣆⢣⢤⣓⡱⢎⣴⢣⣌⡰⣌⣶⡩⢆⠁⣴⣿⡟⠰⢤⡉⡃⡔⢫⠀⢈⢆⡁⡀⠀⠄⡀⠠⢡⠈⠀⠌
⠱⣈⠦⢃⣴⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣌⢻⣿⣿⣿⣾⣿⣾⣼⣵⣿⣿⣿⣮⣷⣯⣻⣵⢋⣼⣿⣿⠃⢀⠂⢣⠣⠔⡉⢎⡀⢾⣷⣷⣿⡶⣥⢃⣦⣌⢆⠻
⠁⠆⣶⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠿⠿⠿⠿⠿⣿⣿⣧⡙⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⢃⣼⣿⣿⡏⠀⠀⠄⠈⡌⡅⡌⢸⠄⢰⣯⢿⡿⣝⢾⢳⢃⡌⢎⠱
⠀⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡃⠤⣈⡙⠛⠿⢿⣿⣶⣶⣮⣭⣈⡻⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠿⠇⢜⣛⣛⣛⡁⠀⠡⠈⠀⠁⡖⣠⠅⣋⠀⣿⣾⣶⣉⠄⢢⠣⡘⢌⢣
⣸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡆⢡⢃⡝⢣⠸⡀⠈⠙⠻⢿⣿⣿⣿⣷⣮⣝⡻⣿⡿⠟⣋⣥⣶⣾⣿⣿⣿⣿⣿⣿⣿⣷⣶⣄⡀⠀⣇⢠⠃⡇⢸⣿⡟⣋⠞⡥⢞⡰⣩⠖
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣧⠀⢲⢈⠆⢃⠡⠈⠐⠠⠀⠉⠛⢿⣿⣿⣿⣟⣥⣶⣿⣿⣿⠏⢿⣿⡄⢈⠻⣿⣿⣿⣯⣭⣝⣿⣷⣄⠂⢮⠱⢸⣿⡼⡱⢎⠰⣡⢾⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡀⠣⢘⡠⢇⠠⡀⠀⠃⠀⠃⠄⠀⠛⣿⣿⣿⣿⣿⣿⣿⣿⠀⠸⠿⠇⡘⠀⣿⣿⣿⣿⣿⣿⣿⣿⣿⣄⠘⢧⢘⣧⣿⣇⡘⣤⣻⡟⢿⣻
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣇⠘⡤⢑⢎⠠⢁⠀⠀⠄⠀⠂⢀⣴⣿⣿⣿⣿⣿⣿⣿⣿⣧⡁⠆⠎⢈⣴⣿⣿⣿⣿⠟⠉⣀⡀⠉⠻⣷⡐⢨⣿⣒⣿⣿⣿⣿⣾⣽⣜
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣄⢒⢊⠜⣢⠡⠄⠀⠂⠐⢠⣾⣿⣿⣿⢯⣵⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡏⠀⠀⣿⣿⣦⢀⠈⢷⡌⠻⠟⠿⣻⡟⢿⡛⢯⠜
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡟⢿⣿⣿⣿⣿⣿⡄⢘⠺⣤⠘⠨⡀⠈⢀⣾⣿⣿⣿⣷⣿⣿⠿⠿⠿⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣧⠀⠀⠸⣿⣿⡇⢧⠈⢳⠈⠈⠀⢡⠉⠄⠈⠄⠀
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡈⣿⣿⣿⣿⣿⣿⣄⠈⠰⢉⡀⢇⠀⣸⣿⣿⣿⣿⣿⠛⠀⠀⠀⣴⣶⣌⡙⢿⣿⣿⣿⣿⣿⣿⣿⡄⠀⠀⠈⠛⠁⠸⡇⠈⣅⠀⠈⠀⠀⠀⠀⠀⠀
⢻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡇⢸⣿⣿⣿⣿⣿⣿⣆⠈⠦⠨⡔⡀⣿⣿⣿⣿⣿⠃⡀⠀⠀⠀⢻⣿⣿⣷⡀⠹⣿⣿⣿⣿⣿⣿⣷⠀⠀⠀⠀⠀⠀⣿⡀⢨⡄⠀⡀⠀⠀⠀⠀⠀
⣎⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡄⣿⣿⣿⣿⣿⣿⣿⣦⡑⠘⣄⡁⣿⣿⣿⣿⠏⠀⡇⠀⠀⠀⠈⠻⣿⣿⡇⢠⢸⣿⣿⣿⣿⣿⣿⣦⣀⠁⠰⠞⠀⣿⡇⢨⣽⣆⠠⢁⠂⠄⠀⠀
⣿⡿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⡜⢿⣿⣿⣿⣿⣿⣿⣿⣌⠂⠆⣿⣿⣿⣿⠆⠀⣧⠀⠀⠀⠀⠀⠈⠉⠀⠈⡟⣿⣿⣿⣿⣿⣿⣿⣿⣷⡤⢔⣺⣭⣷⣿⡿⢛⣅⠘⣋⠬⡐⠌
⠛⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡟⠈⠻⣿⣿⣿⣿⣿⣿⣿⣷⡆⠹⣿⣿⣿⣿⠀⢸⣧⠀⠀⠀⠀⠀⢀⡰⠀⣿⣿⣿⣿⣿⣿⠿⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣯⠔⢀⡈⠤⠐⡀
⠐⡠⠹⢿⣿⣿⣿⣿⣿⣿⣿⠿⠋⠄⠁⡂⠙⢿⣿⣿⣿⣿⣿⣿⣿⡜⣿⣿⣿⠟⣃⠀⠻⣷⣄⢠⡔⢈⠻⠅⢠⣿⣿⣿⣿⣿⣿⣶⠘⠿⢿⣿⣿⣿⣿⣿⣿⣿⣿⢃⣿⠲⣄⢡⡐
⠐⡀⢂⠀⠉⠙⠟⡛⠋⠍⠡⠈⠐⡀⠆⡑⡀⠀⠹⣿⣿⣿⣿⣿⣿⣷⡸⣿⣿⣿⢛⣠⠀⠘⢿⣷⣦⣤⠄⣚⣭⣾⣿⣿⣿⣿⣿⣷⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⢃⣾⣏⣟⣦⣻⣽
⠐⡀⠂⠄⠈⠰⢠⣀⢥⡬⣔⠢⢁⢰⡈⣴⠷⣔⠣⠌⢿⣿⣿⣿⣿⣿⣷⡘⣿⣿⣿⣷⣶⣦⣤⣿⣫⣶⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟⣰⣿⣿⣿⣿⣿⣿⢿
⠀⠀⠀⠀⠀⣞⣧⣿⣟⡞⢭⠓⣌⢢⣝⢮⡛⠈⠁⠀⠀⠹⣿⣿⣿⣿⣿⣷⣌⢻⣿⣿⣿⠿⠿⢿⡿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟⣡⣾⣿⣿⣿⣿⣿⣿⡽⣞
⠀⠀⠀⠀⠀⢏⡿⣿⣻⢚⡣⢟⣸⣷⡾⣧⠅⠀⠀⠀⠀⠀⠘⢿⣿⣿⠿⢟⣛⣥⣍⠰⠶⣚⣭⡵⢞⣹⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠟⢋⣴⣿⣿⣿⣿⡻⢿⠻⢛⡝⠳⢌
⠀⠀⠀⠀⠀⠀⠜⠢⡑⢎⡭⣏⠷⣻⣿⣷⠀⠀⠀⡀⠄⠀⠀⠈⢱⣶⣿⣿⠟⣋⣥⣶⡬⢙⡩⠶⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠿⠟⠋⠉⠀⡚⡽⣏⣷⡻⢿⢯⡳⢡⠃⢌⠰⡹⣌
⠀⠀⠀⠀⠀⠤⡘⡔⠑⢎⡷⣞⡿⣿⣿⠋⠀⠀⠁⠀⣠⣶⡇⠀⠈⢛⣩⣶⣿⣿⠟⣋⣴⣿⣿⣷⣶⣤⣉⡉⠉⠍⠩⢍⠥⠂⣤⠤⠀⡁⢠⡙⠶⣍⠆⣉⠊⡥⣙⡦⣍⠢⢌⠱⣊
⠀⠀⠀⢠⡍⠒⠁⠀⠀⠀⠀⠉⠛⠉⠀⠀⡄⠈⣴⣿⣿⡟⠋⠀⠀⠘⣿⣿⡟⢡⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣶⣤⠀⡄⠀⠀⠀⠀⠀⣤⠘⣵⢪⢱⣬⢳⡜⣿⣿⣿⣽⣮⣷⣽
⢀⡒⢌⠂⠉⠀⠀⠀⠀⠀⠀⠀⠀⠀⠄⢁⣴⣾⡿⠟⠉⠀⣠⣤⠀⠀⠹⢫⣼⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣷⣯⣶⡄⢀⠚⡰⢿⣴⣿⣳⢮⣻⣿⣽⣿⣿⣿⣿⣿⣿
⠨⠐⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⠀⣡⣾⣿⠟⠋⠀⣀⣴⣿⣿⣿⡧⠀⠀⠹⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠿⠃⣤⣂⢦⣹⢿⣿⣿⣿⢟⢮⣿⣿⣿⣿⣿⣿⣿⣿
⠀⡐⠀⠀⠀⠀⠀⠀⠀⠀⢀⣴⣾⡿⠛⠀⢀⣤⣾⣿⣿⣿⣿⢏⣴⣷⡄⠀⠹⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟⣩⢂⣾⣿⣿⣿⢯⣿⣿⣿⣿⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿
⠀⡄⠀⠀⠀⠀⠀⠀⣀⠶⡯⠙⠁⠀⣠⣶⣿⣿⣿⣿⣿⣿⣣⣿⣿⣿⣿⣆⠀⠘⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⢛⣵⡾⢡⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⢶⣼⣆⠀⠀⢀⣴⢪⡑⠋⠀⣠⣴⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⡄⠈⢿⣿⣿⣿⣿⣿⣿⣿⣿⡇⣼⡟⠐⡏⠛⠩⠘⡽⣿⢿⡿⣿⢿⣷⡟⡻⢿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣶⣴⣿⠞⠁⢀⣤⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟⡟⡹⡉⢗⡋⠞⣙⢻⣿⣦⡀⠹⣿⣿⣿⣿⣿⣿⣿⡇⡟⠀⠉⠤⠁⢂⠁⡐⢩⢧⢻⣾⣿⣿⣿⣿⣶⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣧⣤⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣏⡜⢠⠑⡭⢲⣜⠲⡄⢆⠛⣯⣟⢧⠙⣿⣿⣿⣿⣿⣿⡇⠀⡀⠌⡐⡁⢂⡔⣠⢍⢞⣿⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣻⣿⣿⣿⣿⣿⣿⢟⡟⠿⢟⡹⠿⣻⠟⡿⣿⡿⡟⢷⣌⣧⣿⣿⣿⡷⣹⠜⣤⢋⡴⣩⣇⢣⠈⢿⣿⣿⣿⣿⡇⠠⠐⡠⠂⡍⡖⢯⡵⣮⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
`,
	`
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠿⠿⠛⠛⠛⠛⠻⠿⠿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠿⠛⣉⠄⣢⣴⣟⣫⣤⣾⢿⣿⣷⡶⢦⣬⣉⠻⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟⣡⡴⠋⢑⣬⣴⣿⣿⡻⣿⣿⣶⣝⠻⣿⣷⣾⣿⢿⣦⡉⠻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟⣡⡾⠫⣠⣾⣿⣿⣿⣿⣿⣷⢹⣿⣿⣿⣷⡙⢿⣿⣿⣧⡐⡝⣦⡙⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⢃⣼⣿⠀⣼⣿⣿⣿⣿⣿⣿⣿⣿⣇⣿⣿⣿⣿⣿⡌⠙⢿⣿⣿⡐⣿⣷⠈⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠁⣹⢻⡟⡘⣿⠇⣿⣿⣿⣿⣿⣿⣿⡏⣿⠻⣿⣿⣿⣿⡌⢷⡉⢙⠀⠈⠀⡶⠈⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡟⠀⠐⠀⠾⢠⣷⣜⢸⣿⣿⡇⣿⣿⣿⣿⡇⢻⣧⢻⣿⣿⣿⣷⡀⡁⠀⠁⡁⣦⣄⠁⠘⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠀⢀⠑⣸⡦⣈⡻⠇⢨⢹⣿⡧⣿⣿⣿⣿⡇⣘⣿⡜⣿⣿⣿⢿⢇⠀⠀⣧⢱⡹⣷⡌⠂⠹⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⡟⢠⡄⡈⠻⣿⣿⣿⣿⣿⣿⠁⠌⢀⢇⡿⠀⣿⣿⡇⣦⣾⣿⣃⣿⣿⣿⣿⡇⠸⣟⢃⢛⣋⡴⠂⠎⡀⠘⣿⡌⣷⡘⣿⡄⠀⠘⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⡿⢀⠟⢜⣼⠀⣼⣿⣿⣿⣿⠇⠌⠀⣸⠸⡇⠀⡇⠟⣃⣿⣿⣿⢸⣿⢿⣶⣭⠁⡁⠹⡠⠌⢉⣬⣉⣀⡃⠀⠸⣷⢸⣧⡹⣷⡀⢧⠈⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣷⡈⠲⠛⢉⠀⢻⣿⣿⣿⡟⠀⣼⠀⣿⢰⡇⠀⠀⣿⡇⣿⣿⡟⢨⡿⣸⣿⡟⢠⣿⣄⠁⠀⣿⣿⣿⣿⢰⠀⠀⢿⡆⢿⣷⡙⣷⡈⠀⢹⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⡾⠛⠻⣾⣿⣿⣿⠃⢰⡏⠠⣿⠸⣧⠀⠀⠁⠁⠹⡿⠁⣼⠃⡟⠉⠀⠒⠈⠉⠁⠀⠛⣿⣿⣿⢸⢐⠲⠘⣿⠈⣻⣷⡌⢿⡄⢾⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣅⠀⢀⣿⣿⣿⣿⢀⡿⠁⠘⣿⢨⢻⡀⠀⢦⠂⠀⠠⣤⣥⣤⣦⣶⡆⠀⠀⠙⣿⡇⢀⣿⣿⣿⡏⢐⠳⡄⢿⡇⡟⡏⢻⣆⢈⠀⠙⠛⣛⠉⢸⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⣿⣿⣿⠸⠁⠀⡇⡟⠘⡘⣧⠀⢸⣇⠀⡀⢹⣿⣿⣿⣿⣧⠐⣸⠀⣽⠇⡏⣿⣿⡿⠇⢘⠵⠃⢸⡇⠃⠇⠈⡜⡄⠻⣶⣦⣤⣶⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠀⠌⠛⠋⠀⠰⠀⠁⡇⠀⢧⠘⠧⡀⠿⢦⠡⣾⣿⣿⣿⣿⡿⢓⡿⠶⡟⠘⢰⣿⣿⢱⠀⠀⠀⠀⣼⠇⢰⠀⠀⢱⡀⢧⢹⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣶⣶⠟⠃⠀⠀⠀⠃⠀⠘⡀⠀⠀⠐⢄⣠⣿⣿⣿⣿⣿⣯⡐⣜⢂⡠⠂⣿⣿⡧⢸⠀⠀⠀⢠⠋⠀⠈⠀⠀⠘⠀⠈⢂⣿⣿⣿⣿⣿⣿⣿
⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠃⠆⠀⠸⠇⠀⠀⠀⠄⠥⠀⢀⠀⠀⠙⠛⠿⢶⣼⣿⡿⠿⠛⢉⣤⢰⣿⡿⠃⡇⠀⠀⠀⠀⠀⠀⠀⡴⠀⠀⠀⠀⢸⣿⣿⣿⣿⣿⣿⣿
⠀⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣭⣭⣭⡒⡄⠂⠀⠀⠀⠀⠀⠀⡀⣀⠀⢰⣿⢃⣾⠿⠁⠀⠀⠀⠀⢀⠀⠀⠠⢄⠀⠀⢀⣜⣠⣿⣿⣿⣿⣿⣿⣿⣿
⣇⢹⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⡘⡀⠐⢂⣀⣤⣴⠊⡠⠴⠀⠉⠡⠛⠁⠀⠀⢠⡶⢀⡴⣄⡐⢶⡄⢠⡀⢺⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⡄⢻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣧⢡⠀⠹⢿⣿⢃⢰⣶⣥⣒⡶⠟⣓⣤⣤⡾⢋⠔⣩⣾⣿⣿⠖⣠⣾⠇⠸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣷⡈⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡆⠂⠀⠀⠁⣌⢦⠙⢛⣣⣵⣾⣿⠿⢋⣐⣁⣨⣭⣭⣤⣤⣤⣤⣤⣬⣀⠙⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣧⠘⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣏⠻⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡈⠄⠀⠀⣫⡕⣬⣓⡲⣶⠖⠂⡄⠛⠉⠙⣿⣿⣿⡿⠿⠛⠋⠁⠀⠀⠀⠀⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣆⢹⣿⣿⣿⣿⣿⣿⣿⣿⣿⠏⠁⢰⡙⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣧⠐⠀⠀⣿⣿⠸⣿⠟⣱⣾⡆⢱⢠⢰⠈⠉⣀⣤⢠⣤⣤⠔⣠⣶⡀⠀⢀⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⡄⢻⣿⣿⣿⣿⣿⣿⣿⣿⣆⠀⠈⣷⡈⢛⣿⣿⣿⣿⣿⣿⣿⣿⣿⡆⠁⠈⢿⠿⠿⠈⢺⣿⣿⣷⠈⡌⢸⠀⣿⣿⠇⣾⡟⢁⠠⠤⠄⠠⢀⠸⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⡈⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡐⡄⠰⢊⣀⣀⡛⠻⣿⡿⠀⡇⠆⠀⠿⠋⠀⠋⠀⠤⣴⣶⣶⢶⣤⠀⠘⢿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣧⠘⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⢱⠀⠚⠉⣩⡶⠀⢀⠀⠀⢀⠀⠠⣀⣀⠀⠀⠀⢐⣄⠀⠀⠀⠈⢂⠀⢨⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣇⠸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣇⢧⠀⠖⣉⠤⠴⢂⣀⠈⡓⠦⠄⠀⠀⠀⠐⠤⣌⠛⠓⢄⠀⠀⠀⠀⠘⣿⣿⣿⣿⣿⣿⣿
⣿⠀⠤⠄⠀⢀⣀⣠⣨⣉⣉⣉⣉⣛⣛⡛⣛⢛⡛⠛⠛⠛⠛⠻⠿⠿⠿⠿⠿⠿⠿⠿⠜⡆⠈⠉⠀⠀⠙⠋⠠⠤⠐⠛⠀⠀⠀⠀⠀⠈⠳⠀⠀⠀⠀⠀⠀⠀⠛⠛⠛⠛⠛⠛⠻
⣿⣶⣦⣤⣤⣤⣤⣤⣄⣀⣀⣀⣈⣉⣉⡉⠉⠉⠉⠉⠛⠛⠛⠛⠛⠚⠓⠒⠒⠶⠖⠲⠦⠰⠶⠰⠂⠉⠉⠉⠛⠛⠓⠛⠛⠁⠀⣉⣁⣀⣀⣀⣀⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣼
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣶⣶⣶⣶⣶⣶⣤⣤⣤⣤⣤⣤⣴⣶⣶⣶⣶⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
`,
	`
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⢋⠉⠉⢙⣟⣋⣭⣙⡛⡛⡛⠛⣿⣿⣟⣛⣻⢛⣛⣛⣟⣛⣿⣿⡟⢛⣻⣿⣻⡛⢛⢿⡹⢍⣛⣛⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣧⢠⠀⠀⢈⠳⠾⠙⠀⡆⠇⢱⠘⠀⠈⠇⡆⡆⡁⠲⢈⠺⡆⣿⣿⡅⠀⠃⠀⠁⢰⠀⠀⡆⠈⠰⠀⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣦⠍⡀⠀⠉⡉⡉⠈⠁⠀⠀⠉⠉⠉⠉⠉⠩⣭⣤⣬⣿⣿⣯⣭⣽⣾⣿⣾⣿⣿⣷⣾⣿⣿⣷⣾⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣇⠀⠂⠀⠀⠀⠀⠌⠃⠀⠀⠁⠀⠀⠀⠀⠀⣿⣿⡟⠀⠀⠄⠠⣯⣀⣀⣀⣤⣠⣆⣡⣆⣣⣇⣀⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣶⣾⣿⣿⣿⣿⣿⣷⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣶⣶⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⠟⠛⠛⠛⠛⠛⠛⠿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠿⠛⠛⠋⠉⠉⠉⠉⠉⠙⠛⠛⠿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣶⣶⣿⣿⣿⣷⣶⣄⠈⢻⣿⣿⣿⣿⣿⣿⣿⣿⠿⠛⠋⣉⡀⢀⣠⣤⣶⣾⣿⣿⣿⣿⣿⣿⣿⣶⣶⣤⣀⠉⠛⠿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⣿⣧⠀⢻⣿⣿⣿⣿⠟⠋⣀⣴⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣶⣄⠈⠙⠛⠻⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⡿⠿⠛⠛⠛⠻⠿⠄⠀⠿⠛⠉⣀⣤⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣄⠐⣦⣄⠙⠻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⡿⠋⢀⣤⣶⣶⣶⣶⣶⣶⣶⣶⣶⣿⣿⣿⣿⣿⠟⢩⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣄⠈⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⠁⢰⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠃⣰⣿⡿⢛⠛⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⡀⠙⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⠀⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠁⣸⣿⡿⢡⣏⠇⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠿⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡄⠈⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣀⠀⠻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟⠁⠀⢫⣙⠻⣌⣋⣴⣿⣿⣿⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⢁⣶⣆⢹⣿⣿⣿⣿⡇⢸⣿⣿⣿⣿⣿⣦⡀⠙⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣷⣄⠈⠛⠿⠿⠿⠿⠿⠟⠛⠉⣀⣴⣾⣆⠈⠻⣿⡜⣿⣿⣿⣿⠟⠈⠛⠛⠻⠟⢻⣿⣿⣿⣿⣿⠸⣿⡏⣸⣿⣿⣿⣿⡇⠈⣿⣿⣿⣿⣿⣿⣿⣦⣀⠉⠛⠿⢿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣶⣦⣤⣤⣤⠴⠶⠶⠿⠛⠛⠛⠉⠁⣀⣀⡉⠛⠿⠿⠁⠀⡆⡆⡾⢠⡄⠘⣿⣿⣿⣿⣿⠗⣠⣬⣭⣙⡻⢿⣿⣷⠀⢻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣶⣤⣀⡈⠙⠻⣿⣿
⠟⠛⠛⠉⢉⣉⣡⡀⢀⣤⣄⢶⠶⣡⣌⠛⢃⣤⡙⢿⢟⡠⠀⣠⣶⠀⠀⢠⠃⡟⡴⠀⠻⠿⠿⠿⠿⣸⣿⣿⠿⠿⠿⠆⠋⠁⣀⠈⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣶⣄⠈⠻
⣀⠤⠀⠶⠦⠄⣉⠳⢦⡙⠟⣁⡘⠿⢋⣴⣌⠻⠟⣠⡙⠧⠀⠋⠁⠀⠀⢸⠸⠰⠁⣰⣶⣶⣦⣤⣤⣤⣄⡀⠀⠠⣀⠀⢺⣿⣿⡄⠘⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡇⠀
⠡⢐⣦⡐⠒⣠⡈⠁⡀⠉⣤⠛⢋⣤⣝⢿⢟⣡⡌⠋⠁⠊⠀⣼⣿⣄⣀⠈⠀⣀⣼⣿⣿⣿⣧⡈⠻⢿⣿⣿⡦⠀⢹⣷⡀⠹⣿⣿⡄⠈⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠇⠀
⡀⠈⠋⣩⠤⠤⣤⣄⠉⣄⢩⣴⣌⠻⠟⣠⡘⠛⠀⣴⣷⣄⣸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⢿⣿⣿⣦⣤⣀⡀⠀⣴⠀⣿⣧⠀⣿⣿⣿⣦⡀⠈⠛⠿⣿⣿⣿⣿⣿⣿⣿⡿⠋⠀⣼
⠰⢰⠀⠰⠒⣀⣼⣿⡆⠉⠀⠻⠟⣡⣦⠙⢡⡆⠀⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣟⠁⣀⣄⡈⠻⣿⣿⣿⡇⠀⢁⣰⣿⠇⢀⣿⣿⣿⣿⣿⣷⣦⣄⣀⠈⠉⠉⠉⠉⢀⣠⣴⣿⣿
⠃⠀⠀⠘⠛⠛⠋⢀⣂⣈⣠⠀⣌⠛⣡⣾⣮⡛⠀⠘⠿⠟⠛⠻⠿⣿⣿⣿⣿⣿⡇⠀⣿⣿⣿⣾⣿⣿⠟⠁⠰⠿⠛⠁⠀⣤⣌⠙⠻⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣍⣉⣉⣉⣉⣉⣉⣭⣥⣤⡤⠀⢡⣶⣌⠻⢋⣴⣦⠀⢤⣴⣔⠶⣠⣤⡀⠉⠉⠉⠃⠀⢻⣿⣿⠟⠉⢀⡠⢢⣤⣄⠀⣠⣦⡙⠿⢋⣦⡀⢈⠙⠻⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣚⣻⣿⣿⣿⣿⣿⣿⣿⡿⠁⣼⣦⡙⢋⣤⣌⠛⣡⣾⣦⡙⢫⣤⡙⢋⣴⣷⣌⠛⣡⣦⠀⠉⢁⣠⡘⢋⣴⣌⠛⣡⣶⣌⠻⢋⣤⡙⢋⣴⣦⡙⠖⣠⡈⠛⠻⣿⣿⣿⣿⣿⣿⣿⣿
⠿⢿⣿⣿⣿⡿⠿⠛⠉⣠⣮⡛⢛⣴⣦⠙⢡⣶⣌⠻⢋⣴⡌⠋⣴⣦⡝⠟⣡⣶⡌⠡⣴⣮⡛⢋⣴⣦⠙⢡⡶⣌⠻⢋⣴⡌⠋⣴⣦⡙⠟⠡⠶⠈⠁⠘⣀⣀⣹⣿⣿⣿⣿⣿⣿
⣶⣄⠀⢠⣤⠀⢠⣾⣷⡌⠋⣴⣮⠙⣥⣿⣦⡙⢋⣴⡌⠋⣴⣷⣌⠛⢡⣦⡙⢫⣾⣷⡌⠋⢀⣀⠋⣴⣛⣆⣧⠋⠐⠈⠋⢀⣁⣨⣤⣤⣴⣶⣶⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⠟⢡⣾⡦⠁⣼⣷⡝⠛⣵⣷⠈⢠⣾⢮⠛⢫⣾⡦⠉⢴⡷⡌⠛⣡⣾⠆⠉⣰⣇⣩⣿⣿⣦⣼⣿⣿⣿⣿⣿⣿⣷⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣷⡌⠉⣼⣿⡎⠋⢤⡆⠙⢡⡾⣷⢹⣷⣄⠂⠉⣤⡛⠦⢰⣷⣤⣈⣉⣠⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣾⣷⣮⢋⣴⣦⠀⢀⣿⣿⣿⣶⣾⣿⣿⣷⣦⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣯⣴⣿⣿⣿⣷⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
`,
}
