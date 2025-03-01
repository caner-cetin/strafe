-- +goose Up
-- +goose StatementBegin
CREATE TABLE public.tracks (
	id text NOT NULL,
	vocal_folder_path text NULL,
	instrumental_folder_path text NULL,
	album_id text NULL,
	total_duration numeric NULL,
	info jsonb NULL,
	instrumental bool NULL,
	tempo numeric NULL,
	"key" text NULL,
	vocal_waveform bytea NULL,
	instrumental_waveform bytea NULL,
	album_name text NULL,
	CONSTRAINT tracks_pkey PRIMARY KEY (id)
);
CREATE INDEX idx_tracks_artist ON public.tracks USING btree (((info ->> 'Artist'::text)));
CREATE INDEX idx_tracks_genre ON public.tracks USING btree (((info ->> 'Genre'::text)));
CREATE INDEX idx_tracks_title ON public.tracks USING btree (((info ->> 'Title'::text)));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE public.tracks;
-- +goose StatementEnd
