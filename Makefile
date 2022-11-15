clean:
	del bin /Q /F /S

env:
	set GOOS=linux
	set GOARCH=amd64

build: env build-events build-matches

build-events: env
	go build -o ./bin/events ./cmd/events

build-matches: env
	go build -o ./bin/matches ./cmd/matches