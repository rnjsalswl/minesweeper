package main

import "math/rand"

type Board struct {
	Rows, Cols, Mines int
	Cells             [][]Cell
}

func NewBoard(rows, cols, mines int) *Board {
	board := &Board{
		Rows:  rows,
		Cols:  cols,
		Mines: mines,
		Cells: make([][]Cell, rows),
	}

	for i := range board.Cells {
		board.Cells[i] = make([]Cell, cols)
	}

	return board
}

func (b *Board) PlaceMines(safeR, safeC int) {
	// 첫 클릭에는 안전한 위치를 보장하기 위해, safeR과 safeC 주변 3x3 영역에는 지뢰를 배치하지 않습니다.
	placed := 0
	for placed < b.Mines {

		r := rand.Intn(b.Rows)
		c := rand.Intn(b.Cols)

		if abs(r-safeR) <= 1 && abs(c-safeC) <= 1 {
			continue // 안전한 영역
		}

		if b.Cells[r][c].IsMine {
			continue
		}

		b.Cells[r][c].IsMine = true
		placed++
	}
}

var dirs = [8][2]int{
	{-1, -1}, {-1, 0}, {-1, 1},
	{0, -1}, {0, 1},
	{1, -1}, {1, 0}, {1, 1},
}

func (b *Board) CalcAdjacent() {
	for r := range b.Cells {
		for c := range b.Cells[r] {
			if !b.Cells[r][c].IsMine {
				continue
			}
			// 지뢰 칸 주변 8방향에 +1
			for _, d := range dirs {
				nr, nc := r+d[0], c+d[1]
				if nr >= 0 && nr < b.Rows && nc >= 0 && nc < b.Cols {
					if !b.Cells[nr][nc].IsMine {
						b.Cells[nr][nc].Adjacent++
					}
				}
			}
		}
	}
}

func (b *Board) RevealBFS(startR, startC int) {
	queue := [][2]int{{startR, startC}}
	b.Cells[startR][startC].Revealed = true

	for len(queue) > 0 {
		r, c := queue[0][0], queue[0][1]
		queue = queue[1:]

		if b.Cells[r][c].Adjacent == 0 {
			for _, d := range dirs {
				nr, nc := r+d[0], c+d[1]
				if nr >= 0 && nr < b.Rows && nc >= 0 && nc < b.Cols {
					cell := &b.Cells[nr][nc]
					if !cell.Revealed && !cell.IsMine && !cell.Flagged && !cell.Checked {
						cell.Revealed = true
						queue = append(queue, [2]int{nr, nc})
					}
				}
			}
		}
	}
}
