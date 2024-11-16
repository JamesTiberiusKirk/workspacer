install: 
	go install ./cmd/workspacer

build:
	go build -o ./bin/workspacer ./cmd/workspacer/main.go

