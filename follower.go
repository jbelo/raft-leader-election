package main

func handleHeartbeatTimeout(n *Node) StateFn {
	n.logger.Printf("%d "+emphFollower()+", timedout, term is %d\n", n.id, n.term)
	n.becomeCandidate(n.term + 1)
	return candidateState
}

func handleAppendEntriesReq(n *Node, t AppendEntriesReq) StateFn {
	if t.term < n.term {
		n.logger.Printf("%d "+emphFollower()+", append requested for term %d, ignoring, from %d with prevLogIndex %d and prevLogTerm %d, term is %d\n", n.id, t.term, t.from, n.term)
		return followerState
	}

	n.replyPeer(t.from, AppendEntriesRep{from: n.id, term: n.term, success: n.log.appendEntry(t.prevLogIndex, t.prevLogTerm, t.term, t.value)})
	if t.term == n.term {
		n.logger.Printf("%d " + emphFollower() + ", append requested for term %d, from %d with prevLogIndex %d and prevLogTerm %d, term is %d\nlog is %v\n", n.id, t.term, t.from, t.prevLogIndex, t.prevLogTerm, n.term, n.log)
	} else {
		n.logger.Printf("%d " + emphFollower() + ", append requested for term %d, new term, from %d with prevLogIndex %d and prevLogTerm %d, term is %d\n", n.id, n.term, t.from, t.prevLogIndex, t.prevLogTerm, n.term)
		n.term = t.term
	}
	n.resetElectionTimer()
	return followerState
}

func handleRequestVoteReq(n *Node, t RequestVoteReq) StateFn {
	if t.term <= n.term {
		n.logger.Printf("%d " + emphFollower() + ", vote requested for term %d, from %d, ignoring, term is %d\n", n.id, t.term, t.from, n.term)
	} else {
		n.logger.Printf("%d " + emphFollower() + ", vote requested for term %d, from %d, term is %d\n", n.id, t.term, t.from, n.term)
		n.term = t.term
		n.replyPeer(t.from, GrantVoteRep{from: n.id, term: t.term})
	}
	return followerState
}

func followerState(n *Node) StateFn {
	select {
	case <-n.timeout:
		return handleHeartbeatTimeout(n)
	case t := <-n.own:
		switch t := t.(type) {
		case AppendEntriesReq:
			return handleAppendEntriesReq(n, t)
		case RequestVoteReq:
			return handleRequestVoteReq(n, t)
		case GrantVoteRep:
			n.logger.Printf("%d "+emphFollower()+", vote granted for term %d, ignoring, from %d, term is %d\n", n.id, t.term, t.from, n.term)
			return followerState
		default:
			return followerState
		}
	}
}

