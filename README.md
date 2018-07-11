# Caravela: A Cloud @ Edge [![Build Status](https://travis-ci.com/Strabox/caravela.svg?token=8iyx88Q98Rgp5aaUbkKN&branch=master)](https://travis-ci.com/Strabox/caravela) [![License: GPL v3](https://img.shields.io/badge/License-GPL%20v3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0) ![](https://img.shields.io/badge/awesome-%E2%9C%93-ff69b4.svg?style=flat-square)

## Table of Contents

- [What is CARAVELA ?](https://github.com/Strabox/CARAVELA-A-Cloud-at-Edge#what-is-caravela-)
- [Getting Started](https://github.com/Strabox/CARAVELA-A-Cloud-at-Edge#getting-started)
- [Contributing](https://github.com/Strabox/CARAVELA-A-Cloud-at-Edge#contributing)
- [License](https://github.com/Strabox/CARAVELA-A-Cloud-at-Edge#license)

## What is CARAVELA ?

CARAVELA is a prototype container cloud management/orchestrator platform, inspired by the Docker Swarm and Kubernetes orchestrators,
but with a fully decentralized, hence scalable architecture in order to be deployed in a Edge Computing environment.

This work is being developed in the context of my Masters Degree Thesis in Computer Science and Engineering @ Instituto Superior Técnico - Lisbon.

## Getting Started

This project is a standalone middleware to orchestrate Docker containers but it is highly inspired in the
Docker, Swarm and Kubernetes projects so the APIs and the CLI try to mimic a small subset of commands/features from
that platforms. So it is easily migrate container's deployment for a CARAVELA instance.

### Compile From Source

The Makefile in the project's root makes all the necessary actions to easily compile, develop, test and deploy.

The only *prerequisite* to use the basic functionality of makefile is a properly configured golang
environment/toolchain in the machine.
The golang version used in project: **1.9.4**.

There are two ways to build the project using the Makefile:
- Installing it in the golang environment
    1. `make install`
- Building the project in the Makefile directory (creating the executable in the directory)
    1. `make build`

### Create/Bootstrapping a CARAVELA instance

A CARAVELA instance can be bootstrapped with a single node. The command necessary is the following one where
<machine_ip> is the local ip address of the machine.

`caravela create <machine_ip>`

### Join a CARAVELA instance

To enter and start supplying/use resources, the following command gets the job done. It is only needed to replace
`<machine_ip>` with the local IP address of the machine and `<caravela_machine_ip>` with the IP address of a machine that is
already participating in a CARAVELA instance.

`caravela join <machine_ip> <caravela_machine_ip>`

### Run, Stop Containers and more

#### Run

`caravela run -cpus 2 -ram 256 <image>`

#### Stop

## Contributing

TODO

## License

This project is licensed under the GPL V3 License - see the LICENSE.md file for details.
