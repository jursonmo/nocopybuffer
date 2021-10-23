package nocopybuffer

import (
	"fmt"
	"io"
	"sync/atomic"
)

const (
	blockSize = 8092
)

type BlockList struct {
	head      *NcBlock
	tail      *NcBlock
	free      *NcBlock
	rd        io.Reader
	err       error
	blockPool Pool
	stat
}
type stat struct {
	copyNum     uint64 //需要跨block 取数据而发生copy 的次数
	usedBlocks  uint64 //轮询用过多少给blocks(实际就是NewNcBlock 的调用次数)
	allocBlocks uint64 //how many block have allocated from Pool
	allocPkgs   uint64 //how many block have generated
}

//no copy block
type NcBlock struct {
	pool Pool
	ref  int32
	next *NcBlock
	buf  []byte
	size int32
	r, w int32
}

func (bl *BlockList) NewNcBlock() *NcBlock {
	ncb := bl.free
	if ncb == nil {
		ncb = bl.blockPool.Get().(*NcBlock)
		bl.allocBlocks++
	} else {
		bl.free = nil //reset free point
	}
	ncb.Init(bl.blockPool)
	bl.usedBlocks++
	return ncb
}

func (ncb *NcBlock) Init(p Pool) {
	if ncb.ref != 0 {
		panic("ncb.ref != 0")
	}
	ncb.hold()
	ncb.pool = p
	ncb.next = nil
	ncb.size = int32(len(ncb.buf))
	ncb.r, ncb.w = 0, 0
}

func (ncb *NcBlock) String() string {
	return fmt.Sprintf("ref:%d, r:%d, w:%d, size:%d, isTail:%v",
		atomic.LoadInt32(&ncb.ref), ncb.r, ncb.w, ncb.size, ncb.next == nil)
}

func (ncb *NcBlock) Buffered() int {
	return int(ncb.w - ncb.r)
}

func (ncb *NcBlock) hold() {
	atomic.AddInt32(&ncb.ref, 1)
}

func (ncb *NcBlock) release() bool {
	ref := atomic.AddInt32(&ncb.ref, -1)
	if ref == 0 {
		ncb.pool.Put(ncb)
		return true
	}
	//check
	if ref < 0 {
		panic("ref <0")
	}
	return false
}
func (ncb *NcBlock) canRelease() bool {
	return atomic.AddInt32(&ncb.ref, -1) == 0
}

type BlockListOpt struct {
	blockPool    Pool
	blockBufSize int
}

func NewBlockList(rd io.Reader, pool Pool) *BlockList {
	if pool == nil {
		pool = NewPool(blockSize)
	}
	return &BlockList{rd: rd, blockPool: pool}
}

func (bl *BlockList) Buffered() int {
	sum := 0
	for cursor := bl.head; cursor != nil; cursor = cursor.next {
		sum += cursor.Buffered()
	}
	return sum
}

//check if BlockList have n bytes buffered data
func (bl *BlockList) have(n int) bool {
	sum := 0
	for cursor := bl.head; cursor != nil; cursor = cursor.next {
		sum += cursor.Buffered()
		if sum > n {
			return true
		}
	}
	return false
}

//must fill delta bytes before return
func (bl *BlockList) fill(n int) error {
	if bl.err != nil {
		return bl.err
	}

	var err error
	rn, sum := 0, 0
	ncb := bl.tail
	if ncb == nil {
		//check
		if bl.head != nil {
			panic("bl tail is nil, but head not nil")
		}
		bl.tail = bl.NewNcBlock()
		bl.head = bl.tail
		ncb = bl.tail
	}
	for sum < n {
		if ncb.w == ncb.size { //this block is full, the BlockList should add a new block
			ncb = bl.NewNcBlock()
			bl.tail.next = ncb
			bl.tail = ncb
		}
		rn, err = bl.rd.Read(ncb.buf[ncb.w:])
		if rn < 0 {
			panic("errNegativeRead")
		}
		sum += rn
		ncb.w += int32(rn)
		if err != nil {
			bl.err = err
			return err
		}
	}
	return err
}

//获取指定大小的业务层报文
func (bl *BlockList) GetPkg(n int) (*Pkg, error) {
	//check if blockList have n bytes data
	if !bl.have(n) {
		if err := bl.fill(n - bl.Buffered()); err != nil {
			return nil, err
		}
	}

	pkg := allocPkg()
	bl.allocPkgs++
	//be here, means blocklist have enough data for pkg required
	//if head blokc have have enough data for pkg required, pkg data just references head block underlay buf, don't copy
	head := bl.head
	if head.Buffered() >= n {
		pkg.data = head.buf[head.r : int(head.r)+n]
		pkg.block = head
		pkg.Hold() //hold pkg self and hold block
		head.r += int32(n)
		bl.headMove() //try to move head
		return pkg, nil
	}

	//到这里,表示需要跨block 取数据，那么就用copy 方式, 不需要递增block的引用计数
	pkg.data = make([]byte, n) //TODO:use byteBufferPool
	for copySum, cn := 0, 0; copySum < n; copySum += cn {
		cn = copy(pkg.data[cn:], bl.head.buf[bl.head.r:bl.head.w])
		fmt.Printf("r:%d, w:%d, size:%d, copy:%d\n", bl.head.r, bl.head.w, bl.head.size, cn)
		bl.head.r += int32(cn)
		bl.headMove()
	}
	pkg.Hold() //just hold pkg self
	bl.copyNum++
	return pkg, nil
}

//if the head block's data have been Read , move head to next block
func (bl *BlockList) headMove() {
	if bl.head.r != bl.head.size {
		return
	}
	headOld := bl.head
	if bl.head == bl.tail {
		bl.head = nil
		bl.tail = nil
	} else {
		bl.head = bl.head.next
	}

	if bl.free == nil {
		if headOld.canRelease() { //true说明没有被业务层的Pkg引用(由于跨block 读取数据), 此head block 可以立即释放，
			bl.free = headOld //保存到free, 不需要放回到block Pool, 下次分配时，直接从bl.free 上取得block 直接用
		}
	} else {
		headOld.release()
	}
}

func (bl *BlockList) String() string {
	blocks := 0
	blocksStr := ""
	for cursor := bl.head; cursor != nil; cursor = cursor.next {
		blocksStr += cursor.String() + "\n"
		blocks++
	}
	bls := fmt.Sprintf("useBlocks:%d, allocBlocks:%d, allocPkgs:%d, copyNum:%d, err:%v, blocks:%d; blockPool:%v\n",
		bl.usedBlocks, bl.allocBlocks, bl.allocPkgs, bl.copyNum, bl.err, blocks, bl.blockPool)
	return bls + blocksStr
}
