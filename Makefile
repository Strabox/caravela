init:
	@echo Installing dependencies necessary for coverage report and import dependencies...
	go get github.com/axw/gocov/gocov
	go get -u gopkg.in/matm/v1/gocov-html
	go get github.com/kisielk/godepgraph

build:
	@echo Installing CARAVELA in the local Go environment...
	go install -v -gcflags "-N -l" .

test:
	@echo Testing CARAVELA ...
	go test github.com/strabox/caravela/node/common/guid github.com/strabox/caravela/node/common/resources

test-coverage:
	@echo Testing and coverage report generation...
	gocov test github.com/strabox/caravela/node/guid github.com/strabox/caravela/node/resources | gocov-html > coverage.html

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

debug:
	@echo Debugging Makefile...
	${LOL} = "0"
	@echo ${LOL}