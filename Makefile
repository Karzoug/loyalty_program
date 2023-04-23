include .env

DATABASE_URI=postgres://${PG_USER}:${PG_PASSWORD}@localhost:5432/${PG_DB}?sslmode=disable
GOPATH=$(shell go env GOPATH)

build:
	cd cmd/gophermart && go build -buildvcs=false -o gophermart

clean:
	rm -f cmd/gophermart/gophermart

run:
	RUN_ADDRESS='localhost:${RUN_PORT}' ACCRUAL_SYSTEM_ADDRESS='http://localhost:${ACCRUAL_PORT}' DATABASE_URI='$(DATABASE_URI)' SECRET_KEY='${SECRET_KEY}' DEBUG=TRUE ./cmd/gophermart/gophermart

lint:
	golangci-lint run ./...

migrate-up:
	migrate -path migrations -database $(DATABASE_URI) up

migrate-down:
	migrate -path migrations -database $(DATABASE_URI) down

install-dependency:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin v1.52.2
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest