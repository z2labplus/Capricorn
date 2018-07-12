## Web Server

### Nginx

- [《Ngnix的基本学习-多进程和Apache的比较》](https://blog.csdn.net/qq_25797077/article/details/52200722)
  - Nginx 通过异步非阻塞的事件处理机制实现高并发。Apache 每个请求独占一个线程，非常消耗系统资源。
  - 事件驱动适合于IO密集型服务(Nginx)，多进程或线程适合于CPU密集型服务(Apache)，所以Nginx适合做反向代理，而非web服务器使用。  
- [《nginx与Apache的对比以及优缺点》](https://www.cnblogs.com/cunkouzh/p/5410154.html)
  - nginx只适合静态和反向代理，不适合处理动态请求。

### OpenResty

- [官方网站](http://openresty.org/cn/)
- [《浅谈 OpenResty》](http://www.linkedkeeper.com/detail/blog.action?bid=1034)
  - 通过 Lua 模块可以在Nginx上进行开发。 

### Apache Httpd

- [官方网站](http://httpd.apache.org/)

## 消息队列

- [《消息队列-推/拉模式学习 & ActiveMQ及JMS学习》](https://www.cnblogs.com/charlesblc/p/6045238.html)
  - RabbitMQ 消费者默认是推模式（也支持拉模式）。
  - Kafka 默认是拉模式。
  - Push方式：优点是可以尽可能快地将消息发送给消费者，缺点是如果消费者处理能力跟不上，消费者的缓冲区可能会溢出。
  - Pull方式：优点是消费端可以按处理能力进行拉去，缺点是会增加消息延迟。
- [《Kafka、RabbitMQ、RocketMQ等消息中间件的对比 —— 消息发送性能和区别》](https://blog.csdn.net/yunfeng482/article/details/72856762)

### 消息总线

消息总线相当于在消息队列之上做了一层封装，统一入口，统一管控、简化接入成本。

- [《消息总线VS消息队列》](https://blog.csdn.net/yanghua_kobe/article/details/43877281)

### 消息的顺序

- [《如何保证消费者接收消息的顺序》](https://www.cnblogs.com/cjsblog/p/8267892.html)

### RabbitMQ

支持事务，推拉模式都是支持、适合需要可靠性消息传输的场景。

- [《RabbitMQ的应用场景以及基本原理介绍》](https://blog.csdn.net/whoamiyang/article/details/54954780)
- [《消息队列之 RabbitMQ》](https://www.jianshu.com/p/79ca08116d57) 
- [《RabbitMQ之消息确认机制（事务+Confirm）》](https://blog.csdn.net/u013256816/article/details/55515234)

### RocketMQ

Java实现，推拉模式都是支持，吞吐量逊于Kafka。可以保证消息顺序。

- [《RocketMQ 实战之快速入门》](https://www.jianshu.com/p/824066d70da8)
- [《RocketMQ 源码解析》](http://www.iocoder.cn/categories/RocketMQ/?vip&architect-awesome)

### ActiveMQ

纯Java实现，兼容JMS，可以内嵌于Java应用中。

- [《ActiveMQ消息队列介绍》](https://www.cnblogs.com/wintersun/p/3962302.html)

### Kafka

高吞吐量、采用拉模式。适合高IO场景，比如日志同步。

- [官方网站](http://kafka.apache.org/)
- [《各消息队列对比，Kafka深度解析，众人推荐，精彩好文！》](https://blog.csdn.net/allthesametome/article/details/47362451)
- [《Kafka分区机制介绍与示例》](http://lxw1234.com/archives/2015/10/538.htm)

### Redis 消息推送

生产者、消费者模式完全是客户端行为，list 和 拉模式实现，阻塞等待采用 blpop 指令。

- [《Redis学习笔记之十：Redis用作消息队列》](https://blog.csdn.net/qq_34212276/article/details/78455004)

### ZeroMQ

 TODO

## 定时调度

### 单机定时调度

- [《linux定时任务cron配置》](https://www.cnblogs.com/shuaiqing/p/7742382.html)
- [《Linux cron运行原理》](https://my.oschina.net/daquan/blog/483305)
  - fork 进程 + sleep 轮询
- [《Quartz使用总结》](https://www.cnblogs.com/drift-ice/p/3817269.html)
- [《Quartz源码解析 ---- 触发器按时启动原理》](https://blog.csdn.net/wenniuwuren/article/details/42082981/)
- [《quartz原理揭秘和源码解读》](https://www.jianshu.com/p/bab8e4e32952)
  - 定时调度在 QuartzSchedulerThread 代码中，while()无限循环，每次循环取出时间将到的trigger，触发对应的job，直到调度器线程被关闭。

### 分布式定时调度

- [《这些优秀的国产分布式任务调度系统，你用过几个？》](https://blog.csdn.net/qq_16216221/article/details/70314337)
  - opencron、LTS、XXL-JOB、Elastic-Job、Uncode-Schedule、Antares
- [《Quartz任务调度的基本实现原理》](https://www.cnblogs.com/zhenyuyaodidiao/p/4755649.html)
  - Quartz集群中，独立的Quartz节点并不与另一其的节点或是管理节点通信，而是通过相同的数据库表来感知到另一Quartz应用的 
- [《Elastic-Job-Lite 源码解析》](http://www.iocoder.cn/categories/Elastic-Job-Lite/?vip&architect-awesome)
- [《Elastic-Job-Cloud 源码解析》](http://www.iocoder.cn/categories/Elastic-Job-Cloud/?vip&architect-awesome)

## RPC

- [《从零开始实现RPC框架 - RPC原理及实现》](https://blog.csdn.net/top_code/article/details/54615853)
  - 核心角色：Server: 暴露服务的服务提供方、Client: 调用远程服务的服务消费方、Registry: 服务注册与发现的注册中心。
- [《分布式RPC框架性能大比拼 dubbo、motan、rpcx、gRPC、thrift的性能比较》](https://blog.csdn.net/testcs_dn/article/details/78050590)

### Dubbo

- [官方网站](http://dubbo.apache.org/)
- [dubbo实现原理简单介绍](https://www.cnblogs.com/steven520213/p/7606598.html)

** SPI **
TODO

### Thrift

- [官方网站](http://thrift.apache.org/)
- [《Thrift RPC详解》](https://blog.csdn.net/kesonyk/article/details/50924489)
  - 支持多语言，通过中间语言定义接口。

### gRPC

服务端可以认证加密，在外网环境下，可以保证数据安全。

- [官方网站](https://grpc.io/)
- [《你应该知道的RPC原理》](https://www.cnblogs.com/LBSer/p/4853234.html)

## 数据库中间件

### Sharding Jdbc

- [官网](http://shardingjdbc.io/)

## 日志系统

### 日志搜集

- [《从零开始搭建一个ELKB日志收集系统》](http://cjting.me/misc/build-log-system-with-elkb/)
- [《用ELK搭建简单的日志收集分析系统》](https://blog.csdn.net/lzw_2006/article/details/51280058)
- [《日志收集系统-探究》](https://www.cnblogs.com/beginmind/p/6058194.html)

## 配置中心

- [Apollo - 携程开源的配置中心应用](https://github.com/ctripcorp/apollo)
  - Spring Boot 和 Spring Cloud
  - 支持推、拉模式更新配置
  - 支持多种语言 
- [《基于zookeeper实现统一配置管理》](https://blog.csdn.net/u011320740/article/details/78742625)
- [《 Spring Cloud Config 分布式配置中心使用教程》](https://www.cnblogs.com/shamo89/p/8016908.html)

servlet 3.0 异步特性可用于配置中心的客户端

- [《servlet3.0 新特性——异步处理》](https://www.cnblogs.com/dogdogwang/p/7151866.html)

## API 网关

主要职责：请求转发、安全认证、协议转换、容灾。

- [《API网关那些儿》](http://yunlzheng.github.io/2017/03/14/the-things-about-api-gateway/)
- [《谈API网关的背景、架构以及落地方案》](http://www.infoq.com/cn/news/2016/07/API-background-architecture-floo)
- [《使用Zuul构建API Gateway》](https://blog.csdn.net/zhanglh046/article/details/78651993)
- [《Spring Cloud Gateway 源码解析》](http://www.iocoder.cn/categories/Spring-Cloud-Gateway/?vip&architect-awesome)
- [《HTTP API网关选择之一Kong介绍》](https://mp.weixin.qq.com/s/LIq2CiXJQmmjBC0yvYLY5A)