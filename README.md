# raft-leader-election

> Studying the Raft Consensus Algorithm, Leader Election

An implementation of the Leader Election component of Raft. The leader sporadically gives up to force new election.

# Build

`go build`

# Run

`./raft`

# Observations

1. A new term is started iff some node fails to receive a heartbeat from the leader or a leader fails to be elected;
2. A node moves to a new term iff it becomes a candidate or receives a message from another node, necessarily a candidate or a leader, already in that term;
3. A vote is granted to a node iff the node is in a higher term;

Note that if there is a log to maintain, then:

4. Entries are considered committed iff they are acknowledged by a majority (see example bellow about the possibility of two leaders);

# Example of Pathological Cases

## Two leaders In Two Distinct Terms
## (But Only One Having a Majority of Followers)

- Configuration:
  - A seven node cluster
  - One leader in term n with two followers
  - Another leader in term n+k with three followers
  - Only the other leader forms a majority with its followers;
- How:
  - The one leader was first granted four votes;
  - The other leader was then granted four votes, one of them from followers of the one leader;

or simpler,

- Configuration:
  - A three node cluster
  - One leader in term n with no followers
  - Another leader in term n+k with one follower
  - Only the other leader forms a majority with its follower;
- How:
  - The one leader was first granted one vote;
  - The other leader was then granted one vote, from the one leader follower;

# References

- [The Raft Consensus Algorithm](https://raft.github.io/)
- [The Raft Paper](http://ramcloud.stanford.edu/raft.pdf)
- [Diego Ongaro's Ph.D. Dissertation](https://github.com/ongardie/dissertation#readme)
