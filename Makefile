clean:
	rm -rf bin/*

build: clean build-events build-matches

build-events:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/events ./cmd/events

build-matches:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/matches ./cmd/matches

dist: build
	zip -j ./bin/events.zip ./bin/events
	zip -j ./bin/matches.zip ./bin/matches