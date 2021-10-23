package nocopybuffer

import (
	"fmt"
	"sync"
)

type Pool interface {
	Put(interface{})
	PutNum() uint64
	Get() interface{}
	GetNum() uint64
	String() string
}
type builtinPool struct {
	sync.Pool
	putNum uint64
	getNum uint64
}

//use sync.Pool as built-in Block Pool
func NewPool(elemBufSize int) Pool {
	return &builtinPool{
		Pool: sync.Pool{
			New: func() interface{} {
				return &NcBlock{buf: make([]byte, elemBufSize)}
			},
		},
	}
}

func (p *builtinPool) Put(x interface{}) {
	p.Pool.Put(x)
	p.putNum++
}
func (p *builtinPool) Get() interface{} {
	x := p.Pool.Get()
	p.getNum++
	return x
}
func (p *builtinPool) PutNum() uint64 {
	return p.putNum
}
func (p *builtinPool) GetNum() uint64 {
	return p.getNum
}
func (p *builtinPool) String() string {
	return fmt.Sprintf("putNum:%d, getNum:%d", p.putNum, p.getNum)
}
