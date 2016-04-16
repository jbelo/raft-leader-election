package main

func candidateState(n *Node) StateFn {
	select {
	case <-n.timeout:
		n.logger.Printf("%d "+emphCandidate()+", timedout, term is %d\n", n.id, n.term)
		n.becomeCandidate(n.term + 1)
        return candidateState
	case t := <-n.own:
		switch t := t.(type) {
		case AppendEntriesReq:
			if t.term < n.term {
				n.logger.Printf("%d " + emphCandidate() + ", append requested for term %d, ignoring, from %d, term is %d\n", n.id, t.term, t.from, n.term)
				return candidateState
			}

			// new leader has been elected
			n.logger.Printf("%d "+emphCandidate()+", append requested for term %d, leader elected, new leader is %d, term is %d\n", n.id, t.term, t.from, n.term)
			n.replyPeer(t.from, AppendEntriesRep{from: n.id, term: n.term, success: n.log.appendEntry(t.prevLogIndex, t.prevLogTerm, t.term, t.value)})
			n.becomeFollower(t.term)
            return followerState
		case RequestVoteReq:
			// the candidate votes for itself in the current term, so grant vote only to higher terms
			if t.term <= n.term {
				n.logger.Printf("%d "+emphCandidate()+", vote request for term %d, ignoring, from %d, term is %d\n", n.id, t.term, t.from, n.term)
				return candidateState
			}

			n.logger.Printf("%d "+emphCandidate()+", vote requested for term %d, from %d, term is %d\n", n.id, t.term, t.from, n.term)
			n.replyPeer(t.from, GrantVoteRep{from: n.id, term: t.term})
			n.becomeFollower(t.term)
            return followerState
		case GrantVoteRep:
			if t.term < n.term || n.term < t.term {
				n.logger.Printf("%d "+emphCandidate()+", vote granted for term %d, from %d, ignoring, term is %d\n", n.id, t.term, t.from, n.term)
				return candidateState
			}

			n.votes++
			if n.votes < (noOfPeers+1)/2 {
				n.logger.Printf("%d " + emphCandidate() + ", vote granted for term %d, from %d, term is %d\n", n.id, t.term, t.from, n.term)
				return candidateState
			}

			n.logger.Printf("%d "+emphCandidate()+", vote granted for term %d, become "+emphLeader()+", from %d, term is %d\n", n.id, t.term, t.from, n.term)
			n.becomeLeader()
			return leaderState
        default:
            return candidateState
		}
	}
}

