---
layout: blog
title: 'Flask源码剖析（十四）：测试'
date: 2017-08-19 21:10:23
categories: flask
tags: flask
lead_text: '测试'
---

## 一. 测试
Flask提供的测试渠道是Werkzeug的Client来处理本地环境，未经测试的应用难于改进现有的代码。

## 二. 代码
Flask实现测试的代码存放于flask/testing.py文件中，如下
```python
# -*- coding: utf-8 -*-

import werkzeug
from contextlib import contextmanager
from werkzeug.test import Client, EnvironBuilder
from flask import _request_ctx_stack

try:
    from werkzeug.urls import url_parse
except ImportError:
    from urlparse import urlsplit as url_parse


def make_test_environ_builder(app, path='/', base_url=None, *args, **kwargs):
    http_host = app.config.get('SERVER_NAME')
    app_root = app.config.get('APPLICATION_ROOT')
    if base_url is None:
        url = url_parse(path)
        base_url = 'http://%s/' % (url.netloc or http_host or 'localhost')
        if app_root:
            base_url += app_root.lstrip('/')
        if url.netloc:
            path = url.path
            if url.query:
                path += '?' + url.query
    return EnvironBuilder(path, base_url, *args, **kwargs)


class FlaskClient(Client):

    preserve_context = False

    def __init__(self, *args, **kwargs):
        super(FlaskClient, self).__init__(*args, **kwargs)
        self.environ_base = {
            "REMOTE_ADDR": "127.0.0.1",
            "HTTP_USER_AGENT": "werkzeug/" + werkzeug.__version__
        }

    @contextmanager
    def session_transaction(self, *args, **kwargs):
        if self.cookie_jar is None:
            raise RuntimeError('Session transactions only make sense '
                               'with cookies enabled.')
        app = self.application
        environ_overrides = kwargs.setdefault('environ_overrides', {})
        self.cookie_jar.inject_wsgi(environ_overrides)
        outer_reqctx = _request_ctx_stack.top
        with app.test_request_context(*args, **kwargs) as c:
            sess = app.open_session(c.request)
            if sess is None:
                raise RuntimeError('Session backend did not open a session. '
                                   'Check the configuration')

            _request_ctx_stack.push(outer_reqctx)
            try:
                yield sess
            finally:
                _request_ctx_stack.pop()

            resp = app.response_class()
            if not app.session_interface.is_null_session(sess):
                app.save_session(sess, resp)
            headers = resp.get_wsgi_headers(c.request.environ)
            self.cookie_jar.extract_wsgi(c.request.environ, headers)

    def open(self, *args, **kwargs):
        kwargs.setdefault('environ_overrides', {}) \
            ['flask._preserve_context'] = self.preserve_context
        kwargs.setdefault('environ_base', self.environ_base)

        as_tuple = kwargs.pop('as_tuple', False)
        buffered = kwargs.pop('buffered', False)
        follow_redirects = kwargs.pop('follow_redirects', False)
        builder = make_test_environ_builder(self.application, *args, **kwargs)

        return Client.open(self, builder,
                           as_tuple=as_tuple,
                           buffered=buffered,
                           follow_redirects=follow_redirects)

    def __enter__(self):
        if self.preserve_context:
            raise RuntimeError('Cannot nest client invocations')
        self.preserve_context = True
        return self

    def __exit__(self, exc_type, exc_value, tb):
        self.preserve_context = False

        top = _request_ctx_stack.top
        if top is not None and top.preserved:
            top.pop()

```

## 三. 解析
先来看看flask/app.py，里面有个调用到测试模块，如下
```python
class Flask(_PackageBoundObject):
    test_client_class = None

    def test_client(self, use_cookies=True, **kwargs):
        cls = self.test_client_class
        if cls is None:
            from flask.testing import FlaskClient as cls
        return cls(self, self.response_class, use_cookies=use_cookies, **kwargs)
```
这段代码是导入测试模块，为程序提供一个测试客户端的入口。
接下来具体看看FlaskClient类是怎么提供测试客户端入口的支持的。
1. FlaskClient类使用了魔术方法__enter__和__exit__实现了上下文管理协议，当使用with关键字调用FlaskClient类时，如果当前的flask程序已经有另外的测试实例在运行，触发__enter__则报错退出，如果只有当前测试实例，退出时把测试环境创建的上下文的最新堆栈数据清空。
2. session_transaction方法是contextmanager装饰器来装饰的，contextmanager其实也是上下文管理协议，
- 如果当前测试用例环境下的cookies为空，则抛出异常
- 模拟客户端调用，从而在测试客户端的上下文中打开一个Session，在事务结束时，抛出当前的模拟数据，Session将恢复到原来的值
3. open函数是基于flask程序的配置和werkzeug.test的Client类和EnvironBuilder类，从而创建一个虚拟的客户端请求对象

## 四. 使用
应用使用可以参考[测试 Flask 应用](http://dormousehole.readthedocs.io/en/latest/testing.html)

## 五. 问题