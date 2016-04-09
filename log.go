package main

type LogEntry struct {
	term  int
	value int
}

type Log struct {
	entries     []LogEntry
	commitIndex int
}

func (log *Log) appendAcceptable(prevLogIndex int, prevLogTerm int) bool {
	if 0 == prevLogIndex+1 {
		return true
	}

	if 0 < prevLogIndex+1 && prevLogIndex+1 <= len(log.entries) {
		return log.entries[prevLogIndex].term == prevLogTerm
	}

	return false
}

func (log *Log) AppendEntry(prevLogIndex int, prevLogTerm int, logEntry LogEntry) bool {
	if !log.appendAcceptable(prevLogIndex, prevLogTerm) {
		return false
	}
	if prevLogIndex+1 == len(log.entries) {
		log.entries = append(log.entries, logEntry)
	} else {
		log.entries[prevLogIndex+1] = logEntry
		log.entries = log.entries[0 : prevLogIndex+1]
	}
	return true
}

func (log *Log) noOfEntries() {
	return len(log.entries)
}

func (log *Log) isEmpty() {
	return len(log.entries) == 0
}
