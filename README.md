# DISYS Mandatory Exercise 2

## Ricart-Agrawala Algorithm

Our implementation is a UDP based approach to the Ricart-Agrawala algorithm for mutual exclusion on a distributed system

## Usage

The program consists of two golang files;
1. shared_resource.go -> _Represents the critical section of the system_
2. client.go -> _Represents the nodes in the system_

To run the program, first run the shared_resource.go
```bash
$ go run ./Implementation/shared_resource.go
```
Next open multiple terminals that you would like to use as clients. \
Client.go takes multiple arguments, depending on the number of clients.
- The first argument is the id of the client you want to initialize.
- The second (and other) argument(s) is the port on localhost on which the clients will operate.

Under here is an example of executing three different clients (each client to be executed in its own terminal).




```bash
$ go run ./Implementation/client.go 1 8080 8081 8082
$ go run ./Implementation/client.go 2 8080 8081 8082
$ go run ./Implementation/client.go 3 8080 8081 8082
```

To make a request to access the critical section, write the command *request* within a terminal executing the client.
```bash
$ request
```

### Authored by
- Gustav Christoffersen - _guch@itu.dk_
- Jacob Walter Bentsen - _jawb@itu.dk_
- Markus Grand Petersen - _mgrp@itu.dk_

### References
As research on the topic, we gained some insight and references by looking at implementations by [DylanNS](https://github.com/DylanNS/RicartAgrawalaAlgorithm) and [joaopmgd](https://github.com/joaopmgd/RicartAgrawala)