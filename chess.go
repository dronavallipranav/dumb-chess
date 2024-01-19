package main

import (
	"errors"
	"strings"
)

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

// function for taking a fen representation of a board and returning a board
func FEN(fen string) (b Board, err error) {
	parts := strings.Split(fen, " ")
	rows := strings.Split(parts[0], "/")
	if len(rows) != 8 {
		return b, errors.New("FEN should have 8 rows")
	}
	for i := 0; i < len(b); i++ {
		b[i] = ' '
	}
	for i := 0; i < 8; i++ {
		index := i*10 + 21
		for _, c := range rows[i] {
			q := Piece(c)
			if q >= '1' && q <= '8' {
				for j := Piece(0); q-j >= '1'; j++ {
					b[index] = '.'
					index++
				}
			} else if q.value() == 0 && q.Flip().value() == 0 {
				return b, errors.New("invalid piece value: " + string(c))
			} else {
				b[index] = q
				index++
			}
		}
		if index%10 != 9 {
			return b, errors.New("invalid row length")
		}
	}
	if len(parts) > 1 && parts[1] == "b" {
		b = b.Flip()
	}
	return b, nil
}

// Square represents an index of the chess board.
type Square int

const A1, H1, A8, H8 Square = 91, 98, 21, 28

func (s Square) Flip() Square   { return 119 - s }
func (s Square) String() string { return string([]byte{" abcdefgh "[s%10], "  87654321  "[s/10]}) }

// Move direction constants, horizontal moves +/-1, vertical moves +/-10
const N, E, S, W = -10, 1, 10, -1

// Move represents a movement of a piece from one square to another.
type Move struct {
	from Square
	to   Square
}

// Moves are printed in algebraic notation, i.e "e2e4".
func (m Move) String() string { return m.from.String() + m.to.String() }

// Position describes a board with the current game state (en passant and castling rules).
type Position struct {
	board Board   // current board
	score int     // board score, the higher the better
	wc    [2]bool // white castling possibilities
	bc    [2]bool // black castling possibilities
	ep    Square  // en-passant square where pawn can be captured
	kp    Square  // king passent during castling, where kind can be captured
}

// Rotate returns a modified position where the board is flipped, score is
// inverted, castling rules are preserved, en-passant and king passant rules
// are reset.
func (pos Position) Flip() Position {
	np := Position{
		score: -pos.score,
		wc:    [2]bool{pos.bc[0], pos.bc[1]},
		bc:    [2]bool{pos.wc[0], pos.wc[1]},
		ep:    pos.ep.Flip(),
		kp:    pos.kp.Flip(),
	}
	np.board = pos.board.Flip()
	return np
}

func (pos Position) Moves() (moves []Move) {
	//all possible movement directions for each piece type
	var directions = map[Piece][]Square{
		'P': {N, N + N, N + W, N + E},
		'N': {N + N + E, E + N + E, E + S + E, S + S + E, S + S + W, W + S + W, W + N + W, N + N + W},
		'B': {N + E, S + E, S + W, N + W},
		'R': {N, E, S, W},
		'Q': {N, E, S, W, N + E, S + E, S + W, N + W},
		'K': {N, E, S, W, N + E, S + E, S + W, N + W},
	}
	//iterate over all squares, considering squares with our pieces only
	for index, p := range pos.board {
		if !p.ours() {
			continue
		}
		i := Square(index)
		//iterate over all move directions for the given piece
		for _, d := range directions[p] {
			for j := i + d; ; j = j + d {
				q := pos.board[j]
				if q == ' ' || (q != '.' && q.ours()) {
					break
				}
				//if piece is a pawn then check en passant rules/capture squares
				if p == 'P' {
					if (d == N || d == N+N) && q != '.' {
						break
					}
					if d == N+N && (i < A1+N || pos.board[i+N] != '.') {
						break
					}
					if (d == N+W || d == N+E) && q == '.' && (j != pos.ep && j != pos.kp && j != pos.kp-1 && j != pos.kp+1) {
						break
					}
				}
				moves = append(moves, Move{from: i, to: j})
				//crawling pieces should stop after a single move
				if p == 'P' || p == 'N' || p == 'K' || (q != ' ' && q != '.' && !q.ours()) {
					break
				}
				//castling rules
				if i == A1 && pos.board[j+E] == 'K' && pos.wc[0] {
					moves = append(moves, Move{from: j + E, to: j + W})
				}
				if i == H1 && pos.board[j+W] == 'K' && pos.wc[1] {
					moves = append(moves, Move{from: j + W, to: j + E})
				}
			}
		}
	}
	return moves
}
