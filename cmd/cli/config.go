package cli

import (
	"os"
	"strings"

	"github.com/caner-cetin/strafe/internal"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	sectionColor = color.New(color.FgBlue, color.Bold)
	keyColor     = color.New(color.FgCyan)
	valueColor   = color.New(color.FgGreen)
	secretColor  = color.New(color.FgYellow)
)
var (
	printSensitiveCFGVars bool
	configCmd             = &cobra.Command{
		Use:   "cfg",
		Short: "print config variables and exit",
		Run:   displayConfig,
	}
)

func getConfigCmd() *cobra.Command {
	configCmd.PersistentFlags().BoolVarP(
		&printSensitiveCFGVars,
		"sensitive",
		"s",
		false,
		"print sensitive configuration variables such as password, set to false by default",
	)
	return configCmd
}

func displayConfig(cmd *cobra.Command, args []string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Section", "Key", "Value"})
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(true)
	table.SetRowLine(true)

	password := viper.GetString(internal.CREDENTIALS_PASSWORD)
	if !printSensitiveCFGVars {
		password = strings.Repeat("*", len(password))
	}

	section := sectionColor.Sprint("Uploader")
	table.Append([]string{
		section,
		keyColor.Sprint("Username"),
		valueColor.Sprint(viper.GetString(internal.CREDENTIALS_USERNAME)),
	})
	table.Append([]string{
		"",
		keyColor.Sprint("Password"),
		secretColor.Sprint(password),
	})

	section = sectionColor.Sprint("Docker")
	table.Append([]string{
		section,
		keyColor.Sprint("Image Name"),
		valueColor.Sprint(viper.GetString(internal.DOCKER_IMAGE_NAME)),
	})
	table.Append([]string{
		"",
		keyColor.Sprint("Image Tag"),
		valueColor.Sprint(viper.GetString(internal.DOCKER_IMAGE_TAG)),
	})
	table.Append([]string{
		"",
		keyColor.Sprint("Socket"),
		valueColor.Sprint(viper.GetString(internal.DOCKER_SOCKET)),
	})
	section = sectionColor.Sprint("Database")
	db_url := viper.GetString(internal.DB_URL)
	if !printSensitiveCFGVars {
		db_url = strings.Repeat("*", len(db_url))
	}
	table.Append([]string{
		section,
		keyColor.Sprint("URL"),
		valueColor.Sprint(db_url),
	})
	table.Render()
}
