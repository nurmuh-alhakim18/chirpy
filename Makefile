export DB_URL="postgres://postgres:postgres@localhost:5432/chirpy"

migrate-create:
	@ goose -dir sql/schema create $(name) sql

migrate-up:
	@ goose postgres ${DB_URL} -dir sql/schema up

migrate-down:
	@ goose postgres ${DB_URL} -dir sql/schema down