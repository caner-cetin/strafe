-- public.tracks definition

-- Drop table

-- DROP TABLE public.tracks;

CREATE TABLE public.tracks (
	id text NOT NULL,
	vocal_folder_path text NULL,
	instrumental_folder_path text NULL,
	album_id text NULL,
	total_duration numeric NULL,
	vocal_waveform BYTEA NULL,
	instrumental_waveform BYTEA NULL,
	info jsonb NULL,
	instrumental bool NULL,
	tempo numeric NULL,
	"key" text NULL,
	CONSTRAINT tracks_pkey PRIMARY KEY (id)
);
CREATE INDEX idx_tracks_artist ON public.tracks ("(info ->> 'Artist'::text)");
CREATE INDEX idx_tracks_genre ON public.tracks ("(info ->> 'Genre'::text)");
CREATE INDEX idx_tracks_title ON public.tracks ("(info ->> 'Title'::text)");

-- public.listening_histories definition

-- Drop table

-- DROP TABLE public.listening_histories;

CREATE TABLE public.listening_histories (
	track_id text NULL,
	anon_id text NULL,
	listened_at timestamptz NULL
);

-- public.albums definition

-- Drop table

-- DROP TABLE public.albums;

CREATE TABLE public.albums (
	id text NOT NULL,
	"name" text NULL,
	cover_extension text NULL,
	CONSTRAINT albums_pkey PRIMARY KEY (id)
);