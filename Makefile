build:
	go install .

test:
	go test .

coverage:
	gocov test github.com/strabox/caravela/node/guid | gocov-html > coverage.html
	
docker-upload:
	docker build --rm -t "strabox/caravela:latest" .
	docker push "strabox/caravela:latest"

docker-build:
	docker build --rm -t "strabox/caravela:latest" .