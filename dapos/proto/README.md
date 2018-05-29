
<img src="https://dispatchlabs.io/wp-content/themes/ccprototypev5/images/dispatchlabs-logo.png" width="250">

# DAPoS/proto

<a name="overview"></a>
### Overview

##### Proto Library

The  current transport protocol on Dispatch is GRPC. However, we are allowing to scale it to any transport protocol (e.g. JSON RPC, among other RPC mechanisms). Therefore, we created Proto3 as an decoupled package that defines the interfaces which need to be configured for the transport protocol to be compatible with Dispatchlabs consensus.

To see the interfaces, refer to:

[DAPoS.proto](https://github.com/dispatchlabs/disgo/dapos/blob/master/proto/dapos.proto) 




<a name="configuration"></a>
### Configuration
To be able to generate the go language bindings:

`protoc --go_out=plugins=grpc:. *.proto`