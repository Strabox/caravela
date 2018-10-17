# Caravela: Cloud @ Edge [![Build Status](https://travis-ci.com/Strabox/caravela.svg?token=8iyx88Q98Rgp5aaUbkKN&branch=master)](https://travis-ci.com/Strabox/caravela) [![License: GPL v3](https://img.shields.io/badge/License-GPL%20v3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0) ![](https://img.shields.io/badge/awesome-%E2%9C%93-ff69b4.svg?style=flat-square)

## Table of Contents

- [What is Caravela ?](https://github.com/Strabox/caravela#what-is-caravela-)
    - [Advantages over Docker Swarm](https://github.com/Strabox/caravela#advantages-over-docker-swarm-)
- [Getting Started](https://github.com/Strabox/caravela#getting-started)
    - [Install - Go Toolchain](https://github.com/Strabox/caravela#install---go-toolchain)
    - [Install - From Source](https://github.com/Strabox/caravela#install---from-source)
    - [Create - Create/Bootstrapping a Caravela instance](https://github.com/Strabox/caravela#create---createbootstrapping-a-caravela-instance)
    - [Join - Join a Caravela instance](https://github.com/Strabox/caravela#join---join-a-caravela-instance)
    - [Run - Deploy containers](https://github.com/Strabox/caravela#run---deploy-containers)
    - [Stop - Stop containers](https://github.com/Strabox/caravela#stop---stop-containers)
    - [List - List container deployed](https://github.com/Strabox/caravela#list---list-containers-deployed)
- [Contributing](https://github.com/Strabox/caravela#contributing)
- [License](https://github.com/Strabox/caravela#license)

## What is Caravela ?

Caravela is a prototype for a **Docker container orchestrator platform**, inspired by the Docker Swarm and Kubernetes orchestrators,
but with a fully decentralized architecture, resource discovery and scheduling algorithms, that are scalable in order to be deployed in a Edge Computing environment.

This work was developed in the context of [my Masters Degree Thesis](https://github.com/Strabox/caravela-thesis) in Computer Science and Engineering @ [Instituto Superior Técnico - University of Lisbon.](https://tecnico.ulisboa.pt/en/)

The main goal of the work was to test resource constrained scheduling requests evaluating its performance compared
with centralized versions like Docker Swarm. In order to test its performance and scalability with thousands of nodes,
[Caravela-Sim](https://github.com/Strabox/caravela-sim) was developed.

### Advantages over Docker Swarm

- Caravela has a P2P architecture (on top of the Chord P2P overlay). All the algorithms that run on top it are fully
 decentralized. Caravela is highly scalable and failure tolerant. It has no Single Point of Failure (SPoF).
- In the thesis we prove that for a network with **1,048,576** nodes providing resources, only Caravela
could continue to operate offering a very good performance at all levels. After a few thousands nodes Docker Swarm
master nodes would crush due to the load of requests.
- On contrary of Swarm user's can request the speed of the node's CPU where the container must be deployed.
- We extended the stack deployment of swarm to offer a request-level scheduling policy. A user can request in 
a stack deployment for a set of container to be deployed in the same node (e.g. due to low latency requirements),
we called this **co-location** policy. User can also require that the containers must be spread over different nodes
(e.g. due to resilience requirements), we called this **spread** policy. These proeprties are orthogonal to the system
level scheduling policies (**binpack** and **spread**) that are supported in Docker Swarm and also in Caravela.

## Getting Started

This project is a standalone middleware to orchestrate Docker containers, but it is highly inspired in the
Docker, Swarm and Kubernetes projects so the APIs and the CLI mimic a small. but fundamental, subset of commands/features from
that platforms. So it is easily migrate container's deployment for a Caravela instance.

### Install - Go Toolchain

Using the go tool chain it is possible to easily install Caravela in Go's local environment:

`go get -u github.com/Strabox/caravela`

### Install - From Source

The Makefile in the project's root makes all the necessary actions to easily compile, develop, test and deploy.

The only *prerequisite* to use the basic functionality of makefile is a properly configured golang
environment/toolchain in the machine.
The golang version used in project: **1.9.4**.

There are two ways to build the project using the Makefile:
- Installing it in the golang environment
    1. `make install`
- Building the project in the Makefile directory (creating the executable in the directory)
    1. `make build`

### Create - Create/Bootstrapping a Caravela instance

A Caravela instance can be bootstrapped with a single node. The command necessary is the following one where
<machine_ip> is the local ip address of the machine.

**Important:** This command should be issued to run as daemon because it will run all the Caravela's node operations!!

`caravela create <machine_ip>`

### Join - Join a Caravela instance

To enter and start supplying/use resources, the following command gets the job done. It is only needed to replace
 `<caravela_machine_ip>` with the IP address of a machine that is already participating in a Caravela instance.

**Important:** This command should be issued to run as daemon because it will run all the Caravela's node operations!!

`caravela join <caravela_machine_ip>`

### Run - Deploy containers

To deploy a container in Caravela that requires: a fast CPU (FastCPU: CPUClass = 1, SlowCPU: CPUClass = 0),
2 cores, 256MB of RAM and the port mapping 8080:80 the following command can be issued. The `<container_image>` parameter
should be replaced by a container's image allocated in DockerHub, e.g. `strabox/caravela:latest`.

`caravela run -cpuClass 1 -cpus 2 -memory 256 -p 8080:80 <container_image>`

### Stop - Stop containers

To stop containers it is only necessary to replace `<containerID_1>` for the container's ID. The container's IDs
can be obtained with the command described in next section.

`caravela container stop <containerID_1> <containerID_N>`

### List - List containers deployed

To list all the container deployed in Caravela the following command outputs all of the information about each one.

`caravela container ps`

## Contributing

Use the GitHub [issues tracker](https://github.com/Strabox/caravela/issues) for exposing doubts, recommendations and questions.

## License

This project is licensed under the GPL V3 License - see the LICENSE.md file for details.
