package main

type PeerLog struct {
	log 	   *Log
	nextIndex  int		// 0 <= nextIndex <= log.noOfEntries()
	matchIndex int		// -1 <= matchIndex < nextIndex
}

func makePeerLog(log *Log) PeerLog {
	return PeerLog{log: log}
}

func (ls *PeerLog) reset() {
	ls.nextIndex = ls.log.noOfEntries()
	ls.matchIndex = -1
}

func (ls *PeerLog) ack() {
	if ls.nextIndex < ls.log.noOfEntries() {
		ls.matchIndex = ls.nextIndex
		ls.nextIndex++
	} else {
		ls.matchIndex = ls.nextIndex - 1
	}
}

func (ls *PeerLog) nack() {
	if ls.nextIndex == 0 {
		return
	}
	ls.nextIndex--
}
