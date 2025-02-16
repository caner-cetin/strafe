// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Album struct {
	ID             string
	Name           pgtype.Text
	CoverExtension pgtype.Text
}

type ListeningHistory struct {
	TrackID    pgtype.Text
	AnonID     pgtype.Text
	ListenedAt pgtype.Timestamptz
}

type Track struct {
	ID                     string
	VocalFolderPath        pgtype.Text
	InstrumentalFolderPath pgtype.Text
	AlbumID                pgtype.Text
	TotalDuration          pgtype.Numeric
	VocalWaveform          []byte
	InstrumentalWaveform   []byte
	Info                   []byte
	Instrumental           pgtype.Bool
	Tempo                  pgtype.Numeric
	Key                    pgtype.Text
}
