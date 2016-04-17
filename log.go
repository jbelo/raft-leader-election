package main

import "log"

type LogEntry struct {
	term  int
	value int
}

type Log struct {
	logger      *log.Logger
	entries     []LogEntry
	commitIndex int
}

func makeLog(logger *log.Logger) Log {
	return Log{logger: logger, entries: make([]LogEntry, 0), commitIndex: -1}
}

func (log *Log) appendAcceptable(prevLogIndex int, prevLogTerm int) bool {
	atIndex := prevLogIndex+1
	if atIndex == 0 {
		return true
	}

	if 0 < atIndex && atIndex <= len(log.entries) {
		return log.entries[prevLogIndex].term == prevLogTerm
	}

	return false
}

func (log *Log) appendNotAcceptable(prevLogIndex int, prevLogTerm int) bool {
	return !log.appendAcceptable(prevLogIndex, prevLogTerm)
}

func (log *Log) appendEntry(prevLogIndex int, prevLogTerm int, term int, value int) bool {
	if log.appendNotAcceptable(prevLogIndex, prevLogTerm) {
		return false
	}

	if value == NULLVALUE {
		return true
	}

	log.entries = append(log.entries[0 : prevLogIndex+1], LogEntry{term: term, value: value})

	return true
}

func (log *Log) appendNewEntry(LogEntry LogEntry) {
	log.entries = append(log.entries, LogEntry)
}

func (log *Log) fetchTerm(index int) int {
	if index < 0 || len(log.entries) <= index {
		log.logger.Printf("Invalid fetch term index specified %d (%d)\n", index, len(log.entries))
	}
	return log.entries[index].term
}

func (log *Log) fetchValue(index int) int {
	if index < 0 || len(log.entries) <= index {
		log.logger.Printf("Invalid fetch value index specified %d (%d)\n", index, len(log.entries))
	}
	return log.entries[index].value
}

func (log *Log) hasEntryAt(index int) bool {
	return index < len(log.entries)
}

func (log *Log) lastIndex() int {
	return len(log.entries) - 1
}

func (log *Log) lastTerm() int {
	if log.isEmpty() {
		return NULLTERM
	}
	return log.entries[log.lastIndex()].term
}

func (log *Log) noOfEntries() int {
	return len(log.entries)
}

func (log *Log) isEmpty() bool {
	return len(log.entries) == 0
}

func (log *Log) moreRecentThan(lastLogIndex int, lastLogTerm int) bool {
	if log.isEmpty() && lastLogIndex == -1 {
		return false
	}
	if lastLogIndex == -1 {
		return true
	}
	if log.lastTerm() > lastLogTerm  {
		return true
	}
	if log.lastTerm() < lastLogTerm {
		return false
	}
	return log.lastIndex() > lastLogIndex
}
