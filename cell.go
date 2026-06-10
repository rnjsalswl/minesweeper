package main

type Cell struct {
	IsMine   bool
	Revealed bool
	Flagged  bool
	Checked  bool
	Adjacent int
}
