// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: query.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const deleteListeningHistoryByAnonID = `-- name: DeleteListeningHistoryByAnonID :exec
DELETE FROM listening_histories
WHERE anon_id = $1
`

// Delete all listening history for a given anonymous user
func (q *Queries) DeleteListeningHistoryByAnonID(ctx context.Context, anonID pgtype.Text) error {
	_, err := q.db.Exec(ctx, deleteListeningHistoryByAnonID, anonID)
	return err
}

const getAlbumById = `-- name: GetAlbumById :one
SELECT a.id, a.name, a.cover_extension
FROM albums a
WHERE a.id = $1
`

func (q *Queries) GetAlbumById(ctx context.Context, id string) (Album, error) {
	row := q.db.QueryRow(ctx, getAlbumById, id)
	var i Album
	err := row.Scan(&i.ID, &i.Name, &i.CoverExtension)
	return i, err
}

const getRandomTrack = `-- name: GetRandomTrack :one
SELECT t.id, t.vocal_folder_path, t.instrumental_folder_path, t.album_id, t.total_duration, t.info, t.instrumental, t.tempo, t.key, t.vocal_waveform, t.instrumental_waveform
FROM tracks t
LEFT JOIN albums a ON a.id = t.album_id
ORDER BY RANDOM()
LIMIT 1
`

// Get a completely random track
func (q *Queries) GetRandomTrack(ctx context.Context) (Track, error) {
	row := q.db.QueryRow(ctx, getRandomTrack)
	var i Track
	err := row.Scan(
		&i.ID,
		&i.VocalFolderPath,
		&i.InstrumentalFolderPath,
		&i.AlbumID,
		&i.TotalDuration,
		&i.Info,
		&i.Instrumental,
		&i.Tempo,
		&i.Key,
		&i.VocalWaveform,
		&i.InstrumentalWaveform,
	)
	return i, err
}

const getRandomUnlistenedTrack = `-- name: GetRandomUnlistenedTrack :one
SELECT t.id, t.vocal_folder_path, t.instrumental_folder_path, t.album_id, t.total_duration, t.info, t.instrumental, t.tempo, t.key, t.vocal_waveform, t.instrumental_waveform
FROM tracks t
LEFT JOIN albums a ON a.id = t.album_id
LEFT JOIN listening_histories lh ON t.id = lh.track_id AND lh.anon_id = $1
WHERE lh.track_id IS NULL
ORDER BY RANDOM()
LIMIT 1
`

// Get a random track that hasn't been listened to by the given anonymous user
func (q *Queries) GetRandomUnlistenedTrack(ctx context.Context, anonID pgtype.Text) (Track, error) {
	row := q.db.QueryRow(ctx, getRandomUnlistenedTrack, anonID)
	var i Track
	err := row.Scan(
		&i.ID,
		&i.VocalFolderPath,
		&i.InstrumentalFolderPath,
		&i.AlbumID,
		&i.TotalDuration,
		&i.Info,
		&i.Instrumental,
		&i.Tempo,
		&i.Key,
		&i.VocalWaveform,
		&i.InstrumentalWaveform,
	)
	return i, err
}

const getTrackBasicByID = `-- name: GetTrackBasicByID :one
SELECT 
    id,
    vocal_folder_path,
    instrumental_folder_path,
    album_id,
    total_duration,
    info,
    instrumental,
    tempo,
    "key"
FROM tracks
WHERE id = $1
`

type GetTrackBasicByIDRow struct {
	ID                     string
	VocalFolderPath        pgtype.Text
	InstrumentalFolderPath pgtype.Text
	AlbumID                pgtype.Text
	TotalDuration          pgtype.Numeric
	Info                   []byte
	Instrumental           pgtype.Bool
	Tempo                  pgtype.Numeric
	Key                    pgtype.Text
}

// Gets basic track information by ID
func (q *Queries) GetTrackBasicByID(ctx context.Context, id string) (GetTrackBasicByIDRow, error) {
	row := q.db.QueryRow(ctx, getTrackBasicByID, id)
	var i GetTrackBasicByIDRow
	err := row.Scan(
		&i.ID,
		&i.VocalFolderPath,
		&i.InstrumentalFolderPath,
		&i.AlbumID,
		&i.TotalDuration,
		&i.Info,
		&i.Instrumental,
		&i.Tempo,
		&i.Key,
	)
	return i, err
}

const getTrackByID = `-- name: GetTrackByID :one
SELECT t.id, t.vocal_folder_path, t.instrumental_folder_path, t.album_id, t.total_duration, t.info, t.instrumental, t.tempo, t.key, t.vocal_waveform, t.instrumental_waveform
FROM tracks t
WHERE t.id = $1
`

// Get track by ID (all columns, use GetTrackBasicByID if waveforms are not needed)
func (q *Queries) GetTrackByID(ctx context.Context, id string) (Track, error) {
	row := q.db.QueryRow(ctx, getTrackByID, id)
	var i Track
	err := row.Scan(
		&i.ID,
		&i.VocalFolderPath,
		&i.InstrumentalFolderPath,
		&i.AlbumID,
		&i.TotalDuration,
		&i.Info,
		&i.Instrumental,
		&i.Tempo,
		&i.Key,
		&i.VocalWaveform,
		&i.InstrumentalWaveform,
	)
	return i, err
}

const getTrackCount = `-- name: GetTrackCount :one
SELECT COUNT(*) FROM tracks
`

// Gets total number of tracks
func (q *Queries) GetTrackCount(ctx context.Context) (int64, error) {
	row := q.db.QueryRow(ctx, getTrackCount)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const getTrackWaveforms = `-- name: GetTrackWaveforms :one
SELECT 
    vocal_waveform,
    instrumental_waveform
FROM tracks
WHERE id = $1
`

type GetTrackWaveformsRow struct {
	VocalWaveform        []byte
	InstrumentalWaveform []byte
}

// Gets only the waveform data for a specific track
func (q *Queries) GetTrackWaveforms(ctx context.Context, id string) (GetTrackWaveformsRow, error) {
	row := q.db.QueryRow(ctx, getTrackWaveforms, id)
	var i GetTrackWaveformsRow
	err := row.Scan(&i.VocalWaveform, &i.InstrumentalWaveform)
	return i, err
}

const getTracksBasic = `-- name: GetTracksBasic :many
SELECT 
    id,
    vocal_folder_path,
    instrumental_folder_path,
    album_id,
    total_duration,
    info,
    instrumental,
    tempo,
    "key"
FROM tracks
`

type GetTracksBasicRow struct {
	ID                     string
	VocalFolderPath        pgtype.Text
	InstrumentalFolderPath pgtype.Text
	AlbumID                pgtype.Text
	TotalDuration          pgtype.Numeric
	Info                   []byte
	Instrumental           pgtype.Bool
	Tempo                  pgtype.Numeric
	Key                    pgtype.Text
}

// Gets all track information except waveforms
func (q *Queries) GetTracksBasic(ctx context.Context) ([]GetTracksBasicRow, error) {
	rows, err := q.db.Query(ctx, getTracksBasic)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetTracksBasicRow
	for rows.Next() {
		var i GetTracksBasicRow
		if err := rows.Scan(
			&i.ID,
			&i.VocalFolderPath,
			&i.InstrumentalFolderPath,
			&i.AlbumID,
			&i.TotalDuration,
			&i.Info,
			&i.Instrumental,
			&i.Tempo,
			&i.Key,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTracksBasicPaginated = `-- name: GetTracksBasicPaginated :many
SELECT 
    id,
    vocal_folder_path,
    instrumental_folder_path,
    album_id,
    total_duration,
    info,
    instrumental,
    tempo,
    "key"
FROM tracks
LIMIT $1
OFFSET $2
`

type GetTracksBasicPaginatedParams struct {
	Limit  int32
	Offset int32
}

type GetTracksBasicPaginatedRow struct {
	ID                     string
	VocalFolderPath        pgtype.Text
	InstrumentalFolderPath pgtype.Text
	AlbumID                pgtype.Text
	TotalDuration          pgtype.Numeric
	Info                   []byte
	Instrumental           pgtype.Bool
	Tempo                  pgtype.Numeric
	Key                    pgtype.Text
}

// Gets basic track information with pagination
func (q *Queries) GetTracksBasicPaginated(ctx context.Context, arg GetTracksBasicPaginatedParams) ([]GetTracksBasicPaginatedRow, error) {
	rows, err := q.db.Query(ctx, getTracksBasicPaginated, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetTracksBasicPaginatedRow
	for rows.Next() {
		var i GetTracksBasicPaginatedRow
		if err := rows.Scan(
			&i.ID,
			&i.VocalFolderPath,
			&i.InstrumentalFolderPath,
			&i.AlbumID,
			&i.TotalDuration,
			&i.Info,
			&i.Instrumental,
			&i.Tempo,
			&i.Key,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTracksByArtist = `-- name: GetTracksByArtist :many
SELECT 
    id,
    vocal_folder_path,
    instrumental_folder_path,
    album_id,
    total_duration,
    info,
    instrumental,
    tempo,
    "key"
FROM tracks
WHERE info->>'Artist' = $1
`

type GetTracksByArtistRow struct {
	ID                     string
	VocalFolderPath        pgtype.Text
	InstrumentalFolderPath pgtype.Text
	AlbumID                pgtype.Text
	TotalDuration          pgtype.Numeric
	Info                   []byte
	Instrumental           pgtype.Bool
	Tempo                  pgtype.Numeric
	Key                    pgtype.Text
}

// Gets basic track information filtered by artist
func (q *Queries) GetTracksByArtist(ctx context.Context, info []byte) ([]GetTracksByArtistRow, error) {
	rows, err := q.db.Query(ctx, getTracksByArtist, info)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetTracksByArtistRow
	for rows.Next() {
		var i GetTracksByArtistRow
		if err := rows.Scan(
			&i.ID,
			&i.VocalFolderPath,
			&i.InstrumentalFolderPath,
			&i.AlbumID,
			&i.TotalDuration,
			&i.Info,
			&i.Instrumental,
			&i.Tempo,
			&i.Key,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTracksByGenre = `-- name: GetTracksByGenre :many
SELECT 
    id,
    vocal_folder_path,
    instrumental_folder_path,
    album_id,
    total_duration,
    info,
    instrumental,
    tempo,
    "key"
FROM tracks
WHERE info->>'Genre' = $1
`

type GetTracksByGenreRow struct {
	ID                     string
	VocalFolderPath        pgtype.Text
	InstrumentalFolderPath pgtype.Text
	AlbumID                pgtype.Text
	TotalDuration          pgtype.Numeric
	Info                   []byte
	Instrumental           pgtype.Bool
	Tempo                  pgtype.Numeric
	Key                    pgtype.Text
}

// Gets basic track information filtered by genre
func (q *Queries) GetTracksByGenre(ctx context.Context, info []byte) ([]GetTracksByGenreRow, error) {
	rows, err := q.db.Query(ctx, getTracksByGenre, info)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetTracksByGenreRow
	for rows.Next() {
		var i GetTracksByGenreRow
		if err := rows.Scan(
			&i.ID,
			&i.VocalFolderPath,
			&i.InstrumentalFolderPath,
			&i.AlbumID,
			&i.TotalDuration,
			&i.Info,
			&i.Instrumental,
			&i.Tempo,
			&i.Key,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const recordListeningHistory = `-- name: RecordListeningHistory :exec
INSERT INTO listening_histories (track_id, anon_id, listened_at)
VALUES ($1, $2, $3)
`

type RecordListeningHistoryParams struct {
	TrackID    pgtype.Text
	AnonID     pgtype.Text
	ListenedAt pgtype.Timestamptz
}

func (q *Queries) RecordListeningHistory(ctx context.Context, arg RecordListeningHistoryParams) error {
	_, err := q.db.Exec(ctx, recordListeningHistory, arg.TrackID, arg.AnonID, arg.ListenedAt)
	return err
}

const searchTracks = `-- name: SearchTracks :many
SELECT 
    id,
    vocal_folder_path,
    instrumental_folder_path,
    album_id,
    total_duration,
    info,
    instrumental,
    tempo,
    "key"
FROM tracks
WHERE 
    info->>'Title' ILIKE '%' || $1 || '%' OR
    info->>'Artist' ILIKE '%' || $1 || '%' OR
    info->>'Genre' ILIKE '%' || $1 || '%'
LIMIT $2 OFFSET $3
`

type SearchTracksParams struct {
	Column1 pgtype.Text
	Limit   int32
	Offset  int32
}

type SearchTracksRow struct {
	ID                     string
	VocalFolderPath        pgtype.Text
	InstrumentalFolderPath pgtype.Text
	AlbumID                pgtype.Text
	TotalDuration          pgtype.Numeric
	Info                   []byte
	Instrumental           pgtype.Bool
	Tempo                  pgtype.Numeric
	Key                    pgtype.Text
}

// Searches tracks by title, artist, or genre
func (q *Queries) SearchTracks(ctx context.Context, arg SearchTracksParams) ([]SearchTracksRow, error) {
	rows, err := q.db.Query(ctx, searchTracks, arg.Column1, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SearchTracksRow
	for rows.Next() {
		var i SearchTracksRow
		if err := rows.Scan(
			&i.ID,
			&i.VocalFolderPath,
			&i.InstrumentalFolderPath,
			&i.AlbumID,
			&i.TotalDuration,
			&i.Info,
			&i.Instrumental,
			&i.Tempo,
			&i.Key,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
