package method

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/yesuu/etcdtool/entity"

	"github.com/BurntSushi/toml"
	"github.com/coreos/etcd/client"
)

func init() {
	m["restore"] = Restore
}

func Restore(conf *entity.Conf) (string, error) {
	var in []byte
	var err error
	var src client.Nodes
	if conf.Src == "-" {
		in, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		if conf.Toml {
			src, err = tomlToNode(in)
			if err != nil {
				return "", err
			}
		} else if conf.Json && conf.Array {
			src, err = jsonToNodeArray(in)
			if err != nil {
				return "", err
			}
		} else if conf.Json && !conf.Array {
			src, err = jsonToNode(in)
			if err != nil {
				return "", err
			}
		} else {
			return "", errors.New("无法自动判断读取格式")
		}
	} else {
		in, err = ioutil.ReadFile(conf.Src)
		if err != nil {
			return "", err
		}
		ext := path.Ext(conf.Src)
		if conf.Toml || ext == ".toml" {
			src, err = tomlToNode(in)
			if err != nil {
				return "", err
			}
		} else if (conf.Json || ext == ".json") && conf.Array {
			src, err = jsonToNodeArray(in)
			if err != nil {
				return "", err
			}
		} else if (conf.Json || ext == ".json") && !conf.Array {
			src, err = jsonToNode(in)
			if err != nil {
				return "", err
			}
		} else {
			return "", errors.New("无法自动判断读取格式")
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

func tomlToNode(src []byte) (client.Nodes, error) {
	nodes := map[string]string{}
	err := toml.Unmarshal(src, &nodes)
	if err != nil {
		return nil, err
	}
	result := client.Nodes{}
	for k, v := range nodes {
		result = append(result, &client.Node{
			Key:   k,
			Value: v,
		})
	}
	return result, nil
}

// src 为 [{},{}]
func jsonToNodeArray(src []byte) (client.Nodes, error) {
	nodes := new(client.Nodes)
	err := json.Unmarshal(src, nodes)
	return *nodes, err
}

// src 为 "{}\n{}\n"
func jsonToNode(src []byte) (client.Nodes, error) {
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
