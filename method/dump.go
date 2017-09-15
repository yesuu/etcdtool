package method

import (
	"context"
	"fmt"

	"etcdtool/entity"

	"github.com/coreos/etcd/client"
)

func init() {
	m["dump"] = Dump
}

func Dump(conf *entity.Conf) (string, error) {
	c, err := client.New(client.Config{
		Endpoints: []string{conf.Src},
	})
	if err != nil {
		return "", err
	}

	kapi := client.NewKeysAPI(c)
	resp, err := kapi.Get(context.Background(), "/", &client.GetOptions{
		Recursive: true,
		Sort:      true,
	})
	if err != nil {
		return "", err
	}

	return printNode(resp.Node), nil
}

const p = `{"key":"%s","value":"%s"}` + "\n"

func printNode(node *client.Node) string {
	if node.Dir {
		s := ""
		for _, v := range node.Nodes {
			s += printNode(v)
		}
		return s
	} else {
		return fmt.Sprintf(p, node.Key, node.Value)
	}
}
