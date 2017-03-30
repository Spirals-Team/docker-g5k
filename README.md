# docker-g5k
A tool to create a Docker Swarm cluster for Docker Machine on the Grid5000 testbed infrastructure.  

## Requirements
* [Docker](https://www.docker.com/products/overview#/install_the_platform)
* [Docker Machine](https://docs.docker.com/machine/install-machine)
* [Docker Machine Driver for Grid5000 (v1.4.1+)](https://github.com/Spirals-Team/docker-machine-driver-g5k)
* [Go tools (Only for installation from sources)](https://golang.org/doc/install)

You need a Grid5000 account to use this tool. See [this page](https://www.grid5000.fr/mediawiki/index.php/Grid5000:Get_an_account) to create an account.

## Installation

## Installation from GitHub releases
Binary releases for Linux, MacOS and Windows using x86/x86_64 CPU architectures are available in the [releases page](https://github.com/Spirals-Team/docker-g5k/releases).  
You can use the following commands to install or upgrade the tool:
```bash
# download the binary for your OS and CPU architecture :
sudo curl -L -o /usr/local/bin/docker-g5k "<link to release>"

# grant execution rigths for everyone :
sudo chmod 755 /usr/local/bin/docker-g5k
```

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

All flags can be set using environment variables, the name is the same as the flag in uppercase and replacing hyphens by underscores.
For example, the env variable for the flag `--g5k-resource-properties` is `G5K_RESOURCE_PROPERTIES`.

Flags marked with `[]` can be set multiple times and the values will be added.

Flags marked with `{}` support brace expansions (same format as sh/bash shells) to generate combinations.
For example, in the flag  `--g5k-reserve-nodes` you can use it to reserve the same number of nodes on multiple sites : `{lille,nantes}:16`,
or select a range of nodes like in the `--swarm-master` flag : `lille-{0..2}`.

#### Cluster creation flags

|             Option             |                       Description                       |     Default value     |  Required  |
|--------------------------------|---------------------------------------------------------|-----------------------|------------|
| `--g5k-username`               | Your Grid5000 account username                          |                       | Yes        |
| `--g5k-password`               | Your Grid5000 account password                          |                       | Yes        |
| `--g5k-reserve-nodes` [ ],{ }  | Reserve nodes on a site                                 |                       | Yes        |
| `--g5k-walltime`               | Timelife of the machine                                 | "1:00:00"             | No         |
| `--g5k-image`                  | Name of the image to deploy                             | "jessie-x64-min"      | No         |
| `--g5k-resource-properties`    | Resource selection with OAR properties (SQL format)     |                       | No         |
| `--engine-opt` [ ],{ }         | Specify flags to include on the selected node(s) engine |                       | No         |
| `--engine-label` [ ],{ }       | Specify labels for the selected node(s) engine          |                       | No         |
| `--swarm-master` [ ],{ }       | Select node(s) to be promoted to Swarm Master           |                       | No         |
| `--swarm-mode-enable`          | Create a Swarm mode cluster                             |                       | No         |
| `--swarm-standalone-enable`    | Create a Swarm standalone cluster                       |                       | No         |
| `--swarm-standalone-discovery` | Discovery service to use with Swarm                     | Generate a new token  | No         |
| `--swarm-standalone-image`     | Specify Docker image to use for Swarm                   | "swarm:latest"        | No         |
| `--swarm-standalone-strategy`  | Define a default scheduling strategy for Swarm          | "spread"              | No         |
| `--swarm-standalone-opt`       | Define arbitrary global flags for Swarm master          |                       | No         |
| `--swarm-standalone-join-opt`  | Define arbitrary global flags for Swarm join            |                       | No         |
| `--weave-networking`           | Use Weave for networking (Only with Swarm standalone)   |                       | No         |

Engine flags `--engine-*` format is `node-name:key=val` and brace expansions are supported.  
For example, `lille-0:mykey=myval`, `lille-{0..5}:mykey=myval`.  

For `--engine-opt` flag, please refer to [Docker documentation](https://docs.docker.com/engine/reference/commandline/dockerd/) for supported parameters.  
**Test your parameters on a single node before deploying a cluster ! If your flags are incorrect, Docker wont start and you should redeploy the entire cluster !**

#### Cluster deletion flags

|             Option             |                       Description                       |     Default value     |  Required  |
|--------------------------------|---------------------------------------------------------|-----------------------|------------|
| `--g5k-job-id` [ ]             | Only remove nodes related to the provided job ID        | "1:00:00"             | No         |

### Examples

#### Cluster creation

An example of a 16 nodes Docker reservation (Only Docker, Swarm is not configured):
```bash
docker-g5k create-cluster \
--g5k-username "user" \
--g5k-password "********" \
--g5k-reserve-nodes "lille:16" \
```

An example of a 16 nodes Docker reservation with Engine options and labels:
```bash
docker-g5k create-cluster \
--g5k-username "user" \
--g5k-password "********" \
--g5k-reserve-nodes "lille:16" \
--engine-opt "lille-{0..15}:graph=/tmp" \
--engine-label "lille-0:mylabelname1=mylabelvalue1" \
--engine-label "lille-{1..5}:mylabelname2=mylabelvalue2"
```

An example of a 16 nodes Docker reservation using resource properties (nodes with more thant 8GB of RAM and at least 4 CPU cores):
```bash
docker-g5k create-cluster \
--g5k-username "user" \
--g5k-password "********" \
--g5k-reserve-nodes "lille:16" \
--g5k-resource-properties "memnode > 8192 and cpucore >= 4"
```

An example of multi-sites cluster creation:
```bash
docker-g5k create-cluster \
--g5k-username "user" \
--g5k-password "********" \
--g5k-reserve-nodes "lille:16" \
--g5k-reserve-nodes "nantes:8"
```

An example of multi-sites cluster creation using brace expansion:
```bash
docker-g5k create-cluster \
--g5k-username "user" \
--g5k-password "********" \
--g5k-reserve-nodes "{lille,nantes}:16"
```

An example of a 16 nodes Docker Swarm mode cluster creation using the first 3 nodes as Swarm Master:
```bash
docker-g5k create-cluster \
--g5k-username "user" \
--g5k-password "********" \
--g5k-reserve-nodes "lille:16" \
--swarm-mode-enable \
--swarm-master "lille-{0..2}"
```

An example of a 16 nodes Docker Swarm standalone cluster creation using the first node as Swarm Master and Weave Networking:
```bash
docker-g5k create-cluster \
--g5k-username "user" \
--g5k-password "********" \
--g5k-reserve-nodes "lille:16" \
--swarm-standalone-enable \
--swarm-master "lille-0" \
--weave-networking
```

#### Cluster deletion

An example of deleting all provisionned nodes:
```bash
docker-g5k remove-cluster
```

An example of deleting only nodes related to a job ID:
```bash
docker-g5k remove-cluster --g5k-job-id 1234567
```

#### Normal use

After creating a cluster, you should be able to use it with usual Docker Machine commands.  
By example, to list all nodes in the cluster :
```bash
docker-machine ls
```

**If you remove a node with Docker Machine 'rm' command, the job will be deleted and ALL nodes related to this job will become unavailable**  

### Use with Weave networking (Only with Swarm standalone)

First, you need to configure your Docker client to use the Swarm mode (You can get the Swarm master hostname with 'docker-machine ls'):
```bash
eval $(docker-machine env --swarm swarm-master-node-name)
```

Then run a container using Weave networking:
```bash
docker run --net=weave -h foo.weave.local --name foo --dns=172.17.0.1 --dns-search=weave.local. -td your-image:version
```
Your containers can now communicate with each other using theirs short ('foo') or long ('foo.weave.local') name.  
The name used NEED to be the one given in parameter '-h'. The name of the container (parameter '--name') is not used by Weave.