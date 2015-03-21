GOPATH=$(realpath ../../../../)
GOBIN=$(GOPATH)/bin
GOENV=GOPATH=$(GOPATH) GOBIN=$(GOBIN)
GO=$(GOENV) go
GOX=$(GOENV) $(GOBIN)/gox
APPS=\
	s3autoindex

OK_COLOR=\033[32;01m
NO_COLOR=\033[0m
BOLD=\033[1m

build:
	@for app in $(APPS); do \
		echo "$(OK_COLOR)->$(NO_COLOR) Building $(BOLD)$${app}$(NO_COLOR)"; \
		echo "$(OK_COLOR)==>$(NO_COLOR) Installing dependencies"; \
		$(GO) get -v -d ./...; \
		echo "$(OK_COLOR)==>$(NO_COLOR) Compiling"; \
		$(GO) install -v ./bin/$${app}; \
		echo; \
	done;

run: build
	@echo "$(OK_COLOR)==>$(NO_COLOR) Running"
	$(GOBIN)/s3autoindex -b=:8000

test:
	$(GO) test -v ./...

clean:
	rm -rf $(GOBIN)/*
	rm -rf $(GOPATH)/pkg/*
	rm -rf dist/

gox:
	$(GO) get github.com/mitchellh/gox
	@for app in $(APPS); do \
		echo "$(OK_COLOR)->$(NO_COLOR) Cross-compiling $(BOLD)$${app}$(NO_COLOR)"; \
		$(GOX) -output="dist/{{.OS}}_{{.Arch}}/{{.Dir}}" ./bin/$${app}; \
		echo; \
	done;

.PHONY: build run test clean gox
