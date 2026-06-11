package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gdamore/tcell/v2"
)

type Game struct {
	board             *Board
	firstClick        bool
	curR, curC        int
	screen            tcell.Screen
	hintCount         int
	rows, cols, mines int
	quit              bool
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
	return &Game{
		board:      NewBoard(rows, cols, mines),
		firstClick: true,
		rows:       rows,
		cols:       cols,
		mines:      mines,
	}
}

func (g *Game) Run() {
	var lastRune rune
	var lastRuneTime time.Time

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
				case 'c', 'C':
					g.comment()
				case 'h', 'H':
					g.hint()
				case 'r', 'R':
					g.restart()
				case 'q', 'Q':
					return
				}
			case tcell.KeyEscape:
				return
			}
			if g.quit {
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
		g.waitQuitOrRestart()
		return
	}

	g.board.RevealBFS(r, c)

	if g.checkWin() {
		g.draw()
		g.showMessage("클리어! 모든 지뢰를 찾았어요! (r: 재시작, q: 종료)", tcell.ColorGreen)
		g.waitQuitOrRestart()
	}
}

func (g *Game) flag() {
	r, c := g.curR, g.curC
	cell := &g.board.Cells[r][c]
	if !cell.Revealed {
		cell.Flagged = !cell.Flagged
		if cell.Flagged {
			cell.Commented = false
		}
	}
}

func (g *Game) comment() {
	r, c := g.curR, g.curC
	cell := &g.board.Cells[r][c]
	if !cell.Revealed {
		cell.Commented = !cell.Commented
		if cell.Commented {
			cell.Flagged = false
		}
	}
}

func (g *Game) hint() {
	if g.firstClick {
		return
	}
	var safeCells [][2]int
	for r := range g.board.Cells {
		for c := range g.board.Cells[r] {
			cell := g.board.Cells[r][c]
			if !cell.IsMine && !cell.Revealed && !cell.Flagged {
				safeCells = append(safeCells, [2]int{r, c})
			}
		}
	}
	rand.Shuffle(len(safeCells), func(i, j int) {
		safeCells[i], safeCells[j] = safeCells[j], safeCells[i]
	})
	count := 0
	for _, pos := range safeCells {
		if count >= 3 {
			break
		}
		g.board.Cells[pos[0]][pos[1]].Revealed = true
		count++
	}
	g.hintCount++
}

func (g *Game) restart() {
	g.board = NewBoard(g.rows, g.cols, g.mines)
	g.firstClick = true
	g.curR = 0
	g.curC = 0
	g.hintCount = 0
}

func (g *Game) countFlags() int {
	count := 0
	for r := range g.board.Cells {
		for c := range g.board.Cells[r] {
			if g.board.Cells[r][c].Flagged {
				count++
			}
		}
	}
	return count
}

func (g *Game) draw() {
	sc := g.screen
	sc.Clear()

	flags := g.countFlags()
	guide := fmt.Sprintf("방향키:이동  스페이스:열기  f:깃발  c:메모  h:힌트(%d회)  r:재시작  q:종료   |  깃발:%d / 지뢰:%d", g.hintCount, flags, g.board.Mines)
	for i, ch := range guide {
		sc.SetContent(i, 0, ch, nil, tcell.StyleDefault.Foreground(tcell.ColorGray))
	}

	for c := 0; c < g.board.Cols; c++ {
		label := fmt.Sprintf(" %2d", c+1)
		for i, ch := range label {
			sc.SetContent(4+c*4+i, 1, ch, nil, tcell.StyleDefault.Foreground(tcell.ColorGray))
		}
	}

	for r := range g.board.Cells {
		label := fmt.Sprintf("%2d", r+1)
		for i, ch := range label {
			sc.SetContent(i, r+2, ch, nil, tcell.StyleDefault.Foreground(tcell.ColorGray))
		}

		for c := range g.board.Cells[r] {
			cell := g.board.Cells[r][c]
			x := 4 + c*4
			y := r + 2

			style := tcell.StyleDefault
			isCursor := r == g.curR && c == g.curC
			isNeighbor := !isCursor && abs(r-g.curR) <= 1 && abs(c-g.curC) <= 1

			if isCursor {
				style = style.Background(tcell.ColorNavy)
			} else if isNeighbor {
				style = style.Background(tcell.NewRGBColor(30, 30, 80))
			}

			var text string
			var color tcell.Color

			if !cell.Revealed {
				if cell.Flagged {
					text = "[F ]"
					color = tcell.ColorRed
				} else if cell.Commented {
					text = "[? ]"
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

func (g *Game) waitQuitOrRestart() {
	for {
		ev := g.screen.PollEvent()
		if ev, ok := ev.(*tcell.EventKey); ok {
			switch {
			case ev.Key() == tcell.KeyEscape || ev.Rune() == 'q' || ev.Rune() == 'Q':
				g.quit = true
				return
			case ev.Rune() == 'r' || ev.Rune() == 'R':
				g.restart()
				g.draw()
				return
			}
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
