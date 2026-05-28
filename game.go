package main

import (
	"fmt"
)

type Game struct {
	board      *Board
	firstClick bool
}

func NewGame() *Game {
	// 난이도 입력 받기 (예: 1-초급, 2-중급, 3-고급)
	var choice, rows, cols, mines int
	fmt.Print("난이도를 선택하세요 (1-초급, 2-중급, 3-고급): ")
	fmt.Scanf("%d", &choice)
	switch choice {
	case 1:
		rows, cols, mines = 9, 9, 10
	case 2:
		rows, cols, mines = 16, 16, 40
	case 3:
		rows, cols, mines = 16, 30, 99
	}

	// newBoard 생성
	return &Game{board: NewBoard(rows, cols, mines), firstClick: true}
}

// 게임 보드 출력
//
//	1  2  3  4  5
//
// 1  [ ][ ][1][ ][ ]
// 2  [ ][F][ ][ ][ ]
// 3  [1][1][ ][ ][ ]
func (g *Game) printBoard() {

	b := g.board

	// 열 번호
	fmt.Print("  ")
	for c := 0; c < b.Cols; c++ {
		fmt.Printf(" %2d ", c+1)
	}
	fmt.Println()

	for r := range b.Cells {
		// 행 번호
		fmt.Printf("%2d ", r+1)
		for c := range b.Cells[r] {
			cell := b.Cells[r][c]
			if !cell.Revealed {
				if cell.Flagged {
					fmt.Print("[F ]")
				} else {
					fmt.Print("[. ]")
				}
			} else if cell.IsMine {
				fmt.Print("[* ]")
			} else if cell.Adjacent > 0 {
				fmt.Printf("[%d ]", cell.Adjacent)
			} else {
				fmt.Print("[  ]") // 열린 빈 칸
			}
		}
		fmt.Println()
	}
}

func (g *Game) Run() {
	// 첫 화면
	g.printBoard()

	for {
		var r, c int
		var action string
		fmt.Print("행 열 액션 (o/f): ")
		fmt.Scan(&r, &c, &action)

		// 0 인덱스로 변환
		r--
		c--

		// 범위 체크
		if r < 0 || r >= g.board.Rows || c < 0 || c >= g.board.Cols {
			fmt.Println("올바른 좌표를 입력하세요.")
			continue
		}

		switch action {
		case "o":
			// 첫 클릭이면 지뢰 배치
			if g.firstClick {
				g.board.PlaceMines(r, c)
				g.board.CalcAdjacent()
				g.firstClick = false
			}

			cell := &g.board.Cells[r][c]

			if cell.Revealed {
				fmt.Println("이미 열린 칸이에요.")
				continue
			}
			if cell.Flagged {
				fmt.Println("깃발이 있는 칸이에요. f로 먼저 해제하세요.")
				continue
			}

			if cell.IsMine {
				cell.Revealed = true
				g.printBoard()
				fmt.Println("지뢰를 밟았어요! 게임오버.")
				return
			}

			g.board.RevealBFS(r, c)

		case "f":
			cell := &g.board.Cells[r][c]
			if cell.Revealed {
				fmt.Println("이미 열린 칸이에요.")
				continue
			}
			cell.Flagged = !cell.Flagged

		default:
			fmt.Println("o 또는 f를 입력하세요.")
			continue
		}

		g.printBoard()

		// 승리 체크
		if g.checkWin() {
			fmt.Println("클리어! 모든 지뢰를 찾았어요!")
			return
		}
	}
}

func (g *Game) checkWin() bool {
	for r := range g.board.Cells {
		for c := range g.board.Cells[r] {
			cell := g.board.Cells[r][c]
			if !cell.IsMine && !cell.Revealed {
				return false
			}
		}
	}
	return true
}
