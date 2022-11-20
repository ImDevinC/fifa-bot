version := 0.0.6

clean:
	rm -rf bin/*

build: clean build-events build-matches

build-events:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-X 'main.release=${version}'" -o ./bin/events ./cmd/events

build-matches:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-X 'main.release=${version}'" -o ./bin/matches ./cmd/matches

dist: build
	zip -r ./bin/events.zip ./bin/events ./cmd/ ./pkg/
	zip -r ./bin/matches.zip ./bin/matches ./cmd/ ./pkg/