upload:
	docker build --rm -t "strabox/caravela:latest" .
	docker push "strabox/caravela:latest"

build:
	docker build --rm -t "strabox/caravela:latest" .

test:
	go test .

cov:
	gocov test github.com/armon/go-chord | gocov-html > /tmp/coverage.html
	open /tmp/coverage.html