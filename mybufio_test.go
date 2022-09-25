package nocopybuffer

import (
	"bytes"
	"reflect"
	"testing"
)

func TestGetPkg(t *testing.T) {
	buf := []byte("0123456789")
	rd := bytes.NewBuffer(buf)
	bl := NewBlockList(rd, NewPool(4))

	pkgLen := 6
	pkg, err := bl.GetPkg(pkgLen)
	if err != nil {
		t.Fatal(err)
	}

	if len(pkg.Bytes()) != pkgLen {
		t.Fatal("")
	}

	if !reflect.DeepEqual(pkg.Bytes(), buf[:pkgLen]) {
		panic(pkg.Bytes())
	}

	released := pkg.Release()
	if released != true {
		panic(released)
	}

	pkg2Len := 4
	pkg2, err := bl.GetPkg(pkg2Len)
	if err != nil {
		t.Fatal(err)
	}
	if len(pkg2.Bytes()) != pkg2Len {
		t.Fatal("")
	}

	if !reflect.DeepEqual(pkg2.Bytes(), buf[pkgLen:pkgLen+pkg2Len]) {
		t.Fatal(pkg2.Bytes())
	}

	released = pkg2.Release()
	if released != true {
		t.Fatalf("expect release ok, but :%v", released)
	}
}

func TestUpdatePkg(t *testing.T) {
	buf := []byte("0123456789")
	rd := bytes.NewBuffer(buf)
	bl := NewBlockList(rd, NewPool(4))

	pkgLen := 2
	pkg, err := bl.GetPkg(pkgLen)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(pkg.Bytes(), buf[:pkgLen]) {
		t.Fatal(pkg.Bytes())
	}

	data := pkg.Bytes()
	copy(data[:], []byte("xy"))

	if !reflect.DeepEqual(pkg.Bytes(), []byte("xy")) {
		t.Fatal(string(pkg.Bytes()))
	}
}

func TestClonePkg(t *testing.T) {
	buf := []byte("0123456789")
	rd := bytes.NewBuffer(buf)
	bl := NewBlockList(rd, NewPool(4))

	pkgLen := 2
	pkg, err := bl.GetPkg(pkgLen)
	if err != nil {
		t.Fatal(err)
	}

	//clone will alloc new pkg
	pkg2 := pkg.Clone()
	ok := pkg2.CanUpdateSafe()
	if !ok {
		t.Fatal(ok)
	}

	//update pkg2'data,
	data := pkg2.Bytes()
	data[0] = 'x'
	data[1] = 'y'

	if reflect.DeepEqual(pkg.Bytes(), pkg2.Bytes()) {
		t.Fatal("expect pkg'data is not same with pkg2'data")
	}

	pkg.Release()
	pkg2.Release()
}
