---
layout: blog
title: 'Flask源码剖析（一）：总览'
date: 2017-04-25 01:03:16
categories: flask
tags: flask
lead_text: '有人说，人成熟的开始就是在不断认识自己的渺小和平凡，然后不断的妥协。'
---


## 一. 背景

做一件事无非就三个过程，做这件事的目的是什么？有什么宗旨？想达到什么样的结果？

阅读一份源码应当也是同样的。

说到目的，要也得先知道这个目的从何而来。

有人说，人成熟的开始就是在不断认识自己的渺小和平凡，然后不断的妥协。


因此，为什么要剖析Flask源码？

目的大概是以下几点：

* 如果这份代码在我的项目中使用了，那么它就是我的程序，必须对它负责，搞懂它。
* 用户使用软件，他们并不在意是我的错误代码还是别人的引起的。所以，所有的bug都应当是我的bug。
* 不要相信文档，文档永远都是过时的，只有不过时的代码，代码能给予最好的指引。


宗旨，这里指的是剖析源码采用的方法，参考知乎上面的回答，[如何去阅读并学习一些优秀的开源框架的源码？](https://www.zhihu.com/question/26766601)/[如何以“正确的姿势”阅读开源软件代码](http://mp.weixin.qq.com/s?__biz=MjM5Mjg4NDMwMA==&mid=2652973508&idx=1&sn=1281837abb0530893f8b42e05ea35a7e#rd)。

我的建议只有三点：

* 使用IDE看源码，一是直观，二是方便调试。
* 几万行的项目随便折腾，找准一点研究就好啦，别管那么多方法论，怎么舒服怎么来。当然，掌握一定方法还是很有帮助的。
* 抽象细节，纵览全局；抓住细节，深究原理。


那么所要达到预期的结果是：

* 熟悉Flask组织结构，把握细节实现原理。
* 更加深入理解python的基本实现方法。

## 二. 简介

> Flask is a microframework for Python based on Werkzeug, Jinja 2 and good intentions.

> “Micro” does not mean that your whole web application has to fit into a single Python file (although it certainly can), nor does it mean that Flask is lacking in functionality. The “micro” in microframework means Flask aims to keep the core simple but extensible. Flask won’t make many decisions for you, such as what database to use. Those decisions that it does make, such as what templating engine to use, are easy to change. Everything else is up to you, so that Flask can be everything you need and nothing you don’t.
>
> By default, Flask does not include a database abstraction layer, form validation or anything else where different libraries already exist that can handle that. Instead, Flask supports extensions to add such functionality to your application as if it was implemented in Flask itself. Numerous extensions provide database integration, form validation, upload handling, various open authentication technologies, and more. Flask may be “micro”, but it’s ready for production use on a variety of needs.

以上来自[**Flask官网文档**](http://flask.pocoo.org/docs/dev/foreword/#what-does-micro-mean)，从这段引用，我们可以得知几点：

* Flask是一个微框架。
* Flask是依赖[Werkzeug](http://werkzeug.pocoo.org/docs/0.11/)，[Jinja2](http://jinja.pocoo.org/docs/2.9/)开发的。
* Flask是旨在保持核心简单，易于扩展。
* 默认情况下，Flask不包含数据库抽象层、表单验证，或是其他任何多种库可以胜任的功能。
* Flask支持扩展来给应用添加功能，其中包括数据库集成、表单验证、上传处理、各种各样的开放认证技术等。
* Flask虽然看起来是微小的，但也可以在复杂的环境下使用。

## 三. 结构 

这里研读的源码是[Flask-0.12.tar.gz](https://pypi.python.org/packages/4b/3a/4c20183df155dd2e39168e35d53a388efb384a512ca6c73001d8292c094a/Flask-0.12.tar.gz#md5=c1d30f51cff4a38f9454b23328a15c5a)，结构如下：

```python
.
├── app.py                  # 主要提供创建flask实例和对请求、响应进行处理的功能
├── blueprints.py           # 提供模块化/蓝本功能
├── cli.py                  # 提供了命令行与flask app交互的功能
├── _compat.py              # 定义了对py2和py3的兼容，涉及到不同版本的对象，先在该文件中进行校验处理
├── config.py               # 主要提供配置文件的功能
├── ctx.py                  # 主要定义了上下文管理器的类
├── debughelpers.py         # 定义了各种debug模式的错误类型
├── ext                     # 提供flask.ext.扩展名的方式来导入“flask_扩展名”和“flaskext.扩展名”的功能
│   └── __init__.py
├── exthook.py              # 提供了ext目录需要用到的类，即导入钩子的类
├── globals.py              # 主要提供全局变量，局部变量和上下文管理器的实例
├── helpers.py              # 主要提供诸多辅助功能
├── __init__.py             # 提供了版本信息和需要导入的模块
├── json.py                 # 主要提供json格式数据的解析功能
├── logging.py              # 定义了日志管理器的类和创建函数
├── __main__.py             # 提供执行命令行交互的别名
├── sessions.py             # 提供session的类定义，包含了cookie机制
├── signals.py              # 主要提供不同机制的信号实例
├── templating.py           # 主要提供模板渲染功能
├── testing.py              # 定义用于测试而非生产的一些基类、函数等
├── views.py                # 提供另一种以类来定义视图函数的方式
└── wrappers.py             # 定义了对请求和响应的封装类

1 directory, 21 files
```

> Tips: 初略浏览每个模块的注释，了解大概功能。

有效代码行数：

```python
21 text files.
21 unique files.                              
0 files ignored.

http://cloc.sourceforge.net v 1.60  T=0.08 s (259.6 files/s, 80793.3 lines/s)
-------------------------------------------------------------------------------
Language                     files          blank        comment           code
-------------------------------------------------------------------------------
Python                          21           1142           2653           2741
-------------------------------------------------------------------------------
SUM:                            21           1142           2653           2741
-------------------------------------------------------------------------------
```

app.py的有效代码行数：

```python
1 text file.
1 unique file.                              
0 files ignored.

http://cloc.sourceforge.net v 1.60  T=0.01 s (67.2 files/s, 134301.5 lines/s)
-------------------------------------------------------------------------------
Language                     files          blank        comment           code
-------------------------------------------------------------------------------
Python                           1            318            998            684
-------------------------------------------------------------------------------
```

Flask的有效代码函数不过2k多，其中最主要的创建实例的文件app.py占据了1/4。

因此，我的建议是采用自顶而下的研读方法，理清楚源码的整体脉络，按照模块去阅读代码，记录类/函数之间的调用关系。

## 四. 组成

### 1. 标准库

[**String Services**](https://docs.python.org/2.7/library/strings.html)

* [*cStringIO*: Faster version of *StringIO*](https://docs.python.org/2.7/library/stringio.html#module-cStringIO)
* [*codecs*: Codec registry and base classes](https://docs.python.org/2.7/library/codecs.html)
* [*unicodedata*: Unicode Database](https://docs.python.org/2.7/library/unicodedata.html)

[**Data Types**](https://docs.python.org/2.7/library/datatypes.html)

* [*datetime*: Basic date and time types](https://docs.python.org/2.7/library/datetime.html)
* [*types*: Names for built-in types](https://docs.python.org/2.7/library/types.html)

[**Numeric and Mathematical Modules**](https://docs.python.org/2.7/library/numeric.html)

* [*itertools*: Functions creating iterators for efficient looping](https://docs.python.org/2.7/library/itertools.html)
* [*functools*: Higher-order functions and operations on callable objects](https://docs.python.org/2.7/library/functools.html)

[**File and Directory Access**](https://docs.python.org/2.7/library/filesys.html)

* [*os.path*: Common pathname manipulations](https://docs.python.org/2.7/library/os.path.html)

[**Data Compression and Archiving**](https://docs.python.org/2.7/library/archiving.html)

* [*zlib*: Compression compatible with *gzip*](https://docs.python.org/2.7/library/zlib.html)

[**Cryptographic Services**](https://docs.python.org/2.7/library/crypto.html)

* [*hashlib*: Secure hashes and message digests](https://docs.python.org/2.7/library/hashlib.html)

[**Generic Operating System Services**](https://docs.python.org/2.7/library/allos.html)

* [*os*: Miscellaneous operating system interfaces](https://docs.python.org/2.7/library/os.html)
* [*io*: Core tools for working with streams](https://docs.python.org/2.7/library/io.html)
* [*time*: Time access and conversions](https://docs.python.org/2.7/library/time.html)
* [*logging*: Logging facility for Python](https://docs.python.org/2.7/library/logging.html)
* [*errno*: Standard errno system symbols](https://docs.python.org/2.7/library/errno.html)

[**Optional Operating System Services**](https://docs.python.org/2.7/library/someos.html)

* [*threading*: Higher-level threading interface](https://docs.python.org/2.7/library/threading.html)

[**Internet Data Handling**](https://docs.python.org/2.7/library/netdata.html)

* [*mimetypes*: Map filenames to MIME types](https://docs.python.org/2.7/library/mimetypes.html)
* [*base64*: RFC 3548: Base16, Base32, Base64 Data Encodings](https://docs.python.org/2.7/library/base64.html)

[**Internet Protocols and Support**](https://docs.python.org/2.7/library/internet.html)

* [*uuid*: UUID objects according to RFC 4122](https://docs.python.org/2.7/library/uuid.html)
* [*urlparse*: Parse URLs into components](https://docs.python.org/2.7/library/urlparse.html)

[**Python Runtime Services**](https://docs.python.org/2.7/library/python.html)

* [*sys*: System-specific parameters and functions](https://docs.python.org/2.7/library/sys.html)
* [*warnings*: Warning control](https://docs.python.org/2.7/library/warnings.html)
* [*contextlib*: Utilities for *with*-statement contexts](https://docs.python.org/2.7/library/contextlib.html)
* [*traceback*: Print or retrieve a stack traceback](https://docs.python.org/2.7/library/traceback.html)


* [\_\_future\_\_: Future statement definitions](https://docs.python.org/2.7/library/__future__.html)

[**Importing Modules**](https://docs.python.org/2.7/library/modules.html)

* [*pkgutil*: Package extension utility](https://docs.python.org/2.7/library/pkgutil.html)

### 2. 第三方库

* [*werkzeug*: The Python WSGI Utility Library ](http://werkzeug.pocoo.org/)
* [*jinja2*: The Python Template Engine](http://jinja.pocoo.org/)
* [*itsdangerous*: Use HMAC and SHA1 for signing ](http://pythonhosted.org/itsdangerous/)
* [*click*: Beautiful command line interfaces](http://click.pocoo.org/5/)
* [*blinker*: Fast & simple object-to-object and broadcast signaling](http://pythonhosted.org/blinker/)

