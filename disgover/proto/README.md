<img src="https://dispatchlabs.io/wp-content/themes/ccprototypev5/images/dispatchlabs-logo.png" width="250">

# disgover/proto

<a name="overview"></a>
### Overview

##### Proto Library

The  current transport protocol on Dispatch is GRPC. However, we are allowing to scale it to any transport protocol (e.g. JSON RPC, among other RPC mechanisms). Therefore, we created Proto3 as an decoupled package that defines the interfaces which need to be configured for the transport protocol to be compatible with Didpatchlabs node discovery.

To see the interfaces, refer to:

[disgover.proto](https://github.com/dispatchlabs/disgo/disgover/blob/master/proto/disgover.proto) 




<a name="configuration"></a>
### Configuration
To be able to generate the go language bindings:

`protoc --go_out=plugins=grpc:. *.proto`




