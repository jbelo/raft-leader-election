package main

import (
	"time"
	"log"
	"math/rand"
)

type Node struct {
	term     int
	id       int
	votes    int
	peers    []chan interface{}
	own      chan interface{}
	timeout  <-chan time.Time
	logger   *log.Logger
	peerLogs [noOfPeers]PeerLog
	log      Log
}

func makeNode(id int, channels []chan interface{}, logger *log.Logger) Node {
	return Node{
		term:   0,
		id:     id,
		peers:  []chan interface{}{channels[0], channels[1], channels[2]},
		own:	channels[id],
		logger: logger,
		log:    Log{make([]LogEntry, 0), 0},
	}
}

func (n *Node) replyPeer(peerId int, reply interface{}) {
	n.peers[peerId] <- reply
}

func (n *Node) resetElectionTimer() {
	n.timeout = time.After(time.Duration(timeoutInMs+rand.Intn(150)) * time.Millisecond)
}

func (n *Node) resetHeartbeatTimer() {
	n.timeout = time.After(time.Duration(heartbeatInMs) * time.Millisecond)
}

func (n *Node) resetPeerLogs() {
	// generate a few entries
	n.log.entries = append(n.log.entries, LogEntry{n.term, 3})
	n.log.entries = append(n.log.entries, LogEntry{n.term, 1})
	n.log.entries = append(n.log.entries, LogEntry{n.term, 4})
	// Setup own log state
	for peer := range n.peerLogs {
		if peer != n.id {
			n.peerLogs[peer].reset(n.log.noOfEntries())
		}
	}
}

func (n *Node) peerLogState(peer int) (int, int) {
	prevLogIndex := n.peerLogs[peer].nextIndex - 1
	if n.log.isEmpty() {
		return prevLogIndex, -1
	}
	return prevLogIndex, n.log.entries[prevLogIndex].term
}

func (n *Node) appendAcceptable(prevLogIndex int, prevLogTerm int) bool {
	return n.log.appendAcceptable(prevLogTerm, prevLogTerm)
}

func (n *Node) becomeFollower(term int) {
	n.term = term
	n.resetElectionTimer()
}

func (n *Node) becomeCandidate(term int) {
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

func (n *Node) sendHeartbeats() {
	for peer, ch := range n.peers {
		if n.id != peer {
			prevLogIndex, prevLogTerm := n.peerLogState(peer)
			n.logger.Printf("%d "+emphLeader()+", heartbeating, sent to %d with prevLogIndex %d and prevLogTerm %d, term is %d\n", n.id, peer, prevLogIndex, prevLogTerm, n.term)
			select {
			case ch <- AppendEntriesReq{term: n.term, from: n.id, prevLogIndex: prevLogIndex, prevLogTerm: prevLogTerm}:
			default:
			}
		}
	}
	n.resetHeartbeatTimer()
}

func (n *Node) becomeLeader() {
	n.resetPeerLogs()
	n.sendHeartbeats()
}


func (n *Node) Run() {
	n.becomeFollower(0)
	for state := followerState(n); n != nil; {
		state = state(n)
	}
}
