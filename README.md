<img src="https://www.dispatchlabs.io/wp-content/uploads/2018/12/Dispatch_Logo.png" width="250">
 
&nbsp;

![Go Version 1.10.3](http://b.repl.ca/v1/Go_Version-1.10.3-brightgreen.png)
[![Build status](https://ci.appveyor.com/api/projects/status/9hil01is3aflgg6l?svg=true)](https://ci.appveyor.com/project/DispatchLabs/disgo-26m8t)

<a name="overview"></a>
### Overview

Disgo (*dispatch* + *go*) is the first client implementation of the Dispatch protocol. Dispatch enables the Zero-Knowledge Analytics of distributed data without comprimising data ownership, privacy, or security.

### Come talk with us!
If you have any questions or just want to get to know us better, come say hi in our [discord channel](https://Dispatchlabs.io/discord) (https://Dispatchlabs.io/discord)

### Download

With Go installed, enter either of the following commands into your terminal:

`go get github.com/dispatchlabs/disgo`  
or  
`git clone http://github.com/dispatchlabs/disgo` (into your GOPATH)

If you have yet to install Go, visit our [tutorial](https://github.com/dispatchlabs/samples/tree/master/golang-setup) or download straight from the [Golang website.](https://golang.org/dl/)
<a name="running"></a>
### How to run a Disgo node

Disgo is a full dispatch node, and you can run it right out of the box. Simply run the following commands in your terminal:

```
cd $GOPATH/src/github.com/dispatchlabs/disgo
go get ./...
go run main.go
```
<a name="using"></a>
### Using the protocol (Dancing the Disgo 🕺)
- Non-technical users of the protocol can use [the network scanner](http://scanner.dispatchlabs.io) to interact with the protocol. 

- Developers "dance the disgo" with our decentralized [HTTP API](https://api.dispatchlabs.io), [JavaScript SDK](https://github.com/dispatchlabs/dev-tools), or [Java SDK](https://github.com/dispatchlabs/java-sdk). 

### Configuration
The disgo package relies on the configuration loaded by [commons](https://github.com/dispatchlabs/disgo/tree/master/commons) 

<a name="contributing"></a>
### Contributing
*We would love your help developing the protocol!* It's a big project and we're a small team, so all contributions are encouraged. For more information on how to get a developer environment set up, please check out our [dev-tools](https://github.com/dispatchlabs/dev-tools) repo.

<a name="License"></a>
### License
*The Disgo library is free software: you can redistribute it and/or modify it under the terms of version 3 of the GNU General Public License as published by the Free Software Foundation.*
 
*The Disgo library is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more details.*
