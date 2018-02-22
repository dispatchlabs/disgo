FROM golang:alpine as builder

RUN apk update && apk add git

RUN go get github.com/dispatchlabs/disgo
RUN cd /go/src/github.com/dispatchlabs/disgo \
  && go build

# -----------------------
FROM alpine:latest

COPY --from=builder /go/src/github.com/dispatchlabs/disgo/disgo /usr/bin/disgo

ENTRYPOINT ["disgo"]
