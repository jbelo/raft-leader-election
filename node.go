package main

import (
	"log"
	"math/rand"
	"time"
)

type Node struct {
	term     int
	id       int
	votes    int
	peers    [noOfPeers]chan interface{}
	own      chan interface{}
	timeout  <-chan time.Time
	logger   *log.Logger
	log      *Log
	peerLogs [noOfPeers]PeerLog
}

func makeNode(id int, channels []chan interface{}, logger *log.Logger) Node {
	log := makeLog(logger)
	return Node{
		term:     0,
		id:       id,
		peers:    [noOfPeers]chan interface{}{channels[0], channels[1], channels[2]},
		own:      channels[id],
		logger:   logger,
		log:      &log,
		peerLogs: [noOfPeers]PeerLog{makePeerLog(&log), makePeerLog(&log), makePeerLog(&log)},
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
	n.log.appendNewEntry(LogEntry{n.term, 3})
	n.log.appendNewEntry(LogEntry{n.term, 1})
	n.log.appendNewEntry(LogEntry{n.term, 4})
	// Setup own log state
	for peer := range n.peerLogs {
		if peer != n.id {
			n.peerLogs[peer].reset()
		}
	}
}

func (n *Node) updateCommitIndex(from int) {
	matchIndex := n.peerLogs[from].matchIndex
	if matchIndex <= n.log.commitIndex {
		return
	}
	if n.log.fetchValue(matchIndex) < n.term {
		return
	}

	count := 1
	for peer, peerLog := range n.peerLogs {
		if n.id != peer && matchIndex <= peerLog.matchIndex {
			count++
			if (noOfPeers + 1) / 2 == count {
				n.log.commitIndex = peerLog.matchIndex
				return
			}
		}
	}
}

func (n *Node) peerLogState(peer int) (int, int) {
	prevLogIndex := n.peerLogs[peer].nextIndex - 1
	if prevLogIndex == -1 {
		return prevLogIndex, NULLTERM
	}
	return prevLogIndex, n.log.fetchTerm(prevLogIndex)
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
			value := NULLVALUE
			if n.log.hasEntryAt(prevLogIndex + 1) {
				value = n.log.fetchValue(prevLogIndex + 1)
			}
			select {
			case ch <- AppendEntriesReq{term: n.term, from: n.id, prevLogIndex: prevLogIndex, prevLogTerm: prevLogTerm, value: value}:
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
