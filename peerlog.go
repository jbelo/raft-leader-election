package main

type PeerLog struct {
	nextIndex  int
	matchIndex int
}

func (ls *PeerLog) reset(nextIndex int) {
	ls.nextIndex = nextIndex
	ls.matchIndex = -1
}

func (ls *PeerLog) ack() {
	ls.matchIndex = ls.nextIndex
	ls.nextIndex++
}

func (ls *PeerLog) nack() {
	if ls.nextIndex == 0 {
		return
	}
	ls.nextIndex--
	if ls.matchIndex == ls.nextIndex {
		ls.matchIndex--
	}
}
