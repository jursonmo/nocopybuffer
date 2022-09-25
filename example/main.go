package main

//TODO: test case
import (
	"fmt"
	"reflect"
	"strings"

	"github.com/jursonmo/nocopybuffer"
)

func main() {
	test()
}

func test() {
	rd := strings.NewReader("0123456789abcdefg")
	bl := nocopybuffer.NewBlockList(rd, nocopybuffer.NewPool(4))
	//------------ 1. get pkg , from two under block -----------------
	pkg, err := bl.GetPkg(6)
	if err != nil {
		return
	}
	fmt.Printf("pkg data:%s\n", string(pkg.Bytes()))
	if !reflect.DeepEqual(pkg.Bytes(), []byte("012345")) {
		panic(pkg.Bytes())
	}

	released := pkg.Release()
	if released != true {
		panic(released)
	}
	fmt.Printf("blocklist:%v\n", bl)

	//------------2. get pkg and hold by another owner ---------------
	pkg, err = bl.GetPkg(2)
	if err != nil {
		return
	}
	if !reflect.DeepEqual(pkg.Bytes(), []byte("67")) {
		panic(pkg.Bytes())
	}
	fmt.Printf("blocklist:%v\n", bl)
	fmt.Printf("pkg data:%s\n", string(pkg.Bytes()))
	pkg2 := pkg.Hold()

	//check pkg2'data
	if !reflect.DeepEqual(pkg2.Bytes(), []byte("67")) {
		panic(pkg.Bytes())
	}

	data := pkg.Bytes()
	data[0] = 'x'
	data[1] = 'y'

	//check pkg2'data
	if !reflect.DeepEqual(pkg2.Bytes(), []byte("xy")) {
		panic(pkg.Bytes())
	}

	released = pkg.Release() //just dec the reference
	if released == true {
		panic(released)
	}

	released = pkg2.Release() // release pkg  and put back to the pkgPool
	if released != true {
		panic(released)
	}

	//--------- 3. get pkg and check pkg clone --------------------------------
	pkg, err = bl.GetPkg(2)
	if err != nil {
		fmt.Printf("blocklist:%v\n", bl)
		panic(err)
	}
	if !reflect.DeepEqual(pkg.Bytes(), []byte("89")) {
		panic(pkg.Bytes())
	}

	//clone a pkg for update
	pkg3 := pkg.Clone()
	canupdatesafe := pkg3.CanUpdateSafe()
	if !canupdatesafe {
		panic("expected: we can update pkg3")
	}
	data = pkg3.Bytes()
	data[0] = 'x' //change pkg3'data, but don't change pkg'data
	data[1] = 'y'

	//check pkg'data still is "89"
	if !reflect.DeepEqual(pkg.Bytes(), []byte("89")) {
		panic(pkg.Bytes())
	}

	fmt.Printf("pkg data:%s\n", string(pkg.Bytes()))
	released = pkg.Release()
	if released == false {
		panic(released)
	}
	fmt.Printf("blocklist:%v\n", bl)

	//------4. ----------------

}
