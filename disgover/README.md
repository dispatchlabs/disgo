<img src="https://dispatchlabs.io/wp-content/themes/ccprototypev5/images/dispatchlabs-logo.png" width="250">
&nbsp;

![Go Version 1.9.2](http://b.repl.ca/v1/Go_Version-1.9.2-brightgreen.png)

<a name="overview"></a>
### Overview

##### Dispatch KDHT based node discovery engine
dosgover is a distributed node discovery mechanism that enables locating any 
entity (server, worker, drone, actor) based on node address.

The intent is to be able to `PING` and `FIND-NODE` through some transport protocol (currently using grpc). It is not meant for data storage/distribution mechanism. (That will be another package)

Nodes use a Kademlia Hash Table (KDHT) for the following:

 - store contact information for other nodes
 - provide a list of contacts to new nodes joining the network
 - find specific nodes on the network
 - functions as a gateway to outside local network


### Download

`go get github.com/dispatchlabs/disgo/disgover`  
or  
`git clone http://github.com/dispatchlabs/disgo/disgover`


<a name="wiki"></a>
### Wiki Documentation
For more technical details on how disgover works, please visit the [Wiki](https://github.com/dispatchlabs/disgo/disgover/wiki)
 - [Design Approach](https://github.com/dispatchlabs/disgo/disgover/wiki#design)
 - [Getting Started Sample](https://github.com/dispatchlabs/disgo/disgover/wiki#sample)
 - [Packages](https://github.com/dispatchlabs/disgo/disgover/wiki#packages)
 
<a name="dependencies"></a>
### Dependencies

[commons](https://github.com/dispatchlabs/disgo/commons) for common types domain types.

<a name="configuration"></a>
### Configuration
The disgover package only relies on the configuration loaded by [commons](https://github.com/dispatchlabs/disgo/commons) 

<a name="protobuf"></a>
##### protobuf ([see common install instructions](https://github.com/dispatchlabs/disgo/wiki#protoc)


<a name="tests"></a>
### Tests
*Tests to be added*

<a name="acknowledgments"></a>
### Acknowledgments
*Add lists of contributors*

<a name="contributing"></a>
### Contributing
*Add link to common CONRIBUTING.md file*

<a name="license"></a>
### License
*Add License data*

