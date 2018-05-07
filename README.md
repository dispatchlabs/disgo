<img src="https://dispatchlabs.io/wp-content/themes/ccprototypev5/images/dispatchlabs-logo.png" width="250">
 
&nbsp;

![Go Version 1.9.2](http://b.repl.ca/v1/Go_Version-1.9.2-brightgreen.png)

<a name="overview"></a>
### Overview

Disgo is the main client / node for the Dispatch platform. It initializes all of the services. The services are from the other packages in this repository.  The approach is to have several smaller "building block" modules that are usable components of common blockchain architectures. Our intention is to facilitate a buildable blockchain so that can be used individually so that other projects don't have to start from scratch.

### Download

`go get github.com/dispatchlabs/disgo`  
or  
`git clone http://github.com/dispatchlabs/disgo`


<a name="wiki"></a>
### Wiki Documentation
For more technical details on how disgo works, please visit the [Wiki](https://github.com/dispatchlabs/disgo/wiki). 

 - [Development](https://github.com/dispatchlabs/disgo/wiki#development)
 - [Design Approach](https://github.com/dispatchlabs/disgo/wiki#design-approach) 
 - [Getting Started With Disgo](https://github.com/dispatchlabs/disgo/wiki#getting-started-with-disgo)
 - [Packages](https://github.com/dispatchlabs/disgo/wiki#packages)

<a name="dependencies"></a>
### Dependencies
uses all of the packages in the repo.

 - [commons](https://github.com/dispatchlabs/disgo_commons) for common domain types.
 - [disgover](https://github.com/dispatchlabs/disgover) for node discovery.
 - [dapos](https://github.com/dispatchlabs/dapos) for consensus.

<a name="configuration"></a>
### Configuration
The disgo package relies on the configuration loaded by [commons](https://github.com/dispatchlabs/disgo_commons) 

<a name="protobuf"></a>
##### protobuf ([see common install instructions](https://github.com/dispatchlabs/disgo#-develop))

<a name="usage"></a>
### How to run the disgo package

Disgo is a full node, so you should be able to run it right out of the box:

```
cd ~/go/src/dispatchlabs/disgo
go run main.go
```

For instructions on running Disgo in Docker, please visit the [Wiki Page](https://github.com/dispatchlabs/disgo/wiki#docker)

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
