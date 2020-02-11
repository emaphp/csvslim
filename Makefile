.PHONY: install build get

get:
	go mod tidy

build:
	go build

install: get
	go install
