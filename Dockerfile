FROM golang:1.9-alpine

ARG GOOS="linux"
ARG GOARCH="amd64"

RUN echo [DOCKERFILE] Building CARAVELA for OS=$GOOS and ARCH=$GOARCH

COPY . /go/src/github.com/strabox/caravela
WORKDIR /go/src/github.com/strabox/caravela

RUN set -ex
RUN apk add --no-cache --virtual .build-deps git

RUN GOOS=$GOOS GOARCH=$GOARCH go install -v -gcflags "-N -l" github.com/strabox/caravela

RUN apk del .build-deps


EXPOSE 8000	# Expose the Overlay Port to outside
EXPOSE 8001	# Expose the CARAVELAs Port to outside

VOLUME $HOME/.caravela

ENTRYPOINT ["caravela"]