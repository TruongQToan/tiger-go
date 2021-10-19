package tiger

import "fmt"

type Pos struct {
	fileName string
	line     int
	col      int
	byte     int
	length   int
}

func (p *Pos) String() string {
	return fmt.Sprintf("file: %s, line: %d, col: %d, byte: %d, length: %d", p.fileName, p.line, p.col, p.byte, p.length)
}
