GOCMD=go
GOBUILD=$(GOCMD) build
GOINSTALL=$(GOCMD) install
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=caravela
BINARY_NAME_LINUX=$(BINARY_NAME)_linux
BINARY_NAME_WIN=$(BINARY_NAME)_win


all: test build

build:
	@echo Building...
	$(GOBUILD) -o $(BINARY_NAME).exe -v

build-linux:
	@echo Building for linux...
	GOOS=linux $(GOBUILD) -o $(BINARY_NAME_LINUX).exe -v

clean:
	@echo Cleaning project...
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME_LINUX)
	rm -f $(BINARY_NAME_WIN)

install:
	@echo Installing CARAVELA in the local Go environment...
	$(GOINSTALL) -o $(BINARY_NAME).exe -v -gcflags "-N -l" .

test:
	@echo Testing...
	$(GOTEST) -v ./...

test-cov:
	@echo Testing and coverage report generation...
	gocov test ./... | gocov-html > coverage.html

dep-graph:
	@echo Generating package import dependency graph...
	godepgraph -s -p github.com/docker github.com/strabox/caravela | dot -Tpng -o importsGraph.png

docker-build:
	@echo Building CARAVELA's Docker container ...
	docker build --build-arg GOOS=${OS} --rm -t "strabox/caravela:latest" .
	
docker-upload:
	@echo Building CARAVELA's Docker container and uploading to DockerHub...
	docker build --build-arg GOOS=${OS} --rm -t "strabox/caravela:latest" .
	docker push "strabox/caravela:latest"

install-aux-tools:
	@echo Installing dependencies necessary for coverage report and import dependencies...
	$(GOGET) github.com/axw/gocov/gocov
	$(GOGET) -u gopkg.in/matm/v1/gocov-html
	$(GOGET) github.com/kisielk/godepgraph