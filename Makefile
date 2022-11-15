clean:
	rm bin/*

env:
	export GOOS=linux
	export GOARCH=amd64
	export CGO_ENABLED=0

build: clean env build-events build-matches

build-events: env
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/events ./cmd/events

build-matches: env
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/matches ./cmd/matches