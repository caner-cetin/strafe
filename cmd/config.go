package cmd

import (
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	STRAFE_CONFIG_LOC_ENV = "STRAFE_CFG"
	CREDENTIALS_USERNAME  = "credentials.username"
	CREDENTIALS_PASSWORD  = "credentials.password"
	DOCKER_IMAGE_NAME     = "docker.image.name"
	DOCKER_IMAGE_TAG      = "docker.image.tag"
	DOCKER_SOCKET         = "docker.socket"
	DB_PATH               = "db.path"
)

const (
	DOCKER_IMAGE_NAME_DEFAULT = "strafe"
	DOCKER_IMAGE_TAG_DEFAULT  = "latest"
)

var (
	sectionColor = color.New(color.FgBlue, color.Bold)
	keyColor     = color.New(color.FgCyan)
	valueColor   = color.New(color.FgGreen)
	secretColor  = color.New(color.FgYellow)
)

type ContextKey string

const (
	APP_CONTEXT_KEY ContextKey = "strafe_ctx.app"
)

var (
	PrintSensitiveCFGVars bool
	verbosity             int
	configCmd             = &cobra.Command{
		Use:   "cfg",
		Short: "print config variables and exit",
		Run: func(cmd *cobra.Command, args []string) {
			log.WithFields(log.Fields{
				"username_set": viper.IsSet(CREDENTIALS_USERNAME),
				"password_set": viper.IsSet(CREDENTIALS_PASSWORD),
			}).Debug("uploader settings")
			log.WithFields(log.Fields{
				"image_name_set": viper.IsSet(DOCKER_IMAGE_NAME),
				"image_tag_set":  viper.IsSet(DOCKER_IMAGE_TAG),
				"socket_set":     viper.IsSet(DOCKER_SOCKET),
			}).Debug("docker settings")
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Section", "Key", "Value", "Status"})
			table.SetAutoWrapText(false)
			table.SetAutoFormatHeaders(true)
			table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.SetBorder(true)
			table.SetRowLine(true)

			checkmark := color.GreenString("✓")
			warning := color.YellowString("!")

			password := viper.GetString(CREDENTIALS_PASSWORD)
			if !PrintSensitiveCFGVars {
				password = strings.Repeat("*", len(password))
			}

			section := sectionColor.Sprint("Uploader")
			table.Append([]string{
				section,
				keyColor.Sprint("Username"),
				valueColor.Sprint(viper.GetString(CREDENTIALS_USERNAME)),
				checkmark,
			})
			table.Append([]string{
				"",
				keyColor.Sprint("Password"),
				secretColor.Sprint(password),
				warning,
			})

			section = sectionColor.Sprint("Docker")
			table.Append([]string{
				section,
				keyColor.Sprint("Image Name"),
				valueColor.Sprint(viper.GetString(DOCKER_IMAGE_NAME)),
				checkmark,
			})
			table.Append([]string{
				"",
				keyColor.Sprint("Image Tag"),
				valueColor.Sprint(viper.GetString(DOCKER_IMAGE_TAG)),
				checkmark,
			})
			table.Append([]string{
				"",
				keyColor.Sprint("Socket"),
				valueColor.Sprint(viper.GetString(DOCKER_SOCKET)),
				checkmark,
			})
			table.Render()
		},
	}
)

func getConfigCmd() *cobra.Command {
	configCmd.PersistentFlags().BoolVar(
		&PrintSensitiveCFGVars,
		"sensitive",
		false,
		"print sensitive configuration variables such as password, set to false by default",
	)
	return configCmd
}
