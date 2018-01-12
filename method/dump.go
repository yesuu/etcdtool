package method

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/yesuu/etcdtool/entity"

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

	if conf.Dest == "-" {
		if conf.Toml {
			return nodeToToml(resp.Node), nil
		}
		if conf.Json {
			return nodeToJson(resp.Node), nil
		}
	} else {
		ext := path.Ext(conf.Dest)
		if conf.Toml || ext == ".toml" {
			return "", writeFile(conf.Dest, nodeToToml(resp.Node))
		}
		if conf.Json || ext == ".json" {
			return "", writeFile(conf.Dest, nodeToJson(resp.Node))
		}
	}
	return "", errors.New("无法自动判断储存格式")
}

const (
	tomlF = `"%s"="%s"` + "\n"
	jsonF = `{"key":"%s","value":"%s"}` + "\n"
)

func nodeToToml(node *client.Node) string {
	if node.Dir {
		s := ""
		for _, v := range node.Nodes {
			s += nodeToToml(v)
		}
		return s
	} else {
		return fmt.Sprintf(tomlF, node.Key, node.Value)
	}
}

func nodeToJson(node *client.Node) string {
	if node.Dir {
		s := ""
		for _, v := range node.Nodes {
			s += nodeToJson(v)
		}
		return s
	} else {
		return fmt.Sprintf(jsonF, node.Key, node.Value)
	}
}

func writeFile(name, data string) error {
	return ioutil.WriteFile(name, []byte(data), 0644)
}
