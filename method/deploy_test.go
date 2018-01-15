package method

import (
	"testing"

	"github.com/coreos/etcd/client"
)

func TestNodeToMap(t *testing.T) {
	m := nodeToMap(client.Nodes{
		&client.Node{
			Key:   "/value",
			Value: "val",
		},
		&client.Node{
			Key: "/abc",
			Dir: true,
			Nodes: []*client.Node{
				&client.Node{
					Key: "/abc/abc",
					Dir: true,
					Nodes: []*client.Node{
						&client.Node{
							Key:   "/abc/abc/value",
							Value: "val",
						},
					},
				},
				&client.Node{
					Key:   "/abc/value",
					Value: "val",
				},
			},
		},
	})
	t.Logf("%+v", m)

	if len(m) != 3 {
		t.Error("长度不对 ", len(m))
	}
	if m["/value"] != "val" {
		t.Error("值不对 ", m["/value"])
	}
	if m["/abc/value"] != "val" {
		t.Error("值不对 ", m["/abc/value"])
	}
	if m["/abc/abc/value"] != "val" {
		t.Error("值不对 ", m["/abc/abc/value"])
	}
}
