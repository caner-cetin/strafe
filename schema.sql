-- public.tracks definition

-- Drop table

-- DROP TABLE public.tracks;

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
	cover text NULL,
	artist text NULL,
	CONSTRAINT albums_pkey PRIMARY KEY (id)
);