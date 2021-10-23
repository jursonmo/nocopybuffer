
 当从socket read 读取数据时，如果使用Read(p []byte) 时，不好确定p的大小，如果按业务数据大小来读取数据，每执行一次read(p []byte),只会读取一个业务报文，读取多个报文就需要多次调用Read(p []byte),这样会导致过多的系统调用，影响性能。往往通过使用bufio.Reader 一次性读取很多个业务报文缓存在一个相对大的底层buf里,这样相对减少系统调用次数，然后业务代码再调用bufio.Read(p []byte) 时，其实是拷贝bufio.Reader底层的buf的数据。 虽然bufio.Reader 减少系统调用，但是增加了内存的拷贝次数。

 有没有更好的方法在减少系统调用的同时，同时也尽可能减少内存copy 次数?

 字节在[Go 网络库上的实践](https://juejin.cn/post/6844904153173458958) 有提到相应的方案，

 我根据自己业务代码的需求写nocopybuffer, 大概的原理跟bufio 很像，只是业务层代码读取业务报文时，是引用nocopybuffer底层buf的数据，不是拷贝，底层buf可以被多个业务报文引用，当底层buf被引用时，不可以修改底层buf.如果nocopybuffer需要再调用read读取内核数据时，需要从池里分配新的底层buf； 当底层buf没有被任何业务报文引用时，就把这个buf 释放会池里。

 跟bufio标准库的区别是， bufio只有一个底层buf, 现在的nocopybuffer 可以有多个底层buf, 用链表联结起来, 底层buf的分配或释放都通过池来分配或释放。
