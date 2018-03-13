init:
	@echo Installing dependencies necessary for "make coverage" command ...
	go get github.com/axw/gocov/gocov
	go get -u gopkg.in/matm/v1/gocov-html

build:
	@echo Installing CARAVELA in the local Go environment ...
	go install -v -gcflags "-N -l" .

test:
	@echo Testing CARAVELA ...
	go test .

test-coverage:
	@echo Testing and coverage report generation ...
	gocov test github.com/strabox/caravela/node/guid | gocov-html > coverage.html

docker-build:
	@echo Building CARAVELA's Docker container ...
	docker build --build-arg GOOS=${OS} --rm -t "strabox/caravela:latest" .
	
docker-upload:
	@echo Building CARAVELA's Docker container and uploading to DockerHub ...
	docker build --build-arg GOOS=${OS} --rm -t "strabox/caravela:latest" .
	docker push "strabox/caravela:latest"

debug:
	@echo Debugging Makefile ...
	${LOL} = "0"
	@echo ${LOL}