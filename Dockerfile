FROM golang:1.8-alpine

RUN apk -U add musl-dev gcc make git

RUN mkdir -p /go/src/s3autoindex
WORKDIR /go/src/s3autoindex

COPY . /go/src/s3autoindex

RUN set -x \
    && go get -v -d ./... \
    && go build -a -installsuffix cgo -ldflags "-linkmode external -extldflags \"-static\"" -v -o bin/s3autoindex ./...

FROM scratch
COPY --from=0 /go/src/s3autoindex/bin/s3autoindex s3autoindex
EXPOSE 8000
ENTRYPOINT ["/s3autoindex"]
