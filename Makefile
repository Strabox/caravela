##################### CARAVELA's MAKEFILE #########################
GOCMD=go

######### Builtin GO tools #########
GOBUILD=$(GOCMD) build
GOINSTALL=$(GOCMD) install
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet

GOGET=$(GOCMD) get

######### External GO tools #########
GOLINT=golint
GOCOV=gocov
GOCOVHTML=gocov-html
GODEPGRAPH=godepgraph

############ Output Files ###########
EXE=.exe
BINARY_NAME=caravela$(EXE)
BINARY_NAME_LINUX=$(BINARY_NAME)_linux$(EXE)
BINARY_NAME_WIN=$(BINARY_NAME)_win$(EXE)

############################## COMMANDS ############################

all: test build

build:
	@echo Building for the current machine settings...
	$(GOBUILD) -o $(BINARY_NAME) -v

build-linux:
	@echo Building for linux...
	env GOOS=linux $(GOBUILD) -o $(BINARY_NAME) -v

build-windows:
	@echo Building for windows...
	env GOOS=windows $(GOBUILD) -o $(BINARY_NAME) -v

clean:
	@echo Cleaning project...
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME_LINUX)
	rm -f $(BINARY_NAME_WIN)

install:
	@echo Installing CARAVELA in the local GO environment...
	$(GOINSTALL) -v -gcflags "-N -l" .

test:
	@echo Testing...
	$(GOTEST) -v ./...

test-cov:
	@echo Testing and coverage report generation...
	$(GOCOV) test ./... | $(GOCOVHTML) > coverage.html

test-verify:
	$(MAKE) test
	@echo Running vet tool to static analyze the code
	$(GOVET) -v ./...
	@echo Running lint tool to static analyze the code style

dep-graph:
	@echo Generating package import dependency graph...
	$(GODEPGRAPH) -s -p github.com/docker github.com/strabox/caravela | dot -Tpng -o importsGraph.png

docker-build:
	@echo Building Docker container...
	docker build --build-arg exec_file=$(BINARY_NAME) --rm -t strabox/caravela:latest .

docker-upload:
	@echo Building Docker container and uploading to DockerHub...
	docker build --build-arg exec_file=$(BINARY_NAME) --rm -t strabox/caravela:latest .
	docker push strabox/caravela:latest

install-external-tools:
	@echo Installing external tools...
	@echo Installing lint - Code style analyzer (from google)
	$(GOGET) github.com/golang/lint
	@echo Installing gocov - Code coverage generator
	$(GOGET) github.com/axw/gocov/gocov
	@echo Installing gocov-html - Code coverage html generator
	$(GOGET) -u gopkg.in/matm/v1/gocov-html
	@echo Installing godepgraph - Code package dependency graph generator
	$(GOGET) github.com/kisielk/godepgraph

