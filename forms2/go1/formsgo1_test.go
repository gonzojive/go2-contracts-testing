package sexpressionsgo1

import (
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var cmpOpts = []cmp.Option{
	cmp.Transformer("rowcol", func(rc rowCol) struct{ Row, Col int } {
		return struct{ Row, Col int }{rc.row.offset(), rc.col.offset()}
	}),
	cmp.Comparer(func(a, b error) bool { return a == b }),
	// cmp.Transformer("rowcol", func(rc rowCol) struct{ Row, Col int } {
	// 	return rc.String()
	// }),
}

func Test_strSourceFile(t *testing.T) {
	type args struct {
		co cursorOffset
	}
	tests := []struct {
		name      string
		want, got interface{}
	}{
		// linestart() tests
		{"line-starts 1", []cursorOffset{0, 3, 4}, mustSourceFile("", "ab\n\n").lineStarts()},
		{"line-starts empty file", []cursorOffset{0}, mustSourceFile("", "").lineStarts()},

		// linelength() tests
		{"line-lengths 1", []int{2, 0, 0}, mustSourceFile("", "ab\n\n").lineLengths()},
		{"line-lengths single line", []int{3}, mustSourceFile("", "abc").lineLengths()},

		// cursorOffset()
		{"cursorOffset_1", 0, mustSourceFile("", "abc").cursorOffset().int()},

		// offsetToRowCol() tests
		{"offsetToRowCol_1", rowCol{0, 1}, mustSourceFile("", "a\n").offsetToRowCol(1)},
		{"offsetToRowCol_2", rowCol{1, 0}, mustSourceFile("", "a\n").offsetToRowCol(2)},
		{
			"hello-world.go",
			rowCol{0, 0},
			mustSourceFile("hi.go", "0123456789\nabcdefghij\nXYZ").offsetToRowCol(0),
		},
		{
			"hello-world.go",
			rowCol{0, 1},
			mustSourceFile("hi.go", "0123456789\nabcdefghij\nXYZ").offsetToRowCol(1),
		},
		// read and unread runes
		{
			"peek-peek-read-read-unread",
			performOps(mustSourceFile("", "a"), peak, peak, read, read, unread, unread),
			[]opResult{
				{0, 'a', nil},
				{0, 'a', nil},
				{1, 'a', nil},
				{1, 0, io.EOF},
				{0, 0, nil},
				{0, 0, io.EOF},
			},
		},
		// {
		// 	"hello-world.go",
		// 	2,
		// 	func() interface{} {
		// 		sf := mustSourceFile("hi.go", "01")
		// 		sf.readRune()
		// 	}(),
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.want, tt.got, cmpOpts...); diff != "" {
				t.Errorf("unexpected diff:\n%s", diff)
			}
		})
	}
}

type op int

const (
	peak op = iota
	read
	unread
)

type opResult struct {
	Cursor cursorOffset
	Rune   rune
	Err    error
}

func performOps(sf sourceFile, ops ...op) []opResult {
	var ret []opResult
	for _, o := range ops {
		switch o {
		case read:
			r, err := sf.readRune()
			ret = append(ret, opResult{sf.cursorOffset(), r, err})
		case unread:
			err := sf.unreadRune()
			ret = append(ret, opResult{sf.cursorOffset(), 0, err})
		case peak:
			r, err := sf.peekRune()
			ret = append(ret, opResult{sf.cursorOffset(), r, err})
		default:
			panic(o)
		}
	}
	return ret
}

func Test_strSourceFile_roundtrip(t *testing.T) {
	type args struct {
		co cursorOffset
	}
	tests := []struct {
		name, code string
	}{
		// linestart() tests
		{"line-starts 1", "abc\n123\nsdfkl,sdkflksdflksdkfllk\n\n\n3533\n"},
		{"line-starts with unicode", "ひ°bc\n1ひ3\nsdfkl,sdkflksdflksdkfllk\n\n\n3533\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sf := mustSourceFile(tt.name, tt.code)
			runes := []rune(tt.code)
			for i := 0; i <= len(runes); i++ {
				offsetIn := cursorOffset(i)
				gotRC := sf.offsetToRowCol(offsetIn)
				gotOffset := sf.rowColToOffset(gotRC)
				t.Logf("%v -> %v -> %v", offsetIn, gotRC, gotOffset)
				if diff := cmp.Diff(offsetIn, gotOffset, cmpOpts...); diff != "" {
					t.Errorf("unexpected diff in roundtrip result:\n%s", diff)
				}
			}
		})
	}
}

func mustSourceFile(name, s string) *strSourceFile {
	sf, err := newStrSourceFile(name, s)
	if err != nil {
		panic(err)
	}
	return sf
}
