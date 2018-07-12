## 基础设施

### 常规监控

- [《腾讯业务系统监控的修炼之路》](https://blog.csdn.net/enweitech/article/details/77849205)
  - 监控的方式：主动、被动、旁路(比如舆情监控)
  - 监控类型： 基础监控、服务端监控、客户端监控、
    监控、用户端监控
  - 监控的目标：全、块、准
  - 核心指标：请求量、成功率、耗时
- [《开源还是商用？十大云运维监控工具横评》](https://www.oschina.net/news/67525/monitoring-tools)
  - Zabbix、Nagios、Ganglia、Zenoss、Open-falcon、监控宝、 360网站服务监控、阿里云监控、百度云观测、小蜜蜂网站监测等。
- [《监控报警系统搭建及二次开发经验》](http://developer.51cto.com/art/201612/525373.htm)

**命令行监控工具**

- [《常用命令行监控工具》](https://coderxing.gitbooks.io/architecture-evolution/di-er-pian-ff1a-feng-kuang-yuan-shi-ren/44-an-quan-yu-yun-wei/445-fu-wu-qi-zhuang-tai-jian-ce/4451-ming-ling-xing-gong-ju.html)
  - top、sar、tsar、nload
- [《20个命令行工具监控 Linux 系统性能》](http://blog.jobbole.com/96846/)
- [《JVM性能调优监控工具jps、jstack、jmap、jhat、jstat、hprof使用详解》](https://my.oschina.net/feichexia/blog/196575)

### APM

APM —  Application Performance Management

- [《Dapper，大规模分布式系统的跟踪系统》](http://bigbully.github.io/Dapper-translation/)
- [CNCF OpenTracing](http://opentracing.io)，[中文版](https://github.com/opentracing-contrib/opentracing-specification-zh)
- 主要开源软件，按字母排序
  - [Apache SkyWalking](https://github.com/apache/incubator-skywalking)
  - [CAT](https://github.com/dianping/cat)
  - [CNCF jaeger](https://github.com/jaegertracing/jaeger)
  - [Pinpoint](https://github.com/naver/pinpoint)
  - [Zipkin](https://github.com/openzipkin/zipkin)
- [《开源APM技术选型与实战》](http://www.infoq.com/cn/articles/apm-Pinpoint-practice)
  - 主要基于 Google的Dapper（大规模分布式系统的跟踪系统）思想。
  
- 应用性能监控
- 异常监控
- 日志
- 流量监控

### 持续集成(CI/CD)

- [《持续集成是什么？》](http://www.ruanyifeng.com/blog/2015/09/continuous-integration.html)
- [《8个流行的持续集成工具》](https://www.testwo.com/article/1170)

#### Jenkins

- [《使用Jenkins进行持续集成》](https://www.liaoxuefeng.com/article/001463233913442cdb2d1bd1b1b42e3b0b29eb1ba736c5e000)

#### 环境分离

开发、测试、生成环境分离。

- [《开发环境、生产环境、测试环境的基本理解和区》](https://my.oschina.net/sancuo/blog/214904)

### 自动化运维

#### Ansible

- [《Ansible中文权威指南》](http://www.ansible.com.cn/)
- [《Ansible基础配置和企业级项目实用案例》](https://www.cnblogs.com/heiye123/articles/7855890.html)

#### puppet

- [《自动化运维工具——puppet详解》](https://www.cnblogs.com/keerya/p/8040071.html)

#### chef

- [《Chef 的安装与使用》](https://www.ibm.com/developerworks/cn/cloud/library/1407_caomd_chef/)

### 虚拟化

- [《VPS的三种虚拟技术OpenVZ、Xen、KVM优缺点比较》](https://blog.csdn.net/enweitech/article/details/52910082)

#### KVM

- [《KVM详解，太详细太深入了，经典》](http://blog.chinaunix.net/uid-20201831-id-5775661.html)
- [《【图文】KVM 虚拟机安装详解》](https://www.coderxing.com/kvm-install.html)

#### Xen

- [《Xen虚拟化基本原理详解》](https://www.cnblogs.com/sddai/p/5931201.html)

#### OpenVZ

- [《开源Linux容器 OpenVZ 快速上手指南》](https://blog.csdn.net/longerzone/article/details/44829255)

### 容器技术

#### Docker

- [《几张图帮你理解 docker 基本原理及快速入门》](https://www.cnblogs.com/SzeCheng/p/6822905.html)
- [《Docker 核心技术与实现原理》](https://draveness.me/docker)
- [《Docker 教程》](http://www.runoob.com/docker/docker-tutorial.html)

### 云技术

#### OpenStack

- [《OpenStack构架知识梳理》](https://www.cnblogs.com/klb561/p/8660264.html)

##3 DevOps

- [《一分钟告诉你究竟DevOps是什么鬼？》](https://www.cnblogs.com/jetzhang/p/6068773.html)
- [《DevOps详解》](http://www.infoq.com/cn/articles/detail-analysis-of-devops)