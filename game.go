package main

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
)

type Game struct {
	board      *Board
	firstClick bool
	curR, curC int
	screen     tcell.Screen
}

func NewGame() *Game {
	var choice, rows, cols, mines int
	fmt.Print("난이도를 선택하세요 (1-초급, 2-중급, 3-고급): ")
	fmt.Scan(&choice)
	switch choice {
	case 1:
		rows, cols, mines = 9, 9, 10
	case 2:
		rows, cols, mines = 16, 16, 40
	case 3:
		rows, cols, mines = 16, 30, 99
	}
	return &Game{board: NewBoard(rows, cols, mines), firstClick: true}
}

func (g *Game) Run() {
	var lastRune rune
	var lastRuneTime time.Time

	// tcell 초기화
	sc, err := tcell.NewScreen()
	if err != nil {
		panic(err)
	}
	if err := sc.Init(); err != nil {
		panic(err)
	}
	defer sc.Fini()
	g.screen = sc

	g.draw()

	for {
		ev := sc.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyUp:
				if g.curR > 0 {
					g.curR--
				}
			case tcell.KeyDown:
				if g.curR < g.board.Rows-1 {
					g.curR++
				}
			case tcell.KeyLeft:
				if g.curC > 0 {
					g.curC--
				}
			case tcell.KeyRight:
				if g.curC < g.board.Cols-1 {
					g.curC++
				}
			case tcell.KeyRune:
				now := time.Now()
				// 같은 키가 100ms 안에 또 들어오면 무시
				if ev.Rune() == lastRune && now.Sub(lastRuneTime) < 100*time.Millisecond {
					continue
				}
				lastRune = ev.Rune()
				lastRuneTime = now

				switch ev.Rune() {
				case ' ':
					g.open()
				case 'f', 'F':
					g.flag()
				case 'q', 'Q':
					return
				}
			case tcell.KeyEscape:
				return
			}
			g.draw()

		case *tcell.EventResize:
			sc.Sync()
			g.draw()
		}
	}
}

func (g *Game) open() {
	r, c := g.curR, g.curC
	cell := &g.board.Cells[r][c]

	if cell.Revealed {
		return
	}
	if cell.Flagged {
		return
	}

	if g.firstClick {
		g.board.PlaceMines(r, c)
		g.board.CalcAdjacent()
		g.firstClick = false
	}

	if cell.IsMine {
		cell.Revealed = true
		g.draw()
		g.showMessage("지뢰를 밟았어요! 게임오버. (r: 재시작, q: 종료)", tcell.ColorRed)
		if g.waitQuit() {
			g.restart()
		}
		return
	}

	g.board.RevealBFS(r, c)

	if g.checkWin() {
		g.draw()
		g.showMessage("클리어! 모든 지뢰를 찾았어요! (r: 재시작, q: 종료)", tcell.ColorGreen)
		if g.waitQuit() {
			g.restart()
		}
	}
}

func (g *Game) flag() {
	r, c := g.curR, g.curC
	cell := &g.board.Cells[r][c]
	if !cell.Revealed {
		cell.Flagged = !cell.Flagged
	}
}

func (g *Game) draw() {
	sc := g.screen
	sc.Clear()

	// 상단 안내
	guide := "방향키: 이동  스페이스: 열기  f: 깃발  q: 종료"
	for i, ch := range guide {
		sc.SetContent(i, 0, ch, nil, tcell.StyleDefault.Foreground(tcell.ColorGray))
	}

	// 열 번호
	for c := 0; c < g.board.Cols; c++ {
		label := fmt.Sprintf(" %2d", c+1)
		for i, ch := range label {
			sc.SetContent(4+c*4+i, 1, ch, nil, tcell.StyleDefault.Foreground(tcell.ColorGray))
		}
	}

	// 보드
	for r := range g.board.Cells {
		// 행 번호
		label := fmt.Sprintf("%2d", r+1)
		for i, ch := range label {
			sc.SetContent(i, r+2, ch, nil, tcell.StyleDefault.Foreground(tcell.ColorGray))
		}

		for c := range g.board.Cells[r] {
			cell := g.board.Cells[r][c]
			x := 4 + c*4
			y := r + 2

			// 커서 및 주변 하이라이트 스타일
			style := tcell.StyleDefault
			if r == g.curR && c == g.curC {
				style = style.Background(tcell.ColorNavy)
			} else if abs(r-g.curR) <= 1 && abs(c-g.curC) <= 1 {
				style = style.Background(tcell.ColorDarkSlateGray)
			}

			var text string
			var color tcell.Color

			if !cell.Revealed {
				if cell.Flagged {
					text = "[F ]"
					color = tcell.ColorRed
				} else if cell.Checked {
					text = "[C ]"
					color = tcell.ColorGray
				} else {
					text = "[. ]"
					color = tcell.ColorGray
				}
			} else if cell.IsMine {
				text = "[* ]"
				color = tcell.ColorRed
			} else if cell.Adjacent > 0 {
				text = fmt.Sprintf("[%d ]", cell.Adjacent)
				color = numberColor(cell.Adjacent)
			} else {
				text = "[  ]"
				color = tcell.ColorWhite
			}

			for i, ch := range text {
				sc.SetContent(x+i, y, ch, nil, style.Foreground(color))
			}
		}
	}

	sc.Show()
}

func (g *Game) showMessage(msg string, color tcell.Color) {
	y := g.board.Rows + 3
	for i, ch := range msg {
		g.screen.SetContent(i, y, ch, nil, tcell.StyleDefault.Foreground(color))
	}
	g.screen.Show()
}

func (g *Game) waitQuit() bool {
	for {
		ev := g.screen.PollEvent()
		if ev, ok := ev.(*tcell.EventKey); ok {
			if ev.Key() == tcell.KeyEscape || ev.Rune() == 'q' || ev.Rune() == 'Q' {
				return false
			}
			if ev.Rune() == 'r' || ev.Rune() == 'R' {
				return true
			}
		}
	}
}

func (g *Game) restart() {
	g.board = NewBoard(g.board.Rows, g.board.Cols, g.board.Mines)
	g.firstClick = true
	g.curR = 0
	g.curC = 0
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

func numberColor(n int) tcell.Color {
	switch n {
	case 1:
		return tcell.ColorBlue
	case 2:
		return tcell.ColorGreen
	case 3:
		return tcell.ColorOrange
	case 4:
		return tcell.ColorNavy
	case 5:
		return tcell.ColorMaroon
	case 6:
		return tcell.ColorTeal
	default:
		return tcell.ColorWhite
	}
}
