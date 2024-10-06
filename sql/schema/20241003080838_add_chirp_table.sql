-- +goose Up
-- +goose StatementBegin
CREATE TABLE chirps (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL,
  body TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,

  CONSTRAINT fk_userchirp FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE chirps;
-- +goose StatementEnd
