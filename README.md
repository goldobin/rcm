# RCM - Redis Cluster Manager

RCM is a simple utility to manage [Redis 3.x](https://github.com/antirez/redis) clusters for development purposes. 
Right now it supports only `localhost` as a host of the cluster. But probably Docker containers will be also supported.

Influenced by [ccm (Cassandra Cluster Manager)](https://github.com/pcmanus/ccm)

# Prerequisites

- Redis 3.x. The `redis-server` and `redis-cli` should be on the `${PATH}`. On OS X it is as simple as 
`brew install redis` 

# Installation

## Using Go

```bash
brew install go

export GOPATH=~/rcm-workspace

go get github.com/codegangsta/cli
go get github.com/fatih/color
go get bitbucket.org/goldobin/rcm

go install rcm
```

The RCM utility will appear in your `${GOPATH}/bin`. Just do not forget to put `${GOPATH}/bin` in your `${PATH}` 
variable by executing `export PATH=${PATH}:${GOPATH}/bin` to be able to run RCM without specifying full path to it.

# Usage

```bash
rcm create test
rcm start
rcm distribute-slots
```

That sequence of commands will create a new Redis 3.x cluster with default parameters: 6 nodes and 1 replica per master. 
So the final cluster will consist of 3 master and 3 slave nodes. Also by default the cluster will be configured to not 
use persistence feature. 

Run `rcm info test` to check if cluster is OK.
  
And of course you can start regular `redis-cli` session:

```bash
rcm cli test
```

or just execute single Redis command

```bash
rcm cli test set x y
```
 
To get the complete list of commands and options please use `rcm help`   





