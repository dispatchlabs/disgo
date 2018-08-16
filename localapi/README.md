<img src="https://dispatchlabs.io/wp-content/themes/ccprototypev5/images/dispatchlabs-logo.png" width="250">
 
&nbsp;

![Go Version 1.9.2](http://b.repl.ca/v1/Go_Version-1.9.2-brightgreen.png)

<a name="overview"></a>
### Overview

The dapos package provides the services for using **Delegated Asynchronous Proof-of-Stake (DAPoS)** consensus for transactions.

DAPoS is a new consensus algorithm developed by the Dispatch team for use in the Dispatch protocol. It aims to maximize parallelizable transaction throughput. DAPoS maximizes scalability of transaction throuput by minimizing the Delegates codependency. Once transaction information is evenly distributed between Delegates, each Delegate autonomously and deterministically accepts the Transaction into their chain and reports the validity of the Transaction. Work done by delegates is redundant for Byzantine security of the network. DAPoS Delegates are elected by Stakeholders based on stake-weighted voting, and gossip to one another about which Transaction they have received from External Actors using ECDSA signatures. Once a Delegate receives 2/3 of Delegate signatures for a given Transaction within maximum Lag Threshold, the Transaction is accepted and added to that Delegate Ledger. The validity of the Transaction is reported back to Bookkeepers, so Delegates can be evaluated and held accountable by Stakeholders.   

For more details on DAPoS, please refer to [Introduction to DAPoS document](https://github.com/dispatchlabs/TechnicalDocs/blob/master/Introduction%20to%20DAPoS.pdf). 

### Download

`go get github.com/dispatchlabs/disgo/dapos`  
or  
`git clone http://github.com/dispatchlabs/disgo/dapos`


<a name="wiki"></a>
### Wiki Documentation
For more technical details on how dapos works, please visit the [Wiki](https://github.com/dispatchlabs/disgo/dapos/wiki)

 - [Concepts in DAPoS](https://github.com/dispatchlabs/disgo/dapos/wiki#genesis)
 - [Design Approach](https://github.com/dispatchlabs/disgo/dapos/wiki#design)
 - [Getting Started Sample](https://github.com/dispatchlabs/disgo/dapos/wiki#sample)
 - [Packages](https://github.com/dispatchlabs/disgo/dapos/wiki#packages)


<a name="dependencies"></a>
### Dependencies
Because dapos relies on sending transactions to other delegates, it is necessary to also have the [disgover](https://github.com/dispatchlabs/disgo/disgover) package for network discovery.
[commons](https://github.com/dispatchlabs/disgo/commons) for common types domain types.
NOTE: if you `go get ./..` then dependencies are getting pulled automatically

<a name="configuration"></a>
### Configuration
The dapos package only relies on the configuration loaded by [commons](https://github.com/dispatchlabs/disgo/commons) 

<a name="protobuf"></a>
##### protobuf ([see common install instructions](https://github.com/dispatchlabs/disgo/wiki#protoc))

<a name="usage"></a>
### How to use the dapos package
See the [wiki page](https://github.com/dispatchlabs/disgo/dapos/wiki#getting-started-sample) for links to full examples on running bare-bones dapos

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
