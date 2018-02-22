FROM golang:alpine as builder

RUN apk update && apk add git

ENV GOPATH /go:$GOPATH
ENV PATH /go/bin:$PATH

ADD . /go/src
WORKDIR /go/src

RUN go get github.com/dispatchlabs/disgo
RUN cd /go/src/github.com/dispatchlabs/disgo \
  && echo $(pwd)  \
  && go build \
  && mv ./disgo /go-binaries

# -----------------------
FROM alpine:latest

COPY --from=builder /go-binaries/disgo /go-binaries/disgo

ENTRYPOINT ["/go-binaries/disgo"]
