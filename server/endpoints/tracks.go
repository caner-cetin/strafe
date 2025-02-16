package endpoints

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strafe/db"
	"strafe/internal"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type GetRandomTrackRequest struct {
	AnonID string `json:"anonId"`
}

func GetRandomTrack(w http.ResponseWriter, r *http.Request) {
	app := r.Context().Value(internal.APP_CONTEXT_KEY).(internal.AppCtx)
	var body GetRandomTrackRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		internal.WriteError(w, internal.MalformedJSONBody)
		return
	}
	var anonId = pgtype.Text{String: body.AnonID, Valid: true}
	var track db.Track
	var err error
	track, err = app.DB.GetRandomUnlistenedTrack(r.Context(), anonId)
	no_rows := errors.Is(err, sql.ErrNoRows)
	if err != nil && (!no_rows) {
		internal.WriteServerError(w, err)
		return
	}
	if no_rows {
		tx, err := app.Conn.Begin(r.Context())
		if err != nil {
			internal.WriteServerError(w, err)
			return
		}
		defer tx.Rollback(r.Context())
		qtx := app.DB.WithTx(tx)
		if err := qtx.DeleteListeningHistoryByAnonID(r.Context(), anonId); err != nil {
			internal.WriteServerError(w, err)
			return
		}
		track, err = qtx.GetRandomTrack(r.Context())
		if err != nil {
			internal.WriteServerError(w, err)
			return
		}
		if err := tx.Commit(r.Context()); err != nil {
			internal.WriteServerError(w, err)
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
		internal.WriteServerError(w, err)
		return
	}
	json.NewEncoder(w).Encode(track)
}
