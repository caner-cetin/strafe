-- name: GetTracksByArtist :many
-- Gets basic track information filtered by artist
SELECT id,
    vocal_folder_path,
    instrumental_folder_path,
    album_id,
    total_duration,
    info,
    instrumental,
    tempo,
    "key"
FROM tracks
WHERE info->>'Artist' = $1;
-- name: GetTracksByAlbumId :many
-- Gets basic track information filtered by album ID, sorted by track list
SELECT id,
    vocal_folder_path,
    instrumental_folder_path,
    album_id,
    total_duration,
    info,
    instrumental,
    tempo,
    "key"
FROM tracks
WHERE album_id = $1
ORDER BY info->>'Track';
-- name: GetTracksByGenre :many
-- Gets basic track information filtered by genre
SELECT id,
    vocal_folder_path,
    instrumental_folder_path,
    album_id,
    total_duration,
    info,
    instrumental,
    tempo,
    "key"
FROM tracks
WHERE info->>'Genre' = $1;
-- name: GetTrackCount :one
-- Gets total number of tracks
SELECT COUNT(*)
FROM tracks;
-- name: SearchTracks :many
-- Searches tracks by title, artist, or genre
SELECT id,
    vocal_folder_path,
    instrumental_folder_path,
    album_id,
    total_duration,
    info,
    instrumental,
    tempo,
    "key"
FROM tracks
WHERE info->>'Title' ILIKE '%' || $1 || '%'
    OR info->>'Artist' ILIKE '%' || $1 || '%'
    OR info->>'Genre' ILIKE '%' || $1 || '%'
LIMIT $2 OFFSET $3;
-- name: GetRandomUnlistenedTrack :one
-- Get a random track that hasn't been listened to by the given anonymous user
SELECT t.*
FROM tracks t
    LEFT JOIN albums a ON a.id = t.album_id
    LEFT JOIN listening_histories lh ON t.id = lh.track_id
    AND lh.anon_id = $1
WHERE lh.track_id IS NULL
ORDER BY RANDOM()
LIMIT 1;
-- name: DeleteListeningHistoryByAnonID :exec
-- Delete all listening history for a given anonymous user
DELETE FROM listening_histories
WHERE anon_id = $1;
-- name: GetRandomTrack :one
-- Get a completely random track
SELECT t.*
FROM tracks t
    LEFT JOIN albums a ON a.id = t.album_id
ORDER BY RANDOM()
LIMIT 1;
-- name: RecordListeningHistory :exec
INSERT INTO listening_histories (track_id, anon_id, listened_at)
VALUES ($1, $2, $3);
-- name: GetAlbumById :one
SELECT a.*
FROM albums a
WHERE a.id = $1;
-- name: GetAlbumByName :one
SELECT a.*
FROM albums a
WHERE a.name = $1;
-- name: GetAlbumByArtist :one
SELECT a.*
FROM albums a
WHERE a.artist = $1;
-- name: GetAlbumByNameAndArtist :one
SELECT a.*
FROM albums a
WHERE a.name = $1
    AND a.artist = $2;
-- name: GetAlbumIDByName :one
SELECT a.id
FROM albums a
WHERE a.name = $1;
-- name: GetAlbumIDByNameAndArtist :one
SELECT a.id
FROM albums a
WHERE a.name = $1
    AND a.artist = $2;
-- name: InsertAlbum :one
-- returns id
INSERT INTO public.albums (id, "name", cover, artist)
VALUES($1, $2, $3, $4)
RETURNING id;
-- name: GetTrackByID :one
-- Get track by ID (all columns, use GetTrackBasicByID if waveforms are not needed)
SELECT t.*
FROM tracks t
WHERE t.id = $1;
-- name: GetAlbumCoverByID :one
-- Get album cover by the album ID
SELECT a.cover
FROM albums a
WHERE a.id = $1;
-- name: InsertTrack :exec
INSERT INTO public.tracks (
        id,
        vocal_folder_path,
        instrumental_folder_path,
        album_id,
        total_duration,
        info,
        instrumental,
        tempo,
        "key",
        vocal_waveform,
        instrumental_waveform,
        album_name
    )
VALUES(
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11,
        $12
    );