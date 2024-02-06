package main

import (
	"errors"
	"strings"
)

type Piece byte

func abs(n int) int { return int((int64(n) ^ int64(n)>>63) - int64(n)>>63) }

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

// Move returns a modified rotated position after the move is applied.
func (pos Position) Move(m Move) (np Position) {
	i, j, p := m.from, m.to, pos.board[m.from]
	np = pos
	np.ep = 0
	np.kp = 0
	np.score = pos.score + pos.value(m)
	np.board[m.to] = pos.board[m.from]
	np.board[m.from] = '.'
	if i == A1 {
		np.wc[0] = false
	}
	if i == H1 {
		np.wc[1] = false
	}
	if j == A8 {
		np.bc[1] = false
	}
	if j == H8 {
		np.bc[0] = false
	}
	if p == 'K' {
		np.wc[0], np.wc[1] = false, false
		if abs(int(j-i)) == 2 {
			if j < i {
				np.board[H1] = '.'
			} else {
				np.board[A1] = '.'
			}
			np.board[(i+j)/2] = 'R'
		}
	}
	if p == 'P' {
		// Pawn promotion
		if A8 <= j && j <= H8 {
			np.board[j] = 'Q'
		}
		// First pawn move
		if j-i == 2*N {
			np.ep = i + N
		}
		// En-passant capture
		if j == pos.ep {
			np.board[j+S] = '.'
		}
	}
	return np.Flip()
}

// Value returns the score of the current position if the move is applied.
func (pos Position) value(m Move) int {
	pst := map[Piece][120]int{
		'P': {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 178, 183, 186, 173, 202, 182, 185, 190, 0, 0, 107, 129, 121, 144, 140, 131, 144, 107, 0, 0, 83, 116, 98, 115, 114, 0, 115, 87, 0, 0, 74, 103, 110, 109, 106, 101, 0, 77, 0, 0, 78, 109, 105, 89, 90, 98, 103, 81, 0, 0, 69, 108, 93, 63, 64, 86, 103, 69, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		'N': {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 214, 227, 205, 205, 270, 225, 222, 210, 0, 0, 277, 274, 380, 244, 284, 342, 276, 266, 0, 0, 290, 347, 281, 354, 353, 307, 342, 278, 0, 0, 304, 304, 325, 317, 313, 321, 305, 297, 0, 0, 279, 285, 311, 301, 302, 315, 282, 0, 0, 0, 262, 290, 293, 302, 298, 295, 291, 266, 0, 0, 257, 265, 282, 0, 282, 0, 257, 260, 0, 0, 206, 257, 254, 256, 261, 245, 258, 211, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		'B': {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 261, 242, 238, 244, 297, 213, 283, 270, 0, 0, 309, 340, 355, 278, 281, 351, 322, 298, 0, 0, 311, 359, 288, 361, 372, 310, 348, 306, 0, 0, 345, 337, 340, 354, 346, 345, 335, 330, 0, 0, 333, 330, 337, 343, 337, 336, 0, 327, 0, 0, 334, 345, 344, 335, 328, 345, 340, 335, 0, 0, 339, 340, 331, 326, 327, 326, 340, 336, 0, 0, 313, 322, 305, 308, 306, 305, 310, 310, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		'R': {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 514, 508, 512, 483, 516, 512, 535, 529, 0, 0, 534, 508, 535, 546, 534, 541, 513, 539, 0, 0, 498, 514, 507, 512, 524, 506, 504, 494, 0, 0, 0, 484, 495, 492, 497, 475, 470, 473, 0, 0, 451, 444, 463, 458, 466, 450, 433, 449, 0, 0, 437, 451, 437, 454, 454, 444, 453, 433, 0, 0, 426, 441, 448, 453, 450, 436, 435, 426, 0, 0, 449, 455, 461, 484, 477, 461, 448, 447, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		'Q': {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 935, 930, 921, 825, 998, 953, 1017, 955, 0, 0, 943, 961, 989, 919, 949, 1005, 986, 953, 0, 0, 927, 972, 961, 989, 1001, 992, 972, 931, 0, 0, 930, 913, 951, 946, 954, 949, 916, 923, 0, 0, 915, 914, 927, 924, 928, 919, 909, 907, 0, 0, 899, 923, 916, 918, 913, 918, 913, 902, 0, 0, 893, 911, 0, 910, 914, 914, 908, 891, 0, 0, 890, 899, 898, 916, 898, 893, 895, 887, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		'K': {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 60004, 60054, 60047, 59901, 59901, 60060, 60083, 59938, 0, 0, 59968, 60010, 60055, 60056, 60056, 60055, 60010, 60003, 0, 0, 59938, 60012, 59943, 60044, 59933, 60028, 60037, 59969, 0, 0, 59945, 60050, 60011, 59996, 59981, 60013, 0, 59951, 0, 0, 59945, 59957, 59948, 59972, 59949, 59953, 59992, 59950, 0, 0, 59953, 59958, 59957, 59921, 59936, 59968, 59971, 59968, 0, 0, 59996, 60003, 59986, 59950, 59943, 59982, 60013, 60004, 0, 0, 60017, 60030, 59997, 59986, 60006, 59999, 60040, 60018, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}

	i, j := m.from, m.to
	p, q := Piece(pos.board[i]), Piece(pos.board[j])
	score := pst[p][j] - pst[p][i]
	if q != '.' && q != ' ' && !q.ours() {
		score += pst[q.Flip()][j.Flip()]
	}
	// Castling check direction
	if abs(int(j-pos.kp)) < 2 {
		score += pst['K'][j.Flip()]
	}
	// Castling
	if p == 'K' && (abs(int(i-j)) == 2) {
		score = score + pst['R'][(i+j)/2]
		if j < i {
			score = score - pst['R'][A1]
		} else {
			score = score - pst['R'][H1]
		}
	}
	if p == 'P' {
		// Pawn promotion to queen
		if A8 <= j && j <= H8 {
			score += pst['Q'][j] - pst['P'][j]
		}
		// En-passant capture
		if j == pos.ep {
			score += pst['P'][(j + S).Flip()]
		}
	}
	return score
}

var (
	MateValue     = Piece('K').value() + 10*Piece('Q').value()
	MaxTableSize  = 10000000
	EvalRoughness = 13
)

type entry struct {
	depth int
	score int
	gamma int
	move  Move
}

// searcher is an recursive alpha-beta search algorithm
type Searcher struct {
	tp    map[Position]entry
	nodes int
}

// lowest score in position
func (s *Searcher) bound(pos Position, gamma, depth int) (int, Move) {
	s.nodes++
	if depth <= 0 {
		return pos.score, Move{}
	}

	var worstMove Move
	worstScore := gamma

	for _, m := range pos.Moves() {
		newPos := pos.Move(m)
		score, _ := s.bound(newPos, worstScore, depth-1)
		if score < worstScore { //look for lower scores
			worstScore = score
			worstMove = m
		}
	}
	return worstScore, worstMove
}

// iteratively searches for lowest score
func (s *Searcher) Search(pos Position, maxNodes int) Move {
	s.nodes = 0
	var worstMove Move
	worstScore := 3 * MateValue

	for depth := 1; depth < 99; depth++ {
		_, move := s.bound(pos, worstScore, depth)
		worstMove = move
		if s.nodes >= maxNodes {
			break
		}
	}
	return worstMove
}
