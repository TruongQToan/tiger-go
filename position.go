package main

import "fmt"

type Pos struct {
	fileName string
	line     int
	col      int
}

func (p *Pos) String() string {
	return fmt.Sprintf("file: %s, line: %d, col: %d", p.fileName, p.line, p.col)
}
