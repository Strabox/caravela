init:
	@echo Installing dependencies necessary for 'make coverage' command...
	go get github.com/axw/gocov/gocov
	go get -u gopkg.in/matm/v1/gocov-html

build:
	@echo Installing CARAVELA in the local Go environment...
	go install .

test:
	go test .

coverage:
	@echo Tests and coverage report generation...
	gocov test github.com/strabox/caravela/node/guid | gocov-html > coverage.html
	
docker-upload:
	docker build --rm -t "strabox/caravela:latest" .
	docker push "strabox/caravela:latest"

docker-build:
	docker build --rm -t "strabox/caravela:latest" .