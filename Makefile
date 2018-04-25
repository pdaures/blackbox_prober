# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=blackbox_prober
BINARY_LINUX=$(BINARY_NAME)_linux_amd64
    
all: test build
build: 
	$(GOBUILD) -o $(BINARY_NAME) -v
test: 
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_LINUX)

# Cross compilation
build-linux:
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) -a -tags netgo -ldflags '-w' -o $(BINARY_LINUX) .

docker: build-linux
	docker build .
