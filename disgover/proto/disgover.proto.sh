
# Install [protoc](https://github.com/google/protobuf/releases) compiler.
#
# `go get -u github.com/golang/protobuf/protoc-gen-go`          # `protoc-gen-go` plugin
# `go get -u google.golang.org/grpc`                            # `grpc` package
# `go get -u golang.org/x/net/context`                          # `context` package
# `go get -u github.com/golang/protobuf/proto`                  # `protobuf` package
# `protoc --go_out=plugins=grpc:. proto/blockchain.proto`       # Build 
#

protoc --go_out=plugins=grpc:. *.proto
