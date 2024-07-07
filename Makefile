.PHONY: build run clean

build:
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/main main.go

run: build
	sls deploy --stage dev

clean:
	sls remove --stage dev
