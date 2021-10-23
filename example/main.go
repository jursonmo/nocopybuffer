package main

import (
	"fmt"
	"learn/pkg/nocopybuffer"
	"strings"
)

func main() {
	test()
}

func test() {
	rd := strings.NewReader("0123456789")
	bl := nocopybuffer.NewBlockList(rd, nocopybuffer.NewPool(4))
	//1. get pkg
	pkg, err := bl.GetPkg(6)
	if err != nil {
		return
	}
	fmt.Printf("pkg data:%s\n", string(pkg.Bytes()))
	pkg.Release()
	fmt.Printf("blocklist:%v\n", bl)
	//2. get pkg
	pkg, err = bl.GetPkg(2)
	if err != nil {
		return
	}
	fmt.Printf("blocklist:%v\n", bl)
	fmt.Printf("pkg data:%s\n", string(pkg.Bytes()))
	pkg.Release()

	//3. get pkg
	pkg, err = bl.GetPkg(2)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("blocklist:%v\n", bl)
		return
	}
	fmt.Printf("pkg data:%s\n", string(pkg.Bytes()))
	pkg.Release()
	fmt.Printf("blocklist:%v\n", bl)
}
