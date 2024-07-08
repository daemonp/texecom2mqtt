# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=texecom2mqtt
BINARY_UNIX=$(BINARY_NAME)_unix

all: test build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/texecom2mqtt

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/texecom2mqtt
	./$(BINARY_NAME)

deps:
	$(GOGET) github.com/eclipse/paho.mqtt.golang
	$(GOGET) gopkg.in/yaml.v2
	$(GOGET) golang.org/x/text/transform
	$(GOGET) golang.org/x/text/unicode/norm

# Cross compilation
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v ./cmd/texecom2mqtt

docker-build:
	docker build -t texecom2mqtt .

.PHONY: all build test clean run deps build-linux docker-build
