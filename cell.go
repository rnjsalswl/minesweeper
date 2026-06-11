package main

type Cell struct {
	IsMine    bool
	Revealed  bool
	Flagged   bool
	Commented bool
	Adjacent  int
}
