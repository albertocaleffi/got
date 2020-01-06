package got_test

import (
	"bytes"
	"github.com/albertocaleffi/got"
	"reflect"
	"testing"
)

func samePos(pos got.Pos, path string, lineNo int) bool {
	return reflect.DeepEqual(pos, got.Pos{Path: "tpl.got", LineNo: lineNo})
}

func TestScanSingleTextBlock(t *testing.T) {
	cases := []struct {
		in, got string
	}{
		{"hello, world", "hello, world"},
		{"<", "<"},
		{">", ">"},
		{"<>", "<>"},
		{"<< >", "<< >"},
		{"<< %>", "<< %>"},
		{"<%% %>", "<%"},
		{"<html><title>OK</title></html>", "<html><title>OK</title></html>"},
	}

	for _, c := range cases {
		s := got.NewScanner(bytes.NewBufferString(c.in), "tpl.got")
		if blk, err := s.Scan(); err != nil {
			t.Fatal(err)
		} else if blk, ok := blk.(*got.TextBlock); !ok {
			t.Fatalf("unexpected block type: %T", blk)
		} else if blk.Content != c.got {
			t.Fatalf("expected %q, got %s", c.got, blk.Content)
		} else if !samePos(blk.Pos, "tpl.got", 1) {
			t.Fatalf("unexpected pos: %#v", blk.Pos)
		}
	}
}

func TestScanSingleCodeBlock(t *testing.T) {
	cases := []struct {
		in, got string
	}{
		{"<% x := 1 %>", " x := 1 "},
		{"<% x := \"start\" %>", " x := \"start\" "},
		{"<% set(x int) %>EXTRA", " set(x int) "},
		{"<% html(w io.Writer, s string) %> extra", " html(w io.Writer, s string) "},
		{`<% text(w, "%>") %>`, ` text(w, "`},
		{`<% text(w, "%%>") %>`, ` text(w, "%>") `},
	}

	for _, c := range cases {
		s := got.NewScanner(bytes.NewBufferString(c.in), "tpl.got")
		if blk, err := s.Scan(); err != nil {
			t.Fatal(err)
		} else if blk, ok := blk.(*got.CodeBlock); !ok {
			t.Fatalf("unexpected block type: %T", blk)
		} else if blk.Content != c.got {
			t.Fatalf("expected %q, got %q", c.got, blk.Content)
		} else if !samePos(blk.Pos, "tpl.got", 1) {
			t.Fatalf("unexpected pos: %#v", blk.Pos)
		}
	}
}

func TestScanUnexpectedEOF(t *testing.T) {
	cases := []string{
		"<%",
		"<% x = 1 ",
		"<% x = 2 %",
		"<% x = 2 % ",
		"<% x = 2 % >",
	}

	want := "Expected close tag, found EOF at tpl.got:1"
	for _, c := range cases {
		s := got.NewScanner(bytes.NewBufferString(c), "tpl.got")
		if _, err := s.Scan(); err == nil || err.Error() != want {
			t.Fatalf("unexpected error: %s", err)
		}
	}
}

func TestScanMultiLineTextBlocks(t *testing.T) {
	c := "hello\nworld<% x := 1 %>\n\ngoodbye"
	s := got.NewScanner(bytes.NewBufferString(c), "tpl.got")

	if blk, err := s.Scan(); err != nil {
		t.Fatal("unexpected EOF(0)")
	} else if blk, ok := blk.(*got.TextBlock); !ok {
		t.Fatalf("unexpected block type(0): %T", blk)
	} else if blk.Content != "hello\nworld" {
		t.Fatalf("unexpected content(1): %s", blk.Content)
	} else if pos := got.Position(blk); !samePos(pos, "tpl.got", 1) {
		t.Fatalf("unexpected pos(0): %#v", pos)
	}

	if blk, err := s.Scan(); err != nil {
		t.Fatal("unexpected EOF(1)")
	} else if blk, ok := blk.(*got.CodeBlock); !ok {
		t.Fatalf("unexpected block type(0): %T", blk)
	} else if blk.Content != " x := 1 " {
		t.Fatalf("unexpected content(1): %s", blk.Content)
	} else if pos := got.Position(blk); !samePos(pos, "tpl.got", 2) {
		t.Fatalf("unexpected pos(1): %#v", pos)
	}

	if blk, err := s.Scan(); err != nil {
		t.Fatal("unexpected EOF(2)")
	} else if blk, ok := blk.(*got.TextBlock); !ok {
		t.Fatalf("unexpected block type(2): %T", blk)
	} else if blk.Content != "\n\ngoodbye" {
		t.Fatalf("unexpected content(1): %s", blk.Content)
	} else if pos := got.Position(blk); !samePos(pos, "tpl.got", 3) {
		t.Fatalf("unexpected pos(2): %#v", pos)
	}
}
