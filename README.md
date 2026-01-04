# Distributed Storage System

## Overview
**Distributed Storage System** is a project aimed at creating a scalable and fault-tolerant file-system.

Serve thousands of big Files.

This system is designed to distribute data across multiple nodes, ensuring high availability and reliability.

## Usage

clone repo via:

```sh
git clone https://github.com/ArpBansal/DistVault.git
cd DistVault
```

Install Dependencies

```sh
go mod tidy
```

```sh
wget https://releases.hashicorp.com/consul/1.20.1/consul_1.20.1_linux_amd64.zip && unzip -o consul_1.20.1_linux_amd64.zip && sudo mv consul /usr/local/bin/ && consul version
```

Run:
```sh
make run
```

## Features
- Will be Optimized to handle large files
- Distributed data storage
- High availability and fault tolerance
- Efficient data retrieval
- Scalable architecture

## System
- GO 1.23(used to develop this), should work with 1.18+ (updated to 1.25 gp-version)
- OS : Ubuntu 22.04

## Project Status
**Phase**: Still in development

Thinking to make it encryption cautious - still boggling over the use case to optimize it over.
Reading different papers till I find that one case.