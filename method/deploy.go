package method

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/yesuu/etcdtool/entity"

	"github.com/BurntSushi/toml"
	"github.com/coreos/etcd/client"
)

func init() {
	m["deploy"] = Deploy
	m["up"] = Deploy
}

func Deploy(conf *entity.Conf) (string, error) {
	out := ""
	if conf.Fake {
		out += "> fake mode\n"
	}

	local, contextUrl, ignoreUrl, err := getLocal(conf)
	if err != nil {
		return out, err
	}
	remote, err := getRemote(conf, contextUrl, ignoreUrl)
	if err != nil {
		return out, err
	}

	createList, updateList, deleteList := []string{}, []string{}, []string{}
	for lk, lv := range local {
		rv, ok := remote[lk]
		if !ok {
			createList = append(createList, lk)
		} else if lv != rv {
			updateList = append(updateList, lk)
		}
	}
	for rk := range remote {
		_, ok := local[rk]
		if !ok {
			deleteList = append(deleteList, rk)
		}
	}
	if len(createList) == 0 && len(updateList) == 0 && len(deleteList) == 0 {
		out += "> 无事可做\n"
		return out, nil
	}

	createStr := "> 添加 %s\n\t\"%s\"\n"
	updateStr := "> 更新 %s\n\t\"%s\" => \"%s\"\n"
	deleteStr := "> 删除 %s\n\t\"%s\"\n"
	for _, k := range createList {
		out += fmt.Sprintf(createStr, k, local[k])
	}
	for _, k := range updateList {
		out += fmt.Sprintf(updateStr, k, remote[k], local[k])
	}
	for _, k := range deleteList {
		out += fmt.Sprintf(deleteStr, k, remote[k])
	}

	if conf.Fake {
		return out, nil
	}

	c, err := client.New(client.Config{
		Endpoints: []string{conf.Src},
	})
	if err != nil {
		return "", err
	}
	kapi := client.NewKeysAPI(c)

	for _, k := range createList {
		resp, err := kapi.Create(context.Background(), k, local[k])
		if err != nil {
			return out, err
		}
		respJson, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return out, err
		}
		out += string(respJson) + "\n"
	}
	for _, k := range updateList {
		resp, err := kapi.Update(context.Background(), k, local[k])
		if err != nil {
			return out, err
		}
		respJson, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return out, err
		}
		out += string(respJson) + "\n"
	}
	for _, k := range deleteList {
		resp, err := kapi.Delete(context.Background(), k, &client.DeleteOptions{
			PrevValue: remote[k],
			Recursive: true,
		})
		if err != nil {
			return out, err
		}
		respJson, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return out, err
		}
		out += string(respJson) + "\n"
	}

	out += "> ok\n"

	return out, nil
}

// return local, contextUrl, ignoreUrl, err
func getLocal(conf *entity.Conf) (map[string]string, []string, []string, error) {
	local := map[string]string{}
	if conf.Dest == "-" {
		in, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return nil, nil, nil, err
		}
		if conf.Toml {
			err := toml.Unmarshal(in, &local)
			if err != nil {
				return nil, nil, nil, err
			}
		} else {
			return nil, nil, nil, errors.New("只支持 toml 格式")
		}
	} else {
		in, err := ioutil.ReadFile(conf.Dest)
		if err != nil {
			return nil, nil, nil, err
		}
		ext := path.Ext(conf.Dest)
		if conf.Toml || ext == ".toml" {
			err := toml.Unmarshal(in, &local)
			if err != nil {
				return nil, nil, nil, err
			}
		} else {
			return nil, nil, nil, errors.New("只支持 toml 格式")
		}
	}

	contextUrl, ignoreUrl := []string{"/"}, []string{}
	if local["context"] != "" {
		contextUrl = strings.Split(local["context"], ",")
	}
	if local["ignore"] != "" {
		ignoreUrl = strings.Split(local["ignore"], ",")
	}
	delete(local, "context")
	delete(local, "ignore")
	for k := range local {
		del := true
		for _, ctxUrl := range contextUrl {
			if strings.HasPrefix(k, ctxUrl) {
				del = false
			}
		}
		if del {
			delete(local, k)
		}
	}
	for _, ignUrl := range ignoreUrl {
		for k := range local {
			if strings.HasPrefix(k, ignUrl) {
				delete(local, k)
			}
		}
	}
	return local, contextUrl, ignoreUrl, nil
}

func getRemote(conf *entity.Conf, contextUrl, ignoreUrl []string) (map[string]string, error) {
	c, err := client.New(client.Config{
		Endpoints: []string{conf.Src},
	})
	if err != nil {
		return nil, err
	}

	kapi := client.NewKeysAPI(c)

	nodes := client.Nodes{}
	for _, ctxUrl := range contextUrl {
		resp, err := kapi.Get(context.Background(), ctxUrl, &client.GetOptions{
			Recursive: true,
			Sort:      true,
		})
		if err != nil {
			err1, ok := err.(client.Error)
			if ok && err1.Code == client.ErrorCodeKeyNotFound {
				continue
			}
			return nil, err
		}
		nodes = append(nodes, resp.Node)
	}

	m := nodeToMap(nodes)
	for _, ignUrl := range ignoreUrl {
		for k := range m {
			if strings.HasPrefix(k, ignUrl) {
				delete(m, k)
			}
		}
	}
	return m, nil
}

func nodeToMap(nodes client.Nodes) map[string]string {
	m := map[string]string{}
	for _, node := range nodes {
		nodeToMapF(node, m)
	}
	return m
}

func nodeToMapF(node *client.Node, m map[string]string) {
	if node.Dir {
		for _, v := range node.Nodes {
			nodeToMapF(v, m)
		}
	} else {
		m[node.Key] = node.Value
	}
}
