package main

import (
	"sync"

	pb "github.com/Markusgp/DISYS_MandaExercise2"
)

type server struct {
	pb.RingServer

	running 	bool
	waiting 	bool
	port 		string
	neighbor 	string
	tc 			chan *pb.Token
	mu 			sync.Mutex
}

