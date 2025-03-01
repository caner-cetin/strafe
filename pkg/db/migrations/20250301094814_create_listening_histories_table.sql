-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.listening_histories (
	track_id text NULL,
	anon_id text NULL,
	listened_at timestamptz NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE public.listening_histories;
-- +goose StatementEnd
