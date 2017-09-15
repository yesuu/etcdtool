package method

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"etcdtool/entity"

	"github.com/coreos/etcd/client"
)

func init() {
	m["restore"] = Restore
}

func Restore(conf *entity.Conf) (string, error) {
	in, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}

	var src client.Nodes
	if conf.Array {
		src, err = strToNode(in)
		if err != nil {
			return "", err
		}
	} else {
		src, err = strToNode2(in)
		if err != nil {
			return "", err
		}
	}

	c, err := client.New(client.Config{
		Endpoints: []string{conf.Dest},
	})
	if err != nil {
		return "", err
	}

	kapi := client.NewKeysAPI(c)

	const p = `{"key":"%s","value":"%s","prevNode":{"key":"%s","value":"%s"}}` + "\n"
	const pn = `{"key":"%s","value":"%s","prevNode":null}` + "\n"
	s := ""
	for _, v := range src {
		resp, err := kapi.Set(context.Background(), v.Key, v.Value, nil)
		if err != nil {
			return "", err
		}
		if resp.PrevNode == nil {
			s += fmt.Sprintf(pn, resp.Node.Key, resp.Node.Value)
		} else if resp.Node.Key != resp.PrevNode.Key ||
			resp.Node.Value != resp.PrevNode.Value {
			s += fmt.Sprintf(p,
				resp.Node.Key, resp.Node.Value,
				resp.PrevNode.Key, resp.PrevNode.Value)
		}
	}

	return s, nil
}

// src 为 [{},{}]
func strToNode(src []byte) (client.Nodes, error) {
	nodes := new(client.Nodes)
	err := json.Unmarshal(src, nodes)
	return *nodes, err
}

// src 为 "{}\n{}\n"
func strToNode2(src []byte) (client.Nodes, error) {
	result := client.Nodes{}
	for _, v := range bytes.Split(bytes.TrimSpace(src), []byte("\n")) {
		node := new(client.Node)
		err := json.Unmarshal(v, node)
		if err != nil {
			return nil, err
		}
		result = append(result, node)
	}
	return result, nil
}
