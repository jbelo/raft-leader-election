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
		// Randomly sleep for a while
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
			if t.term <= n.term {
				n.logger.Printf("%d "+emphLeader()+", append requested for term %d, ignoring, from %d, term is %d\n", n.id, t.term, t.from, n.term)
				return leaderState
			}

			// new leader has been elected
			n.logger.Printf("%d "+emphLeader()+", append requested for term %d, from %d, term is %d\n", n.id, t.term, t.from, n.term)
			n.becomeFollower(n.term)
			return followerState
		case AppendEntriesRep:
			if t.term < n.term {
				n.logger.Printf("%d "+emphLeader()+", append replied for term %d, ignoring, from %d, term is %d\n", n.id, t.term, t.from, n.term)
				return leaderState
			}

			if t.success {
				n.logger.Printf("%d "+emphLeader()+", append success replied from %d, term is %d\n", n.id, t.from, n.term)
				n.peerLogs[t.from].ack()
			} else {
				n.logger.Printf("%d "+emphLeader()+", append failure replied from %d, term is %d\n", n.id, t.from, n.term)
				n.peerLogs[t.from].nack()
			}
			return leaderState
		case RequestVoteReq:
			if t.term <= n.term {
				n.logger.Printf("%d "+emphLeader()+", vote request for term %d, ignoring, from %d, term is %d\n", n.id, t.term, t.from, n.term)
				return leaderState
			}

			n.logger.Printf("%d "+emphLeader()+", vote requested for term %d, from %d, term is %d\n", n.id, t.term, t.from, n.term)
			n.replyPeer(t.from, GrantVoteRep{from: n.id, term: t.term})
			n.becomeFollower(t.term)
			return followerState
		case GrantVoteRep:
			n.logger.Printf("%d "+emphLeader()+", vote granted for term %d, ignoring, from %d, term is %d\n", n.id, t.term, t.from, n.term)
			return leaderState
		default:
			return leaderState
		}
	}
}

