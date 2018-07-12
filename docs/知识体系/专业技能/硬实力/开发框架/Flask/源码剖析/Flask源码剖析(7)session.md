---
layout: blog
title: 'Flask源码剖析（七）：session'
date: 2017-06-27 21:23:43
categories: flask
tags: flask
lead_text: 'session'
---

## 一. 含义
HTTP是无状态协议，当服务需要记录用户的状态时，就需要一种机制来识别具体的某个用户。
Session和Cookie就是解决这个问题所提出的两个机制。

## 二. 场景
### Cookie
Cookie是浏览器（User Agent）访问一些网站后，这些网站存放在客户端的一组数据，用于使网站等跟踪用户，实现用户自定义功能。
当登录一个网站时，第一次需要输入账号密码，第二天要继续访问此网站时，网站页面的脚本可读取cookie中的值，自动填了账号密码，方便用户。
这也是cookie的名称由来，先给用户一点甜头尝尝。

### Session
Session是存放在服务器端的类似于HashTable结构来存放用户数据。
当在一个网站上购物要下单时，客户端知道添加了哪些商品，那服务端是怎么知道是哪个用户提交的呢？这个时候需要用到Session机制了。

## 三. 区别
> session
> 会话，代表服务器与浏览器的一次会话过程，这个过程是连续的，也可以时断时续。
> cookie中存放着一个sessionID，请求时会发送这个ID；
> session因为请求（request对象）而产生；
> session是一个容器，可以存放会话过程中的任何对象；
> session的创建与使用总是在服务端，浏览器从来都没有得到过session对象；
> session是一种http存储机制，目的是为武装的http提供持久机制。

> cookie
> 储存在用户本地终端上的数据，服务器生成，发送给浏览器，下次请求统一网站给服务器。

> cookie与session区别
> cookie数据存放在客户端上，session数据放在服务器上；
> cookie不是很安全，且保存数据有限；
> session一定时间内保存在服务器上,当访问增多，占用服务器性能。

> 摘自[cookie和session的的区别以及应用场景有哪些？](https://www.zhihu.com/question/31079651/answer/149755672)

## 四. 实现
这里先不细说cookie的实现，来看看Flask里面是怎么调用session，实现session的。
下面有个例子：
```python
# -*- coding: utf-8 -*-

from flask import Flask, session, redirect, url_for, escape, request

app = Flask(__name__)

@app.route('/')
def index():
    if 'username' in session:
        return 'User: %s' % escape(session['username'])
    return 'Not login in'


@app.route('/login', methods=['GET', 'POST'])
def login():
    if request.method == 'POST':
        session['username'] = request.form['username']
        return redirect(url_for('index'))
    return '''
        <form action="" method="post">
            <p><input type=text name=username>
            <p><input type=submit value=Login>
        </form>
    '''

@app.route('/logout')
def logout():
    session.pop('username', None)
    return redirect(url_for('index'))

app.secret_key = '44c032ac7e25d3fec5ccab4c71971176'

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000, debug=True)
```
在Flask中使用session是非常简单的，直接导入session变量即可。
Flask中session是一个全局对象，每一个请求的上下文中都有一个session全局对象。

从上下文这一章可以知道，请求上下文演化了两个变量：request/session，即request/session实例都是RequestContext实例。
下面来看看Flask是怎么实现session的。
还是从wsgi_app()函数入手：
```python
    def wsgi_app(self, environ, start_response):
        ctx = self.request_context(environ)
        ctx.push()
        ......
```
```python
    def request_context(self, environ):
        return RequestContext(self, environ)
```
```python
class RequestContext(object):

    def __init__(self, app, environ, request=None):
        self.session = None
        .....
    
    def push(self):
        ......
        self.session = self.app.open_session(self.request)
        if self.session is None:
            self.session = self.app.make_null_session()
```
很明显，每次有请求进来的时候，session在RequestContext类中实例化，保留在请求上下文中。
那开始研究一下session是怎么初始化的
```python
self.session = self.app.open_session(self.request)
        if self.session is None:
            self.session = self.app.make_null_session()
```
这段代码主要是判断是否可以从请求上下文中获取到session的值，open_session()返回None，那么调用make_null_session()生成一个空session。
接下来看看open_session()是怎么获取到session的值。
```python
class SecureCookieSessionInterface(SessionInterface):

    def get_signing_serializer(self, app):
        if not app.secret_key:
            return None
        signer_kwargs = dict(
            key_derivation=self.key_derivation,
            digest_method=self.digest_method
        )
        return URLSafeTimedSerializer(app.secret_key, salt=self.salt,
                                      serializer=self.serializer,
                                      signer_kwargs=signer_kwargs)

    def open_session(self, app, request):
        s = self.get_signing_serializer(app)
        if s is None:
            return None
        val = request.cookies.get(app.session_cookie_name)
        if not val:
            return self.session_class()
        max_age = total_seconds(app.permanent_session_lifetime)
        try:
            data = s.loads(val, max_age=max_age)
            return self.session_class(data)
        except BadSignature:
            return self.session_class()
```
当secret_key变量为空时，open_session()返回None。
当secret_key变量不为空时，open_session()调用相关函数生成session，来看看是怎么运作的。
1. 根据cookies获取session的值且验证数据通过，如果session的值不为空，返回一个session_class()实例化对象，即是上下文中还是原来的session的值。
2. 当session的值为空，调用数据验证和序列化函数生成新的session值，并生成session_class()实例化对象。

## 五. 问题
至此，一个请求的session初始化，数据验证，数据生成和返回都已大致说明白了。其中还有几个点没有说透的。
1. Flask是怎么解析cookies的值为session对象的？
2. session_class对象的值（session的值）是什么格式？包括了什么？
3. session的值是怎么进行数据验证的？
