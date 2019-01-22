GOPATH?=~/go
GOBIN=$(GOPATH)/bin
APP=s3autoindex

OK_COLOR=\033[32;01m
NO_COLOR=\033[0m
BOLD=\033[1m

CROSSPLATFORMS=\
	linux/amd64 \
	darwin/amd64

build:
	@echo "$(OK_COLOR)->$(NO_COLOR) Building $(BOLD)$(APP)$(NO_COLOR)"
	@echo "$(OK_COLOR)==>$(NO_COLOR) Installing dependencies"
	go get -v -d ./...
	@echo "$(OK_COLOR)==>$(NO_COLOR) Compiling"
	go install -v ./...

run: build
	@echo "$(OK_COLOR)==>$(NO_COLOR) Running"
	$(GOBIN)/$(APP) -b=127.0.0.1:8000 -bucket=gs://sentryio-storybook -proxy

test:
	go test -v ./...

clean:
	rm -rf $(GOBIN)/*
	rm -rf dist/

docker:
	docker build --rm -t $(APP) .

.PHONY: build run test clean docker
