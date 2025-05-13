VERSION ?= development

clean:
	rm -rf bin/*

build: clean 
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-X 'main.release=${VERSION}'" -o ./bin/events ./cmd/server.go
