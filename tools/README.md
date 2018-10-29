# tools

## Generate a key and address to use for the Genesis Account
Before you create your cluster, you may want to generate a new Genesis Account to work with.

#### Generate a new account to use for the Genesis Account.
This generation does not store the key information anywhere, so it is important that you copy the values and put them in the place of your choosing.  The private key will be needed to submit transactions using the SDKs.

from the directory with the tools executable, run:

> `./tools makePrivateKey`

You will get something that looks like:

```
Genesis key & address:
***********************************************************************************************************************************************************
Public Key:	0497def3d49ed1996189ac4504695627011466265052e6afda61d06dbf59b52781f3f8bfbb64e12dd59f9134a034cade46bb8d54cdfc0e91ae4433e568bd40dfd4
Private Key:    651ac5dc86c10dbf99b508dcf4bffc74ef45c565cb00619de94edc53722fc498
Address:	eb0ecc73844b3a3bf8f35d23bb279cbd6b021ddc
***********************************************************************************************************************************************************
```

The public key can be derived from the private key, but is provided for convenience (though you don't need to use it with the SDK)

The Address is the address you can use when generating your cluster as the genesis Address (the one that dispurses all the tokens)

The corresponding private key is needed to sign transactions for the Genesis Account.


## Create a new local cluster

The cluster is created in your "$HOME" directory on your machine under the directory disgo_cluster.  If you just use the defaults it will create a cluster with one seed and four delegate nodes.

from the tools directory:

> `./tools newLocalCluster 5 eb0ecc73844b3a3bf8f35d23bb279cbd6b021ddc 100000000`

 - parameter 1 is the directive for creating a new local cluster
 - parameter 2 is for the number of delegates you want to have in your cluster (I typically use 4 or 5 for local)
 	- if you do not specify anything, a cluster with 4 delegates and  
 - parameter 3 is the address for your Genesis Account.  This will be used when generating the genesis account file that you can find in the disgo/config/genesis_account.json file
 - paramater 4 is the balance that you are initializing your cluster with.  This value is immutable, so make sure you put enough in there to play around with.

 
 ### Note about this executable:
 I built this tools project for the purpose of speeding up testing the development environment.  It has a lot of other features, but they are still tied to a specific genesis account.  It won't be difficult to detach that and provide the features like: very simple bulk token transfers, contract deployment and execution.  We will set that up in the coming weeks.  Our plan is to get this to be a real CLI and move it into the public repository. We will make a lot of incremental updates to this in the near future.
 
## Running the cluster

Right now this part is a little bit manual.

Once you have generated your Genesis Account and created the local cluster, you want to run it right!

How I do it (and I'm using a mac, but you can translate this to any OS)

Open up a grid of command promp windows (for a cluster with 5 delegates, you need 6, so you have one for the seed as well.

> cd $HOME/disgo_cluster/seed-0  
> ./disgo  //starts the seed node  
> cd $HOME/disgo_cluster/delegate-0  
> ./disgo  
> cd $HOME/disgo_cluster/delegate-1  
> ./disgo  
> cd $HOME/disgo_cluster/delegate-2  
> ./disgo  
> cd $HOME/disgo_cluster/delegate-3  
> ./disgo  
> cd $HOME/disgo_cluster/delegate-4  
> ./disgo  

When the delegates are started, it is easiest to see the port numbers for each delegate in the seed output (it doesn't log much else) or better yet, once you have started the cluster you can request the seed for the delegates that are available.

If you use curl or postman, run a GET against the seed:

```http://localhost:1975/v1/delegates```


Delegates are set up with port numbers starting at 3502
That will get a cluster of 5 delegates running.  Now you can use any of the SDKs to submit transactions to the cluster using the list of delegate ports (all using localhost)

