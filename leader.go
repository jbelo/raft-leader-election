package main

import (
	"time"
	"math/rand"
)

func leaderState(n *Node) StateFn {
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
			n.sendHeartbeats()
		} else {
			n.sendHeartbeats()
		}
		return leaderState
	case t := <-n.own:
		switch t := t.(type) {
		case AppendEntriesReq:
			// new leader has been elected
			if n.term < t.term {
				n.logger.Printf("%d "+emphLeader()+", append requested for term %d, from %d, term is %d\n", n.id, t.term, t.from, n.term)
				n.becomeFollower(n.term)
				return followerState
			} else {
				n.logger.Printf("%d "+emphLeader()+", append requested for term %d, ignoring, from %d, term is %d\n", n.id, t.term, t.from, n.term)
			}
			return leaderState
		case AppendEntriesRep: //TODO
			if t.success {
				n.peerLogs[t.from].ack()
			} else {
				n.peerLogs[t.from].nack()
			}
			return leaderState
		case RequestVoteReq:
			if n.term < t.term {
				n.logger.Printf("%d "+emphLeader()+", vote requested for term %d, from %d, term is %d\n", n.id, t.term, t.from, n.term)
				n.peers[t.from] <- GrantVoteRep{from: n.id, term: t.term}
				n.becomeFollower(t.term)
				return followerState
			} else {
				n.logger.Printf("%d "+emphLeader()+", vote request for term %d, ignoring, from %d, term is %d\n", n.id, t.term, t.from, n.term)
			}
			return leaderState
		case GrantVoteRep:
			n.logger.Printf("%d "+emphLeader()+", vote granted for term %d, ignoring, from %d, term is %d\n", n.id, t.term, t.from, n.term)
			return leaderState
		default:
			return leaderState
		}
	}
}

