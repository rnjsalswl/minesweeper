package main

type Cell struct {
	IsMine   bool
	Revealed bool
	Flagged  bool
	Adjacent int
}
