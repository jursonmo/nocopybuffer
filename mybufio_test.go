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

	pkg, err := bl.GetPkg(6)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(pkg.Bytes(), buf[:6]) {
		panic(pkg.Bytes())
	}

	released := pkg.Release()
	if released != true {
		panic(released)
	}

	pkg, err = bl.GetPkg(4)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(pkg.Bytes(), buf[6:6+4]) {
		t.Fatal(pkg.Bytes())
	}

	released = pkg.Release()
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

	//clone it
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
