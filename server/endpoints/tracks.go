package endpoints

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strafe/db"
	"strafe/internal"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/valyala/fastjson"
)

type Track struct {
	ID                          string    `json:"id"`
	Cover                       string    `json:"cover"`
	Info                        TrackInfo `json:"info"`
	SavedVocalFolderPath        string    `json:"saved_vocal_folder_path"`
	SavedInstrumentalFolderPath string    `json:"saved_instrumental_folder_path"`
}

type TrackInfo struct {
	Title                string      `json:"title"`
	Artist               string      `json:"artist"`
	Album                string      `json:"album"`
	Length               float64     `json:"length"`
	Genre                string      `json:"genre"`
	VocalWaveform        interface{} `json:"vocal_waveform"`
	InstrumentalWaveform interface{} `json:"instrumental_waveform"`
	Tempo                float64     `json:"tempo"`
	Instrumental         bool        `json:"instrumental"`
	Key                  string      `json:"key"`
}

type GetRandomTrackRequest struct {
	AnonID string `json:"anonId"`
}

func GetRandomTrack(w http.ResponseWriter, r *http.Request) {
	app := r.Context().Value(internal.APP_CONTEXT_KEY).(internal.AppCtx)
	var body GetRandomTrackRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		if err.Error() == "EOF" {
			internal.WriteError(w, internal.MissingJSONBody(err))
			return
		}
		internal.WriteError(w, internal.MalformedJSONBody(err))
		return
	}
	var anonId = pgtype.Text{String: body.AnonID, Valid: true}
	var track db.Track
	var err error
	track, err = app.DB.GetRandomUnlistenedTrack(r.Context(), anonId)
	no_rows := errors.Is(err, sql.ErrNoRows)
	if err != nil && (!no_rows) {
		internal.ServerError(w, err)
		return
	}
	if no_rows {
		tx, err := app.Conn.Begin(r.Context())
		if err != nil {
			internal.ServerError(w, err)
			return
		}
		defer tx.Rollback(r.Context())
		qtx := app.DB.WithTx(tx)
		if err := qtx.DeleteListeningHistoryByAnonID(r.Context(), anonId); err != nil {
			internal.ServerError(w, err)
			return
		}
		track, err = qtx.GetRandomTrack(r.Context())
		if err != nil {
			internal.ServerError(w, err)
			return
		}
		if err := tx.Commit(r.Context()); err != nil {
			internal.ServerError(w, err)
			return
		}
	}
	err = app.DB.RecordListeningHistory(
		r.Context(),
		db.RecordListeningHistoryParams{
			TrackID:    pgtype.Text{String: track.ID, Valid: true},
			AnonID:     anonId,
			ListenedAt: pgtype.Timestamptz{Time: time.Now(), InfinityModifier: pgtype.Infinity},
		},
	)
	if err != nil {
		internal.ServerError(w, err)
		return
	}
	response := processTrackInfo(w, track, app, r)
	if response != nil {
		json.NewEncoder(w).Encode(response)
	}
}

func GetTrack(w http.ResponseWriter, r *http.Request) {
	var trackId = chi.URLParam(r, "trackId")
	app := r.Context().Value(internal.APP_CONTEXT_KEY).(internal.AppCtx)
	track, err := app.DB.GetTrackByID(r.Context(), trackId)
	if err != nil {
		internal.ServerError(w, err)
		return
	}
	response := processTrackInfo(w, track, app, r)
	if response != nil {
		json.NewEncoder(w).Encode(response)
	}
}

func processTrackInfo(w http.ResponseWriter, track db.Track, app internal.AppCtx, r *http.Request) *Track {
	var response Track
	album, err := app.DB.GetAlbumById(r.Context(), track.AlbumID.String)
	if err != nil {
		internal.ServerError(w, err)
		return nil
	}
	response.Cover = fmt.Sprintf("%s/cover.jpg", album.Name.String)
	response.ID = track.ID

	response.SavedVocalFolderPath = track.VocalFolderPath.String
	response.SavedInstrumentalFolderPath = track.InstrumentalFolderPath.String

	trackInfo, err := fastjson.ParseBytes(track.Info)
	if err != nil {
		internal.ServerError(w, err)
		return nil
	}

	getString := func(key string) string {
		val := trackInfo.Get(key)
		if val == nil {
			return ""
		}
		str, err := strconv.Unquote(string(val.MarshalTo(nil)))
		if err != nil {
			internal.ServerError(w, err)
			return ""
		}
		return str
	}

	length, err := track.TotalDuration.Float64Value()
	if err != nil {
		internal.ServerError(w, err)
		return nil
	}

	tempo, err := track.Tempo.Float64Value()
	if err != nil {
		internal.ServerError(w, err)
		return nil
	}

	response.Info = TrackInfo{
		Artist:       getString("Artist"),
		Album:        getString("Album"),
		Genre:        getString("Genre"),
		Title:        getString("Title"),
		Length:       length.Float64,
		Tempo:        tempo.Float64,
		Key:          track.Key.String,
		Instrumental: track.Instrumental.Bool,
	}

	var instrumentalWf []int32
	if err = internal.DecompressJSON(track.InstrumentalWaveform, &instrumentalWf); err != nil {
		internal.ServerError(w, err)
		return nil
	}
	response.Info.InstrumentalWaveform = instrumentalWf
	if !track.Instrumental.Bool {
		var vocalWf []int32
		if err = internal.DecompressJSON(track.VocalWaveform, &vocalWf); err != nil {
			internal.ServerError(w, err)
			return nil
		}
		response.Info.VocalWaveform = vocalWf
	}

	return &response
}
