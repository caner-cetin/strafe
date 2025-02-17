-- name: GetTracksBasic :many
-- Gets all track information except waveforms
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
FROM tracks;

-- name: GetTrackWaveforms :one
-- Gets only the waveform data for a specific track
SELECT 
    vocal_waveform,
    instrumental_waveform
FROM tracks
WHERE id = $1;

-- name: GetTracksBasicPaginated :many
-- Gets basic track information with pagination
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
OFFSET $2;

-- name: GetTracksByArtist :many
-- Gets basic track information filtered by artist
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
WHERE info->>'Artist' = $1;

-- name: GetTracksByGenre :many
-- Gets basic track information filtered by genre
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
WHERE info->>'Genre' = $1;

-- name: GetTrackCount :one
-- Gets total number of tracks
SELECT COUNT(*) FROM tracks;

-- name: GetTrackBasicByID :one
-- Gets basic track information by ID
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
WHERE id = $1;

-- name: SearchTracks :many
-- Searches tracks by title, artist, or genre
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
LIMIT $2 OFFSET $3;

-- name: GetRandomUnlistenedTrack :one
-- Get a random track that hasn't been listened to by the given anonymous user
SELECT t.*
FROM tracks t
LEFT JOIN albums a ON a.id = t.album_id
LEFT JOIN listening_histories lh ON t.id = lh.track_id AND lh.anon_id = $1
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

-- name: GetTrackByID :one
-- Get track by ID (all columns, use GetTrackBasicByID if waveforms are not needed)
SELECT t.*
FROM tracks t
WHERE t.id = $1;