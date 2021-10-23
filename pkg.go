package nocopybuffer

import (
	"sync"
	"sync/atomic"
)

type Pkg struct {
	ref   int32
	block *NcBlock //not nil means reference any block , nil means does't
	data  []byte   //block not nil means this data share with other pkg, don't try update
}

var pkgPool *sync.Pool

func init() {
	pkgPool = &sync.Pool{
		New: func() interface{} {
			return &Pkg{}
		},
	}
}
func allocPkg() *Pkg {
	pkg := pkgPool.Get().(*Pkg)
	if pkg.ref != 0 {
		panic("allocPkg fail, pkg.ref != 0 ")
	}
	pkg.block = nil
	pkg.data = nil
	return pkg
}

func (pkg *Pkg) Hold() *Pkg {
	if pkg.block != nil {
		pkg.block.hold()
	}
	atomic.AddInt32(&pkg.ref, 1)
	return pkg
}

func (pkg *Pkg) Release() bool {
	if pkg.block != nil {
		pkg.block.release()
	}
	ref := atomic.AddInt32(&pkg.ref, -1)
	if ref == 0 {
		pkgPool.Put(pkg)
		return true
	}
	if ref < 0 {
		panic("pkg.ref < 0 ")
	}
	return false
}

//reference pkg data
func (pkg *Pkg) Bytes() []byte {
	return pkg.data
}

//pkg can update it's data when it does't reference any block underlay buf and pkg has only one owner .
func (pkg *Pkg) CanUpdate() bool {
	return pkg.block == nil && atomic.LoadInt32(&pkg.ref) == 1
}

//clone pkg data for update
func (pkg *Pkg) Clone() *Pkg {
	npkg := allocPkg()
	npkg.data = make([]byte, len(pkg.data))
	copy(npkg.data, pkg.data)
	npkg.Hold()
	return npkg
}

//copy pkg data for update
func (pkg *Pkg) CopyData() []byte {
	b := make([]byte, len(pkg.data))
	copy(b, pkg.data)
	return b
}
