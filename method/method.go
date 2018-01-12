package method

import (
	"errors"

	"github.com/yesuu/etcdtool/entity"
)

var m = map[string]func(*entity.Conf) (string, error){}

func Method(name string) func(*entity.Conf) (string, error) {
	mm, ok := m[name]
	if !ok {
		return func(conf *entity.Conf) (string, error) {
			return "", errors.New(name + " 命令不存在")
		}
	}
	return mm
}
