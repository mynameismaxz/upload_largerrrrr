run:
	rm -rf chunks/* && \
	go run http/main.go

build:
	GOOS=linux GOARCH=amd64 go build -x -o upload_api http/main.go