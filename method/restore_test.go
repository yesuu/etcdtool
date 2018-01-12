package method

import (
	"testing"
)

func Test_jsonToNodeArray(t *testing.T) {
	src := `[{"key":"/test/abc","value":"abc"},{"key":"/test/def","value":"def"}]`

	nodes, err := jsonToNodeArray([]byte(src))
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v", nodes)

	if len(nodes) != 2 {
		t.Fatalf("len(nodes): %d != 2", len(nodes))
	}
	if nodes[0].Key != "/test/abc" ||
		nodes[0].Value != "abc" ||
		nodes[1].Key != "/test/def" ||
		nodes[1].Value != "def" {
		t.Fatalf("%+v", nodes)
	}
}

func Test_jsonToNode(t *testing.T) {
	src := `{"key":"/test/abc","value":"abc"}` + "\n" +
		`{"key":"/test/def","value":"def"}` + "\n"

	nodes, err := jsonToNode([]byte(src))
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v", nodes)

	if len(nodes) != 2 {
		t.Fatalf("len(nodes): %d != 2", len(nodes))
	}
	if nodes[0].Key != "/test/abc" ||
		nodes[0].Value != "abc" ||
		nodes[1].Key != "/test/def" ||
		nodes[1].Value != "def" {
		t.Fatalf("%+v", nodes)
	}
}
