package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"os"
	"strafe/internal"
	"strafe/pkg/db"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/qeesung/image2ascii/convert"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	dbCmd = &cobra.Command{
		Use:   "db",
		Short: "database ops",
	}

	searchCmd = &cobra.Command{
		Use:   "search",
		Short: "search albums or tracks",
	}
)

type SearchAlbumConfig struct {
	Name   string
	Artist string
}

var (
	searchAlbumCmd = &cobra.Command{
		Use:   "album",
		Short: "search album",
		Run:   WrapCommandWithResources(searchAlbum, ResourceConfig{Resources: []ResourceType{ResourceDatabase, ResourceS3}}),
	}
	searchAlbumCfg = SearchAlbumConfig{}
)

func getDBRootCmd() *cobra.Command {
	searchAlbumCmd.PersistentFlags().StringVarP(&searchAlbumCfg.Artist, "artist", "a", "", "artist name")
	searchAlbumCmd.PersistentFlags().StringVarP(&searchAlbumCfg.Name, "name", "n", "", "album name")
	searchCmd.AddCommand(searchAlbumCmd)

	dbCmd.AddCommand(searchCmd)
	return dbCmd
}

func searchAlbum(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	app := ctx.Value(internal.APP_CONTEXT_KEY).(internal.AppCtx)

	var artistSet = searchAlbumCfg.Artist != ""
	var nameSet = searchAlbumCfg.Name != ""
	var album db.Album
	var err error
	if artistSet && nameSet {
		album, err = app.DB.GetAlbumByNameAndArtist(ctx,
			db.GetAlbumByNameAndArtistParams{
				Artist: pgtype.Text{String: searchAlbumCfg.Artist, Valid: true},
				Name:   pgtype.Text{String: searchAlbumCfg.Name, Valid: true}})
	} else if artistSet {
		album, err = app.DB.GetAlbumByArtist(ctx, pgtype.Text{String: searchAlbumCfg.Artist, Valid: true})
	} else if nameSet {
		album, err = app.DB.GetAlbumByName(ctx, pgtype.Text{String: searchAlbumCfg.Name, Valid: true})
	} else {
		log.Warn("album artist or name must be specified for search")
		os.Exit(1)
	}
	check(err)
	coverArtBytes, err := app.DownloadFile(ctx, viper.GetString(internal.S3_BUCKET_NAME), album.Cover.String)
	check(err)
	img, _, err := image.Decode(bytes.NewReader(coverArtBytes))
	check(err)

	converter := convert.NewImageConverter()

	asciiArt := converter.Image2ASCIIString(img, &convert.Options{
		Ratio:       1.0,
		FixedWidth:  125,
		FixedHeight: 50,
		FitScreen:   false,
		Colored:     true,
		Reversed:    false,
	})
	asciiLines := strings.Split(asciiArt, "\n")

	spacing := "    "

	fmt.Print(asciiLines[0] + "\n")
	fmt.Print(asciiLines[1] + spacing + "Name: " + album.Name.String + "\n")
	fmt.Print(asciiLines[2] + spacing + "Artist: " + album.Artist.String + "\n")
	fmt.Print(asciiLines[3] + "\n")
	fmt.Print(asciiLines[4] + spacing + "Tracks:\n")
	var trackLine = 5
	tracks, err := app.DB.GetTracksByAlbumId(ctx, pgtype.Text{String: album.ID, Valid: true})
	check(err)
	for i, track := range tracks {
		var trackInfo internal.ExifInfo
		if err := json.Unmarshal(track.Info, &trackInfo); err != nil {
			log.Fatalf("failed to parse track info: %v", err)
		}

		var seconds pgtype.Float8
		seconds, _ = track.TotalDuration.Float64Value()

		var trackNum string
		switch v := trackInfo.Track.(type) {
		case string:
			trackNum = v
		case float64:
			trackNum = fmt.Sprintf("%.0f", v)
		case int:
			trackNum = fmt.Sprintf("%d", v)
		default:
			trackNum = fmt.Sprintf("%d", i+1)
		}

		trackText := fmt.Sprintf("%s. %s (%d:%02d)",
			trackNum,
			trackInfo.Title,
			int(seconds.Float64/60),
			int(seconds.Float64)%60)

		if trackInfo.AudioBitrate != "" {
			trackText += fmt.Sprintf(" [%s]", trackInfo.AudioBitrate)
		}
		if trackLine < len(asciiLines) {
			fmt.Print(asciiLines[trackLine] + spacing + trackText + "\n")
		} else {
			fmt.Print(strings.Repeat(" ", len(asciiLines[0])) + spacing + trackText + "\n")
		}
		trackLine++
	}

	for i := trackLine; i < len(asciiLines); i++ {
		fmt.Print(asciiLines[i] + "\n")
	}

}
