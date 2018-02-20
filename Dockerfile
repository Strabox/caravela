FROM golang:1.7.1-alpine

ARG GOOS="linux"

COPY . /go/src/github.com/strabox/caravela
WORKDIR /go/src/github.com/strabox/caravela

RUN set -ex
RUN apk add --no-cache --virtual .build-deps git

RUN go get github.com/gorilla/mux
RUN go get github.com/strabox/go-chord
RUN GOOS=$GOOS go install -v -gcflags "-N -l" github.com/strabox/caravela

RUN apk del .build-deps


EXPOSE 8000
EXPOSE 8001

VOLUME $HOME/.caravela

ENTRYPOINT ["caravela"]