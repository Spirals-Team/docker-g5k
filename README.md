# docker-g5k
A tool to create a Docker Swarm cluster for Docker Machine on Grid5000 testbed infrastructure.  
It only support creating and deleting nodes in Docker Machine.   

## Requirements
* [Docker](https://www.docker.com/products/overview#/install_the_platform)
* [Docker Machine](https://docs.docker.com/machine/install-machine)
* [Docker Machine Driver for Grid5000](https://github.com/Spirals-Team/docker-machine-driver-g5k)
* [Go tools](https://golang.org/doc/install)

You need a Grid5000 account to use this tool. See [this page](https://www.grid5000.fr/mediawiki/index.php/Grid5000:Get_an_account) to create an account.

## Installation

## Installation from sources
*This procedure was only tested on Ubuntu 16.04.*

To use the Go tools, you need to set your [GOPATH](https://golang.org/doc/code.html#GOPATH) variable environment.

To get the code and compile the binary, run:
```bash
go get -u github.com/Spirals-Team/docker-g5k
```

Then, either put the 'docker-g5k' binary in a directory filled in your PATH environment variable, or run:
```bash
export PATH=$PATH:$GOPATH/bin
```

## How to use

### VPN
You need to be connected to the Grid5000 VPN to create and access your Docker nodes.  
Do not forget to configure your DNS or use OpenVPN DNS auto-configuration.  
Please follow the instructions from the [Grid5000 Wiki](https://www.grid5000.fr/mediawiki/index.php/VPN).

### Command line flags

#### Global flags

|            Option            |                       Description                       |     Default value     |  Required  |
|------------------------------|---------------------------------------------------------|-----------------------|------------|
| `--g5k-username`             | Your Grid5000 account username                          |                       | Yes        |
| `--g5k-password`             | Your Grid5000 account password                          |                       | Yes        |
| `--g5k-site`                 | Site to reserve the resources on                        |                       | Yes        |

#### Cluster creation flags

|            Option            |                       Description                       |     Default value     |  Required  |
|------------------------------|---------------------------------------------------------|-----------------------|------------|
| `--g5k-nb-nodes`             | Number of nodes to allocate                             | 3                     | No         |
| `--g5k-walltime`             | Timelife of the machine                                 | "1:00:00"             | No         |
| `--g5k-ssh-private-key`      | Path of your ssh private key                            | "~/.ssh/id_rsa"       | No         |
| `--g5k-ssh-public-key`       | Path of your ssh public key                             | "< private-key >.pub" | No         |
| `--g5k-image`                | Name of the image to deploy                             | "jessie-x64-min"      | No         |
| `--g5k-resource-properties`  | Resource selection with OAR properties (SQL format)     |                       | No         |
| `--swarm-discovery`          | Discovery service to use with Swarm                     | Generate a new token  | No         |
| `--swarm-image`              | Specify Docker image to use for Swarm                   | "swarm:latest"        | No         |
| `--swarm-strategy`           | Define a default scheduling strategy for Swarm          | "spread"              | No         |
| `--swarm-opt`                | Define arbitrary flags for Swarm master                 |                       | No         |
| `--swarm-join-opt`           | Define arbitrary flags for Swarm join                   |                       | No         |
| `--weave-networking`         | Use Weave for networking (INCOMPLETE)                   | False                 | No         |

#### Cluster deletion flags

|            Option            |                       Description                       |     Default value     |  Required  |
|------------------------------|---------------------------------------------------------|-----------------------|------------|
| `--g5k-job-id`               | Only remove nodes related to the provided job ID        |                       | No         |

### Examples

#### Cluster creation

An example of a 3 nodes Docker Swarm cluster creation:
```bash
docker-g5k --g5k-username user \
--g5k-password ******** \
--g5k-site lille \
create-cluster \
--g5k-ssh-private-key ~/.ssh/g5k-key
```

An example where 3 nodes join an existing Docker Swarm cluster using a discovery token:
```bash
docker-g5k --g5k-username user \
--g5k-password ******** \
--g5k-site lille \
create-cluster \
--swarm-discovery "token://xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" \
--g5k-ssh-private-key ~/.ssh/g5k-key
```

An example of a 16 nodes Docker Swarm cluster creation with resource properties (nodes in cluster `chimint` with more thant 8GB of RAM and at least 4 CPU cores):
```bash
docker-g5k --g5k-username user \
--g5k-password ******** \
--g5k-site lille \
create-cluster \
--g5k-ssh-private-key ~/.ssh/g5k-key \
--g5k-nb-nodes 16 \
--g5k-resource-properties "cluster = 'chimint' and memnode > 8192 and cpucore >= 4"
```

#### Cluster deletion

An example of deleting all provisionned nodes:
```bash
docker-g5k --g5k-username user \
--g5k-password ******** \
--g5k-site lille \
remove-cluster
```

An example of deleting only nodes related to a job ID:
```bash
docker-g5k --g5k-username user \
--g5k-password ******** \
--g5k-site lille \
remove-cluster \
--g5k-job-id 1234567
```

#### Use

After creating a cluster, you should be able to use it with usual Docker Machine commands.  
By example, to list all nodes in the cluster :
```bash
docker-machine ls
```
 
**If you remove a node with Docker Machine 'rm' command, the job will be deleted and ALL nodes related to this job will become unavailable**  
