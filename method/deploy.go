package method

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/yesuu/etcdtool/entity"

	"github.com/BurntSushi/toml"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	"github.com/coreos/etcd/mvcc/mvccpb"
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

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{conf.Src},
		DialTimeout: 6 * time.Second,
	})
	if err != nil {
		return "", err
	}
	defer cli.Close()

	for _, k := range createList {
		resp, err := cli.Put(context.Background(), k, local[k])
		if err != nil {
			return out, err
		}
		out += fmt.Sprintln(resp)
	}
	for _, k := range updateList {
		resp, err := clientv3.NewKV(cli).
			Txn(context.Background()).
			If(clientv3.Compare(clientv3.Value(k), "=", remote[k])).
			Then(clientv3.OpPut(k, local[k])).
			Commit()
		if err != nil {
			return out, err
		}
		out += fmt.Sprintln(resp)
	}
	for _, k := range deleteList {
		resp, err := clientv3.NewKV(cli).
			Txn(context.Background()).
			If(clientv3.Compare(clientv3.Value(k), "=", remote[k])).
			Then(clientv3.OpDelete(k, clientv3.WithPrefix())).
			Commit()
		if err != nil {
			return out, err
		}
		out += fmt.Sprintln(resp)
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
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{conf.Src},
		DialTimeout: 6 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	nodes := []*mvccpb.KeyValue{}
	for _, ctxUrl := range contextUrl {
		resp, err := cli.Get(context.Background(), ctxUrl,
			clientv3.WithPrefix(),
		)
		if err != nil {
			if err == rpctypes.ErrKeyNotFound {
				continue
			}
			return nil, err
		}
		nodes = append(nodes, resp.Kvs...)
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

func nodeToMap(nodes []*mvccpb.KeyValue) map[string]string {
	m := map[string]string{}
	for _, node := range nodes {
		m[string(node.Key)] = string(node.Value)
	}
	return m
}
