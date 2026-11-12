package main

import (
	"strings"
	"testing"
)

func TestRenderUsesTheRealRouter(t *testing.T) {
	got, err := render()
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`{"provider":"fake","status":"ok"} [HTTP 200]`,
		`{"url":"https://hello-demo-ew.a.run.app"} [HTTP 200]`,
		`{"error":"service not found"} [HTTP 404]`,
		`{"error":"invalid service name"} [HTTP 400]`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("render() lacks %q:\n%s", want, got)
		}
	}
}

func TestReplaceBlock(t *testing.T) {
	doc := "before\n" + begin + "\nold\n" + end + "\nafter\n"
	got, err := replaceBlock(doc, "new")
	if err != nil {
		t.Fatal(err)
	}
	if want := "before\n" + begin + "\nnew\n" + end + "\nafter\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if _, err := replaceBlock("no markers", "x"); err == nil {
		t.Error("a document without markers must fail")
	}
	if _, err := replaceBlock(end+"\n"+begin, "x"); err == nil {
		t.Error("markers in the wrong order must fail")
	}
}
