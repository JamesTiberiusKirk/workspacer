install: 
	go install ./cmd/workspacer

build:
	go build -o ./bin/workspacer ./cmd/workspacer/

test:
	go test -race -v ./...

