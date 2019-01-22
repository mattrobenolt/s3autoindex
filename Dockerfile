FROM golang:1.11-alpine

RUN apk -U add musl-dev gcc make git curl

RUN mkdir -p /usr/src/s3autoindex
WORKDIR /usr/src/s3autoindex

COPY go.mod go.sum /usr/src/s3autoindex/
RUN go mod download

COPY . /usr/src/s3autoindex

RUN set -x \
    && curl https://mkcert.org/generate/ | grep -v '#' > cacert.pem \
    && go build -a -installsuffix cgo -ldflags "-linkmode external -extldflags \"-static\"" -v -o bin/s3autoindex ./...

FROM scratch
COPY --from=0 /usr/src/s3autoindex/bin/s3autoindex s3autoindex
COPY --from=0 /usr/src/s3autoindex/cacert.pem /etc/ssl/certs/
EXPOSE 8000
ENTRYPOINT ["/s3autoindex"]
