package main

import (
	"log"
	"os"
	"time"
)

const (
	noOfPeers        = 3
	chanceOfGivingUp = 20
)

const (
	timeoutInMs   = 1000
	heartbeatInMs = 150
)

const (
	NULLTERM  = -1
	NULLVALUE = -1
)

func emphFollower() string {
	return "\033[32m\033[1mfollower\033[0m"
}

func emphCandidate() string {
	return "\033[33m\033[1mcandidate\033[0m"
}

func emphLeader() string {
	return "\033[31m\033[1mleader\033[0m"
}

type AppendEntriesReq struct {
	from  int
	term  int
	value int

	prevLogIndex int
	prevLogTerm  int
}

type AppendEntriesRep struct {
	from    int
	term    int
	success bool
}

type RequestVoteReq struct {
	from int
	term int
}

type GrantVoteRep struct {
	from int
	term int
}

type StateFn func(n *Node) StateFn

func main() {
	logger := log.New(os.Stdout, "", log.Lmicroseconds|log.Lshortfile)
	logger.Printf("Starting up...")
	channels := []chan interface{}{make(chan interface{}, 2), make(chan interface{}, 2), make(chan interface{}, 2)}
	nodes := [3]Node{makeNode(0, channels, logger), makeNode(1, channels, logger), makeNode(2, channels, logger)}
	go nodes[0].Run()
	go nodes[1].Run()
	go nodes[2].Run()
	time.Sleep(600 * time.Second)
	logger.Printf("Finished")
}
