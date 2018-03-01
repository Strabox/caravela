FROM golang:1.9-alpine

ARG GOOS="linux"

COPY . /go/src/github.com/strabox/caravela
WORKDIR /go/src/github.com/strabox/caravela

RUN set -ex
RUN apk add --no-cache --virtual .build-deps git

RUN go get github.com/gorilla/mux
RUN go get github.com/bluele/go-chord
RUN go get github.com/docker/docker/client
RUN GOOS=$GOOS go install -v -gcflags "-N -l" github.com/strabox/caravela

RUN apk del .build-deps


EXPOSE 8000
EXPOSE 8001

VOLUME $HOME/.caravela

ENTRYPOINT ["caravela"]