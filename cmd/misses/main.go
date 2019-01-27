package main

import (
	"fmt"
	"io/ioutil"
	"jcheng/bexp/app"
	"os"
	"path"
)

func main() {


	fmt.Println(os.Args)
	infos, err := ioutil.ReadDir(path.Join( path.Dir(os.Args[0])))
	if err != nil {
		fmt.Println(err)
	} else {
		for _, info := range infos {
			fmt.Println(info.Name())

		}
	}

	fmt.Println(app.RelPath("misses/main.go"))

}
