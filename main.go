package main

import (
	"math/rand"
	"time"
	"log"
	"os"
)

const (
	follower = iota
	candidate
	leader
)

const (
	noOfPeers        = 3
	chanceOfGivingUp = 20
)

const (
	timeoutInMs   = 1000
	heartbeatInMs = 150
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
	from int
	term int
}

type RequestVoteReq struct {
	from int
	term int
}

type GrantVoteRep struct {
	from int
	term int
}

type Node struct {
	role    int
	term    int
	id      int
	votes   int
	peers   []chan interface{}
	timeout <-chan time.Time
	logger  *log.Logger
}

func (n *Node) resetElectionTimer() {
	n.timeout = time.After(time.Duration(timeoutInMs+rand.Intn(150)) * time.Millisecond)
}

func (n *Node) resetHeartbeatTimer() {
	n.timeout = time.After(time.Duration(heartbeatInMs) * time.Millisecond)
}

func (n *Node) becomeFollower(term int) {
	n.role = follower
	n.term = term
	n.resetElectionTimer()
}

func (n *Node) becomeCandidate(term int) {
	n.role = candidate
	n.term = term
	// candidate votes for itself
	n.votes = 1
	for peer, ch := range n.peers {
		if peer != n.id {
			n.logger.Printf("%d "+emphCandidate()+", requested vote to %d, term is %d\n", n.id, peer, n.term)
			select {
			case ch <- RequestVoteReq{term: n.term, from: n.id}:
			default:
			}
		}
	}
	n.resetElectionTimer()
}

func (n *Node) becomeLeader() {
	n.role = leader
	for peer, ch := range n.peers {
		if n.id != peer {
			n.logger.Printf("%d "+emphLeader()+", heartbeating, sent to %d, term is %d\n", n.id, peer, n.term)
			select {
			case ch <- AppendEntriesReq{term: n.term, from: n.id}:
			default:
			}
		}
	}
	n.resetHeartbeatTimer()
}

func (n *Node) followerState() {
	select {
	case <-n.timeout:
		n.logger.Printf("%d "+emphFollower()+", timedout, term is %d\n", n.id, n.term)
		n.becomeCandidate(n.term + 1)
	case t := <-n.peers[n.id]:
		switch t := t.(type) {
		case AppendEntriesReq:
			if n.term == t.term {
				n.logger.Printf("%d "+emphFollower()+", append requested for term %d, from %d, term is %d\n", n.id, t.term, t.from, n.term)
				n.resetElectionTimer()
			} else if n.term < t.term {
				n.logger.Printf("%d "+emphFollower()+", append requested for term %d, new term, from %d, term is %d\n", n.id, n.term, t.from, n.term)
				n.becomeFollower(t.term)
			} else {
				n.logger.Printf("%d "+emphFollower()+", append requested for term %d, ignoring, from %d, term is %d\n", n.id, t.term, t.from, n.term)
			}
		case RequestVoteReq:
			if n.term < t.term {
				n.logger.Printf("%d "+emphFollower()+", vote requested for term %d, from %d, term is %d\n", n.id, t.term, t.from, n.term)
				n.term = t.term
				n.peers[t.from] <- GrantVoteRep{from: n.id, term: t.term}
			} else {
				n.logger.Printf("%d "+emphFollower()+", vote requested for term %d, from %d, ignoring, term is %d\n", n.id, t.term, t.from, n.term)
			}
		case GrantVoteRep:
			n.logger.Printf("%d "+emphFollower()+", vote granted for term %d, ignoring, from %d, term is %d\n", n.id, t.term, t.from, n.term)
		}
	}
}

func (n *Node) candidateState() {
	select {
	case <-n.timeout:
		n.logger.Printf("%d "+emphCandidate()+", timedout, term is %d\n", n.id, n.term)
		n.becomeCandidate(n.term + 1)
	case t := <-n.peers[n.id]:
		switch t := t.(type) {
		case AppendEntriesReq:
			// new leader has been elected
			if n.term <= t.term {
				n.logger.Printf("%d "+emphCandidate()+", append requested for term %d, leader elected, new leader is %d, term is %d\n", n.id, t.term, t.from, n.term)
				n.becomeFollower(t.term)
			} else {
				n.logger.Printf("%d "+emphCandidate()+", append requested for term %d, ignoring, from %d, term is %d\n", n.id, t.term, t.from, n.term)
			}
		case RequestVoteReq:
			// the candidate votes for itself in the current term, so grant vote only to higher terms
			if n.term < t.term {
				n.logger.Printf("%d "+emphCandidate()+", vote requested for term %d, from %d, term is %d\n", n.id, t.term, t.from, n.term)
				n.peers[t.from] <- GrantVoteRep{from: n.id, term: t.term}
				n.becomeFollower(t.term)
			} else {
				n.logger.Printf("%d "+emphCandidate()+", vote request for term %d, ignoring, from %d, term is %d\n", n.id, t.term, t.from, n.term)
			}
		case GrantVoteRep:
			if n.term == t.term {
				n.votes++
				if n.votes == (noOfPeers+1)/2 {
					n.logger.Printf("%d "+emphCandidate()+", vote granted for term %d, become "+emphLeader()+", from %d, term is %d\n", n.id, t.term, t.from, n.term)
					n.becomeLeader()
				} else {
					n.logger.Printf("%d "+emphCandidate()+", vote granted for term %d, from %d, term is %d\n", n.id, t.term, t.from, n.term)
				}
			} else {
				n.logger.Printf("%d "+emphCandidate()+", vote granted for term %d, from %d, ignoring, term is %d\n", n.id, t.term, t.from, n.term)
			}
		}
	}
}

func (n *Node) leaderState() {
	select {
	case <-n.timeout:
		// Randomly give up leading
		//		if rand.Intn(chanceOfGivingUp) == 0 {
		//			n.logger.Printf("%d " + emphLeader() + ", give up, term is %d\n", n.id, n.term)
		//			n.becomeFollower(n.term)
		//		} else {
		//			n.becomeLeader()
		//		}
		// Randomly sleep for 5 seconds
		if rand.Intn(chanceOfGivingUp) == 0 {
			n.logger.Printf("%d "+emphLeader()+", going to sleep for 2 secs..., term is %d\n", n.id, n.term)
			time.Sleep(2 * time.Second)
			n.becomeLeader()
		} else {
			n.becomeLeader()
		}
	case t := <-n.peers[n.id]:
		switch t := t.(type) {
		case AppendEntriesReq:
			// new leader has been elected
			if n.term < t.term {
				n.logger.Printf("%d "+emphLeader()+", append requested for term %d, from %d, term is %d\n", n.id, t.term, t.from, n.term)
				n.becomeFollower(n.term)
			} else {
				n.logger.Printf("%d "+emphLeader()+", append requested for term %d, ignoring, from %d, term is %d\n", n.id, t.term, t.from, n.term)
			}
		case RequestVoteReq:
			if n.term < t.term {
				n.logger.Printf("%d "+emphLeader()+", vote requested for term %d, from %d, term is %d\n", n.id, t.term, t.from, n.term)
				n.peers[t.from] <- GrantVoteRep{from: n.id, term: t.term}
				n.becomeFollower(t.term)
			} else {
				n.logger.Printf("%d "+emphLeader()+", vote request for term %d, ignoring, from %d, term is %d\n", n.id, t.term, t.from, n.term)
			}
		case GrantVoteRep:
			n.logger.Printf("%d "+emphLeader()+", vote granted for term %d, ignoring, from %d, term is %d\n", n.id, t.term, t.from, n.term)
		}
	}
}

func (n *Node) run() {
	n.becomeFollower(0)
	for {
		switch n.role {
		case follower:
			n.followerState()
		case candidate:
			n.candidateState()
		case leader:
			n.leaderState()
		}
	}
}

func main() {
	logger := log.New(os.Stdout, "", log.Lmicroseconds|log.Lshortfile)
	logger.Printf("Starting up...")
	channels := []chan interface{}{make(chan interface{}, 2), make(chan interface{}, 2), make(chan interface{}, 2)}
	nodes := []Node{
		Node{role: follower, term: 0, id: 0, peers: []chan interface{}{channels[0], channels[1], channels[2]}, logger: logger},
		Node{role: follower, term: 0, id: 1, peers: []chan interface{}{channels[0], channels[1], channels[2]}, logger: logger},
		Node{role: follower, term: 0, id: 2, peers: []chan interface{}{channels[0], channels[1], channels[2]}, logger: logger},
	}
	go nodes[0].run()
	go nodes[1].run()
	go nodes[2].run()
	time.Sleep(60 * time.Second)
	logger.Printf("Finished")
}
