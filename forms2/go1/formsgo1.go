package sexpressionsgo1

import (
	"fmt"
	"io"
)

type sourceFile interface {
	fileName() string
	readRune() (rune, error)
	peekRune() (rune, error)
	unreadRune() error
	cursorOffset() cursorOffset
	offsetToRowCol(cursorOffset) rowCol
	rowColToOffset(rc rowCol) cursorOffset
}

type cursorOffset int

type lineNum int

type colNum int

type rowCol struct {
	row lineNum
	col colNum
}

func invalidRowCol() rowCol { return rowCol{-1, -1} }

const (
	invalidCursorOffset = -1
)

func (n lineNum) offset() int { return int(n) }

func (n colNum) offset() int { return int(n) }

func (rc rowCol) String() string {
	return fmt.Sprintf("%d:%d", rc.row.offset()+1, rc.col.offset()+1)
}

func (co cursorOffset) int() int { return int(co) }

type strSourceFile struct {
	name           string
	cursor         cursorOffset
	runes          []rune
	newlineOffsets []cursorOffset
}

func newStrSourceFile(name, code string) (*strSourceFile, error) {
	runes := []rune(code)
	var newlineOffsets []cursorOffset
	for i, r := range runes {
		if r == '\n' {
			newlineOffsets = append(newlineOffsets, cursorOffset(i))
		}
	}
	return &strSourceFile{name, 0, runes, newlineOffsets}, nil
}

func (sf *strSourceFile) fileName() string {
	return sf.name
}
func (sf *strSourceFile) readRune() (rune, error) {
	if int(sf.cursor) == len(sf.runes) {
		return 0, io.EOF
	}
	r := sf.runes[sf.cursor]
	sf.cursor++
	return r, nil
}

func (sf *strSourceFile) peekRune() (rune, error) {
	if int(sf.cursor) == len(sf.runes) {
		return 0, io.EOF
	}
	r := sf.runes[sf.cursor]
	return r, nil
}

func (sf *strSourceFile) unreadRune() error {
	if sf.cursor == 0 {
		return io.EOF
	}
	sf.cursor--
	return nil
}

func (sf *strSourceFile) cursorOffset() cursorOffset {
	return sf.cursor
}

func (sf *strSourceFile) offsetToRowCol(co cursorOffset) rowCol {
	if co < 0 || int(co) > len(sf.runes) {
		return invalidRowCol()
	}
	for i, endOfLineOffset := range sf.newlineOffsets {
		line := lineNum(i)
		if co > endOfLineOffset {
			continue
		}
		return rowCol{line, colNum(co - sf.lineStart(line))}
	}
	lastLine := lineNum(len(sf.newlineOffsets))
	return rowCol{
		lastLine,
		colNum(co - sf.lineStart(lastLine)),
	}
}

func (sf *strSourceFile) rowColToOffset(rc rowCol) cursorOffset {
	start := sf.lineStart(rc.row)
	if start == invalidCursorOffset {
		return invalidCursorOffset
	}
	ll := sf.lineLength(rc.row)
	if rc.col.offset() > ll {
		return invalidCursorOffset
	}
	return cursorOffset(int(start) + rc.col.offset())
}

func (sf *strSourceFile) lineLength(l lineNum) int {
	start := sf.lineStart(l)
	if start == invalidCursorOffset {
		return -1
	}
	nextStart := sf.lineStart(l + 1)
	if nextStart == invalidCursorOffset {
		return len(sf.runes) - start.int()
	}
	return nextStart.int() - start.int() - 1
}

func (sf *strSourceFile) lineLengths() []int {
	var ret []int
	for i := lineNum(0); i.offset() <= len(sf.newlineOffsets); i++ {
		ret = append(ret, sf.lineLength(i))
	}
	return ret
}

func (sf *strSourceFile) lineStart(l lineNum) cursorOffset {
	if l.offset() == 0 {
		return 0
	}
	if l.offset()-1 >= len(sf.newlineOffsets) || l.offset() < 0 {
		return invalidCursorOffset
	}
	return sf.newlineOffsets[l.offset()-1] + 1
}

func (sf *strSourceFile) lineStarts() []cursorOffset {
	var ret []cursorOffset
	for i := lineNum(0); i.offset() <= len(sf.newlineOffsets); i++ {
		ret = append(ret, sf.lineStart(i))
	}
	return ret
}

type FormReader struct {
}
