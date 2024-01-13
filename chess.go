package main

type Piece byte

// Value returns the value of the piece
func (p Piece) value() int {
	switch p {
	case 'P':
		return 100
	case 'N':
		return 280
	case 'B':
		return 320
	case 'R':
		return 479
	case 'Q':
		return 929
	case 'K':
		return 60000
	default:
		return 0
	}
}

// Ours checks if the piece is ours
func (p Piece) ours() bool {
	return p.value() > 0
}

// Flip changes the case of the piece
func (p Piece) Flip() Piece {
	switch p {
	case 'P':
		return 'p'
	case 'N':
		return 'n'
	case 'B':
		return 'b'
	case 'R':
		return 'r'
	case 'Q':
		return 'q'
	case 'K':
		return 'k'
	case 'p':
		return 'P'
	case 'n':
		return 'N'
	case 'b':
		return 'B'
	case 'r':
		return 'R'
	case 'q':
		return 'Q'
	case 'k':
		return 'K'
	default:
		return p
	}
}

type Board [120]Piece

// Flip creates a flipped version of the board
func (a Board) Flip() (b Board) {
	for i := len(a) - 1; i >= 0; i-- {
		b[i] = a[len(a)-i-1].Flip()
	}
	return b
}

// String returns a human-readable board representation as a 8x8 square with
// pieces and dots.
func (a Board) String() (s string) {
	s = "\n"
	for row := 2; row < 10; row++ {
		for col := 1; col < 9; col++ {
			s = s + string(a[row*10+col])
		}
		s = s + "\n"
	}
	return s
}
