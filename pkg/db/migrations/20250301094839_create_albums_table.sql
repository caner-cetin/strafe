-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.albums (
	id text NOT NULL,
	"name" text NULL,
	cover text NULL,
	artist text NULL,
	CONSTRAINT albums_pkey PRIMARY KEY (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE public.albums;
-- +goose StatementEnd
