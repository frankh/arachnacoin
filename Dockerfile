FROM golang:1.9-alpine as gobuild
WORKDIR /go/src/github.com/frankh/arachnacoin
RUN apk add -U git gcc libc-dev ca-certificates
RUN go get github.com/mattn/go-sqlite3 golang.org/x/crypto/ed25519
COPY ./ ./

RUN go build -o arachnacoin .

FROM alpine

COPY --from=gobuild /go/src/github.com/frankh/arachnacoin/arachnacoin /arachnacoin

ENTRYPOINT ["/arachnacoin"]
