migrate-init:
	migrate create -ext sql -dir database/migration -seq init_schema

migrate-up:
	migrate -path database/migration -database "postgresql://postgres:postgres@localhost:5432/crypto-alert?sslmode=disable" -verbose  up 

migrate-down:
	migrate -path database/migration -database "postgresql://postgres:postgres@localhost:5432/crypto-alert?sslmode=disable" -verbose  down 