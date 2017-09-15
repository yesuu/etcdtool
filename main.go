package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"etcdtool/entity"
	"etcdtool/method"
)

func main() {
	if len(os.Args) == 1 {
		help()
		return
	}
	conf := new(entity.Conf)
	flag.BoolVar(&conf.Array, "s", false, "使用数组格式")
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
	fmt.Println("etcdtool dump http://localhost:2379 -")
	fmt.Println("etcdtool restore - http://localhost:2379")
}
