
FROM golang:latest

RUN apt-get update && apt-get install -y jq curl
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

WORKDIR /go/src/github.com/arsen3d/disgo

#COPY . /go/src/github.com/arsen3d/disgo
#COPY . /go/src/github.com/arsen3d/vendor/github.com/dispatchlabs/disgo

#RUN dep ensure
#RUN go install ./cmd/...
RUN go get github.com/dispatchlabs/disgo
EXPOSE 1975:1975

#ENTRYPOINT /go/bin/disgo
#ENTRYPOINT go run main.go
ENTRYPOINT go run $GOPATH/src/github.com/dispatchlabs/disgo/main.go