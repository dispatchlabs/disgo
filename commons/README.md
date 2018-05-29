<img src="https://dispatchlabs.io/wp-content/themes/ccprototypev5/images/dispatchlabs-logo.png" width="250">

![Go Version 1.9.2](http://b.repl.ca/v1/Go_Version-1.9.2-brightgreen.png)

<a name="overview"></a>
### Overview

commons contains the common code that multiple components depend on.   All projects in the repositry depend on commons. At the moment the criteria for living in commons is:

 - Structures, Interfaces, and Services that are used throughout the system.
   - Particularly any dependent files that would result in circular dependencies otherwise.
 - Files with common utility functions

### Download

`go get github.com/dispatchlabs/commons`  
or  
`git clone http://github.com/dispatchlabs/commons`


<a name="wiki"></a>
### Wiki Documentation
Technical details for commons is available on the [Wiki](https://github.com/dispatchlabs/disgo_commons/wiki) page. Here is a shortcut list of interesting topics:

 - [How to use the common config structure](https://github.com/dispatchlabs/commons/wiki#configuration)
 - [What is the IService interface](https://github.com/dispatchlabs/commons/wiki#iservice-interface)
 - [Details of the crypo package](https://github.com/dispatchlabs/commons/wiki#crypto)

<a name="dependencies"></a>
### Dependencies

The significant dependency for commons is the C crypto libraries we are using.  At the moment, it is necessary to have gcc installed to use the crypto features.  We are in the process of creating the platform specific binaries so that additional installs are not necessary.

<a name="configuration"></a>
### Configuration
commons contains the configuration struct that a client using the system needs for setting up system properties. 

Any custom node bootstrap implementation should load the appropriate properties into this structure for the components that are used.  For a concrete example, see how it is done in our [disgo node]()

The details of the configuration setup can be view [here in the wiki](https://github.com/dispatchlabs/commons/wiki#configuration)

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
