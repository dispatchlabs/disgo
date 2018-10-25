<img src="https://dispatchlabs.io/wp-content/themes/ccprototypev5/images/dispatchlabs-logo.png" width="250">
 
&nbsp;

![Go Version 1.10.3](http://b.repl.ca/v1/Go_Version-1.10.3-brightgreen.png)
[![Build status](https://ci.appveyor.com/api/projects/status/9hil01is3aflgg6l?svg=true)](https://ci.appveyor.com/project/DispatchLabs/disgo-26m8t)

<a name="overview"></a>
### Overview

Disgo is the main client / node for the Dispatch platform. It initializes all of the services. The services are from the other packages in this repository.  The approach is to have several smaller "building block" modules that are usable components of common blockchain architectures. Our intention is to facilitate a buildable blockchain so that can be used individually so that other projects don't have to start from scratch.
### Docker Option
Install Docker https://www.docker.com/get-started

Open Termianl
```
git clone https://github.com/arsen3d/disgo.git
cd disgo
docker-compose up
```

### Prerequisite
```
go version
go version go1.10.4 darwin/amd64
```
If version lower, download go1.10.4 from 
https://golang.org/dl/
### Download

With Go installed, enter either of the following commands into your terminal:

`go get github.com/dispatchlabs/disgo`  
or  
`git clone http://github.com/dispatchlabs/disgo` (into your GOPATH)

If you have yet to install Go, visit our [tutorial](https://github.com/dispatchlabs/samples/tree/master/golang-setup) or download straight from the [Golang website.](https://golang.org/dl/)
<a name="running"></a>
### How to run the Disgo package

Disgo is a full node, so you should be able to run it right out of the box:

simply run the following commands in your terminal

```
cd $GOPATH/src/github.com/dispatchlabs/disgo
go run main.go
```

For instructions on running Disgo in Docker, please visit the [Wiki Page](https://github.com/dispatchlabs/disgo/wiki#docker)

<a name="using"></a>
### Dancing with Disgo
To dance with disgo either use our [Java SDK](https://github.com/dispatchlabs/java-sdk), [mobile wallet](https://github.com/dispatchlabs/mobile-wallet), or [ScanDis](https://github.com/dispatchlabs/scandis)

<a name="wiki"></a>
### Wiki Documentation
For more technical details on how disgo works, please visit the [Wiki](https://github.com/dispatchlabs/disgo/wiki). 

 - [Development](https://github.com/dispatchlabs/disgo/wiki#development)
 - [Design Approach](https://github.com/dispatchlabs/disgo/wiki#design-approach) 
 - [Getting Started With Disgo](https://github.com/dispatchlabs/disgo/wiki#getting-started-with-disgo)
 - [Packages](https://github.com/dispatchlabs/disgo/wiki#packages)

<a name="dependencies"></a>
### Dependencies

to get all the dependencies simply run `go get ./...` from disgo directory

<a name="configuration"></a>
### Configuration
The disgo package relies on the configuration loaded by [commons](https://github.com/dispatchlabs/disgo/tree/master/commons) 

<a name="protobuf"></a>
##### protobuf ([see common install instructions](https://github.com/dispatchlabs/disgo#-develop))

<a name="tests"></a>
### Tests
We have multiple unit test throughout disgo, go provides a test framework that is easy to use. Simply go into any directory with a file ending in _test.go and call `go test`

<a name="acknowledgments"></a>
### Acknowledgments
*Add lists of contributors*

<a name="contributing"></a>
### Contributing
*Add link to common CONRIBUTING.md file*

<a name="license"></a>
### License
*Add License data*
