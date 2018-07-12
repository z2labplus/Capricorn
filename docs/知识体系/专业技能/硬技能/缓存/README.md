## 缓存

- [《缓存失效策略（FIFO 、LRU、LFU三种算法的区别）》](https://blog.csdn.net/clementad/article/details/48229243)

### 本地缓存

- [《HashMap本地缓存》](https://coderxing.gitbooks.io/architecture-evolution/di-er-pian-ff1a-feng-kuang-yuan-shi-ren/42-xing-neng-zhi-ben-di-huan-cun/421-ying-yong-ceng-ben-di-huan-cun/4211.html)
- [《EhCache本地缓存》](https://coderxing.gitbooks.io/architecture-evolution/di-er-pian-ff1a-feng-kuang-yuan-shi-ren/42-xing-neng-zhi-ben-di-huan-cun/421-ying-yong-ceng-ben-di-huan-cun/4212-ehcache.html)
  - 堆内、堆外、磁盘三级缓存。
  - 可按照缓存空间容量进行设置。
  - 按照时间、次数等过期策略。
- [《Guava Cache》](https://coderxing.gitbooks.io/architecture-evolution/di-er-pian-ff1a-feng-kuang-yuan-shi-ren/42-xing-neng-zhi-ben-di-huan-cun/421-ying-yong-ceng-ben-di-huan-cun/4213-guava-cache.html)
  - 简单轻量、无堆外、磁盘缓存。

- [《Nginx本地缓存》](https://coderxing.gitbooks.io/architecture-evolution/di-er-pian-ff1a-feng-kuang-yuan-shi-ren/42-xing-neng-zhi-ben-di-huan-cun/422-fu-wu-duan-ben-di-huan-cun/nginx-ben-di-huan-cun.html)
- [《Pagespeed—懒人工具，服务器端加速》](https://coderxing.gitbooks.io/architecture-evolution/di-er-pian-ff1a-feng-kuang-yuan-shi-ren/42-xing-neng-zhi-ben-di-huan-cun/422-fu-wu-duan-ben-di-huan-cun/4222-pagespeed.html)

## 客户端缓存

- [《浏览器端缓存》](https://coderxing.gitbooks.io/architecture-evolution/di-er-pian-ff1a-feng-kuang-yuan-shi-ren/42-xing-neng-zhi-ben-di-huan-cun/423-ke-hu-duan-huan-cun.html)
  - 主要是利用 Cache-Control 参数。
- [《H5 和移动端 WebView 缓存机制解析与实战》](https://mp.weixin.qq.com/s/qHm_dJBhVbv0pJs8Crp77w)

## 服务端缓存

### Web缓存

- [nuster](https://github.com/jiangwenyuan/nuster) - nuster cache
- [varnish](https://github.com/varnishcache/varnish-cache) - varnish cache
- [squid](https://github.com/squid-cache/squid) - squid cache

### Memcached

- [《Memcached 教程》](http://www.runoob.com/Memcached/Memcached-tutorial.html)
- [《深入理解Memcached原理》](https://blog.csdn.net/chenleixing/article/details/47035453)
  - 采用多路复用技术提高并发性。
  - slab分配算法： memcached给Slab分配内存空间，默认是1MB。分配给Slab之后 把slab的切分成大小相同的chunk，Chunk是用于缓存记录的内存空间，Chunk 的大小默认按照1.25倍的速度递增。好处是不会频繁申请内存，提高IO效率，坏处是会有一定的内存浪费。
- [《Memcached软件工作原理》](https://www.jianshu.com/p/36e5cd400580)
- [《Memcache技术分享：介绍、使用、存储、算法、优化、命中率》](http://zhihuzeye.com/archives/2361)
- [《memcache 中 add 、 set 、replace 的区别》](https://blog.csdn.net/liu251890347/article/details/37690045)
  - 区别在于当key存在还是不存在时，返回值是true和false的。
- [**《memcached全面剖析》**](https://pan.baidu.com/s/1qX00Lti?errno=0&errmsg=Auth%20Login%20Sucess&&bduss=&ssnerror=0&traceid=)

### Redis

- [《Redis 教程》](http://www.runoob.com/redis/redis-tutorial.html)
- [《redis底层原理》](https://blog.csdn.net/wcf373722432/article/details/78678504)
  - 使用 ziplist 存储链表，ziplist是一种压缩链表，它的好处是更能节省内存空间，因为它所存储的内容都是在连续的内存区域当中的。
  - 使用 skiplist(跳跃表)来存储有序集合对象、查找上先从高Level查起、时间复杂度和红黑树相当，实现容易，无锁、并发性好。
- [《Redis持久化方式》](http://doc.redisfans.com/topic/persistence.html)
  - RDB方式：定期备份快照，常用于灾难恢复。优点：通过fork出的进程进行备份，不影响主进程、RDB 在恢复大数据集时的速度比 AOF 的恢复速度要快。缺点：会丢数据。
  - AOF方式：保存操作日志方式。优点：恢复时数据丢失少，缺点：文件大，回复慢。
  - 也可以两者结合使用。
- [《分布式缓存--序列3--原子操作与CAS乐观锁》](https://blog.csdn.net/chunlongyu/article/details/53346436)

#### 架构

- [《Redis单线程架构》](https://blog.csdn.net/sunhuiliang85/article/details/73656830)

#### 回收策略

- [《redis的回收策略》](https://blog.csdn.net/qq_29108585/article/details/63251491)

### Tair

- [官方网站](https://github.com/alibaba/tair)
- [《Tair和Redis的对比》](http://blog.csdn.net/farphone/article/details/53522383)
- 特点：可以配置备份节点数目，通过异步同步到备份节点
- 一致性Hash算法。
- 架构：和Hadoop 的设计思想类似，有Configserver，DataServer，Configserver 通过心跳来检测，Configserver也有主备关系。

几种存储引擎:

- MDB，完全内存性，可以用来存储Session等数据。
- Rdb（类似于Redis），轻量化，去除了aof之类的操作，支持Restfull操作
- LDB（LevelDB存储引擎），持久化存储，LDB 作为rdb的持久化，google实现，比较高效，理论基础是LSM(Log-Structured-Merge Tree)算法，现在内存中修改数据，达到一定量时（和内存汇总的旧数据一同写入磁盘）再写入磁盘，存储更加高效，县比喻Hash算法。
- Tair采用共享内存来存储数据，如果服务挂掉（非服务器），重启服务之后，数据亦然还在。