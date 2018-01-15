package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/yesuu/etcdtool/entity"
	"github.com/yesuu/etcdtool/method"
)

func main() {
	if len(os.Args) == 1 {
		help()
		return
	}
	conf := new(entity.Conf)
	flag.BoolVar(&conf.Fake, "fake", false, "只输出执行计划，不实际执行")
	flag.BoolVar(&conf.Toml, "toml", false, "使用 toml 输入输出")
	flag.BoolVar(&conf.Json, "json", false, "使用 json 输入输出")
	flag.BoolVar(&conf.Array, "s", false, "使用数组格式，存在 json 选项时有效")
	flag.Parse()
	conf.Src = flag.Arg(1)
	conf.Dest = flag.Arg(2)
	s, err := method.Method(flag.Arg(0))(conf)
	fmt.Print(s)
	if err != nil {
		log.Fatal(err)
	}
}

func help() {
	fmt.Println("etcdtool dump http://localhost:2379 dump.toml")
	fmt.Println("etcdtool restore dump.toml http://localhost:2379")
	fmt.Println("etcdtool up http://localhost:2379 up.toml")
}
