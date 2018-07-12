---
layout: blog
title: 'Flask源码剖析（四）：上下文'
date: 2017-06-08 20:35:41
categories: flask
tags: flask
lead_text: '上下文'
---

## 一. 概念
>软件工程中，上下文是一种属性的有序序列，它们给驻留在环境内的对象定义了环境。在对象的激活过程中创建上下文，对象被配置为要求某些自动服务，如同步、事务、实时激活、安全性等等。
在计算机技术中，相对于进程而言，上下文就是进程执行的环境。具体来说就是各个变量和数据，包括所有的寄存器变量，进程打开的文件、内存信息等。--百度百科

>每一段程序都有很多外部变量。只有像Add这种简单的函数才是没有外部变量的。一旦你的一段程序有了外部变量，这段程序就不完整，不能独立运行。
你为了使他们运行，就要给所有的外部变量一个一个写一些值进去。这些值的集合就叫上下文。
譬如说在C++的lambda表达是里面，[写在这里的就是上下文](int a, int b){ ... }。
-- [vech 知乎轮子哥](https://www.zhihu.com/question/26387327/answer/32611575)

## 二. 定义
Flask提供了两种上下文环境，一个是应用上下文(Application Context)，另一个是请求上下文(Request Context)。
从名字上就可以知道一个是应用级别的，另一个是单个请求级别的。
应用上下文演化了两个变量：current_app/g
请求上下文演化了两个变量：request/session

通俗地解释一下application context与request context，来自于[flask上下文实现](https://segmentfault.com/a/1190000004223296)：
>1. application 指的就是当你调用app = Flask(name)创建的这个对象app；
>2. request 指的是每次http请求发生时，WSGI server(比如gunicorn)调Flask.call()之后，在Flask对象内部创建的Request对象；
>3. application 表示用于响应WSGI请求的应用本身，request 表示每次http请求;
>4. application的生命周期大于request，一个application存活期间，可能发生多次http请求，所以，也就会有多个request

## 三. 实现
上下文有关的内容定义在globals.py文件中，如下：
```python
# -*- coding: utf-8 -*-

from functools import partial
from werkzeug.local import LocalStack, LocalProxy


_request_ctx_err_msg = '''\
Working outside of request context.

This typically means that you attempted to use functionality that needed
an active HTTP request.  Consult the documentation on testing for
information about how to avoid this problem.\
'''
_app_ctx_err_msg = '''\
Working outside of application context.

This typically means that you attempted to use functionality that needed
to interface with the current application object in a way.  To solve
this set up an application context with app.app_context().  See the
documentation for more information.\
'''


def _lookup_req_object(name):
    top = _request_ctx_stack.top
    if top is None:
        raise RuntimeError(_request_ctx_err_msg)
    return getattr(top, name)


def _lookup_app_object(name):
    top = _app_ctx_stack.top
    if top is None:
        raise RuntimeError(_app_ctx_err_msg)
    return getattr(top, name)


def _find_app():
    top = _app_ctx_stack.top
    if top is None:
        raise RuntimeError(_app_ctx_err_msg)
    return top.app


# context locals
_request_ctx_stack = LocalStack()
_app_ctx_stack = LocalStack()
current_app = LocalProxy(_find_app)
request = LocalProxy(partial(_lookup_req_object, 'request'))
session = LocalProxy(partial(_lookup_req_object, 'session'))
g = LocalProxy(partial(_lookup_app_object, 'g'))
```
globals.py文件非常简单，主要是封装了werkzeug的local.py的文件。
这里用了两个类：LocalStack()和LocalProxy()，是local.py提供的，而LocalStack()类是基于Local()类的，LocalProxy()是提供给Local()和LocalStack()的代理。
因此，在分析LocalStack()和LocalProxy(),先来看看Local()类的实现，如下：
```python
# -*- coding: utf-8 -*-

import copy
from functools import update_wrapper
from werkzeug.wsgi import ClosingIterator
from werkzeug._compat import PY2, implements_bool

try:
    from greenlet import getcurrent as get_ident
except ImportError:
    try:
        from thread import get_ident
    except ImportError:
        from _thread import get_ident
        
def release_local(local):
    local.__release_local__()
        
class Local(object):
    __slots__ = ('__storage__', '__ident_func__')

    def __init__(self):
        object.__setattr__(self, '__storage__', {})
        object.__setattr__(self, '__ident_func__', get_ident)

    def __iter__(self):
        return iter(self.__storage__.items())

    def __call__(self, proxy):
        """Create a proxy for a name."""
        return LocalProxy(self, proxy)

    def __release_local__(self):
        self.__storage__.pop(self.__ident_func__(), None)

    def __getattr__(self, name):
        try:
            return self.__storage__[self.__ident_func__()][name]
        except KeyError:
            raise AttributeError(name)

    def __setattr__(self, name, value):
        ident = self.__ident_func__()
        storage = self.__storage__
        try:
            storage[ident][name] = value
        except KeyError:
            storage[ident] = {name: value}

    def __delattr__(self, name):
        try:
            del self.__storage__[self.__ident_func__()][name]
        except KeyError:
            raise AttributeError(name)
```
Local()类的实例对象包含了两个属性，一个是命名为__storage__的字典，一个是命名为__ident_func__方法，实质上__ident_func__方法是get_ident。
* 其中get_ident是得到当前的线程号,Local通过以线程号作为Key来建立字典,以保证线程间的隔离。
* 这个字典的每个Value也是个字典,用来设置当时线程中的属性。
命名为__storage__的字典是一个嵌套字典形式
* __storage__本身就是一个字典，name和value是一组键值,value是一个列表
* 内部形式实际上是__storage__={{ident1:{name1:value1}},{ident2:{name2:value2}}},
取值方式__getattr__就是__storage__[get_ident()][name]
* 每个线程对应的上下文栈都是自己本身，不会搞混

以后再细看的threading.local的效果--多线程或者多协程情况下全局变量的隔离效果。

下面看看LocalStack(),这个类是基于Local()实现的栈结构。
```python
class LocalStack(object):
    
    def __init__(self):
    '''初始化'''
        self._local = Local()

    def __release_local__(self):
    '''清空当前线程或者协程的栈数据'''
        self._local.__release_local__()

    def _get__ident_func__(self):
    '''获取__ident_func__的值'''
        return self._local.__ident_func__

    def _set__ident_func__(self, value):
    '''设置__ident_func__的值'''
        object.__setattr__(self._local, '__ident_func__', value)
    __ident_func__ = property(_get__ident_func__, _set__ident_func__)
    del _get__ident_func__, _set__ident_func__

    def __call__(self):
    '''返回当前线程或者协程栈顶元素的代理对象'''
        def _lookup():
            rv = self.top
            if rv is None:
                raise RuntimeError('object unbound')
            return rv
        return LocalProxy(_lookup)

    def push(self, obj):
    '''推一个值入栈'''
        rv = getattr(self._local, 'stack', None)
        if rv is None:
            self._local.stack = rv = []
        rv.append(obj)
        return rv

    def pop(self):
    '''退一个值出栈'''
        stack = getattr(self._local, 'stack', None)
        if stack is None:
            return None
        elif len(stack) == 1:
            release_local(self._local)
            return stack[-1]
        else:
            return stack.pop()

    @property
    def top(self):
    '''获取最新的栈值'''
        try:
            return self._local.stack[-1]
        except (AttributeError, IndexError):
            return None
```
LocalProxy()是Local()对象的代理，将所有的操作转到Local()对象处理。
```python
@implements_bool
class LocalProxy(object):
    __slots__ = ('__local', '__dict__', '__name__')

    def __init__(self, local, name=None):
        object.__setattr__(self, '_LocalProxy__local', local)
        object.__setattr__(self, '__name__', name)

    def _get_current_object(self):
        if not hasattr(self.__local, '__release_local__'):
            return self.__local()
        try:
            return getattr(self.__local, self.__name__)
        except AttributeError:
            raise RuntimeError('no object bound to %s' % self.__name__)

    @property
    def __dict__(self):
        try:
            return self._get_current_object().__dict__
        except RuntimeError:
            raise AttributeError('__dict__')

    def __repr__(self):
        try:
            obj = self._get_current_object()
        except RuntimeError:
            return '<%s unbound>' % self.__class__.__name__
        return repr(obj)

    def __bool__(self):
        try:
            return bool(self._get_current_object())
        except RuntimeError:
            return False

    def __unicode__(self):
        try:
            return unicode(self._get_current_object())
        except RuntimeError:
            return repr(self)

    def __dir__(self):
        try:
            return dir(self._get_current_object())
        except RuntimeError:
            return []

    def __getattr__(self, name):
        if name == '__members__':
            return dir(self._get_current_object())
        return getattr(self._get_current_object(), name)

    def __setitem__(self, key, value):
        self._get_current_object()[key] = value

    def __delitem__(self, key):
        del self._get_current_object()[key]

    if PY2:
        __getslice__ = lambda x, i, j: x._get_current_object()[i:j]

        def __setslice__(self, i, j, seq):
            self._get_current_object()[i:j] = seq

        def __delslice__(self, i, j):
            del self._get_current_object()[i:j]

    __setattr__ = lambda x, n, v: setattr(x._get_current_object(), n, v)
    __delattr__ = lambda x, n: delattr(x._get_current_object(), n)
    __str__ = lambda x: str(x._get_current_object())
    __lt__ = lambda x, o: x._get_current_object() < o
    __le__ = lambda x, o: x._get_current_object() <= o
    __eq__ = lambda x, o: x._get_current_object() == o
    __ne__ = lambda x, o: x._get_current_object() != o
    __gt__ = lambda x, o: x._get_current_object() > o
    __ge__ = lambda x, o: x._get_current_object() >= o
    __cmp__ = lambda x, o: cmp(x._get_current_object(), o)
    __hash__ = lambda x: hash(x._get_current_object())
    __call__ = lambda x, *a, **kw: x._get_current_object()(*a, **kw)
    __len__ = lambda x: len(x._get_current_object())
    __getitem__ = lambda x, i: x._get_current_object()[i]
    __iter__ = lambda x: iter(x._get_current_object())
    __contains__ = lambda x, i: i in x._get_current_object()
    __add__ = lambda x, o: x._get_current_object() + o
    __sub__ = lambda x, o: x._get_current_object() - o
    __mul__ = lambda x, o: x._get_current_object() * o
    __floordiv__ = lambda x, o: x._get_current_object() // o
    __mod__ = lambda x, o: x._get_current_object() % o
    __divmod__ = lambda x, o: x._get_current_object().__divmod__(o)
    __pow__ = lambda x, o: x._get_current_object() ** o
    __lshift__ = lambda x, o: x._get_current_object() << o
    __rshift__ = lambda x, o: x._get_current_object() >> o
    __and__ = lambda x, o: x._get_current_object() & o
    __xor__ = lambda x, o: x._get_current_object() ^ o
    __or__ = lambda x, o: x._get_current_object() | o
    __div__ = lambda x, o: x._get_current_object().__div__(o)
    __truediv__ = lambda x, o: x._get_current_object().__truediv__(o)
    __neg__ = lambda x: -(x._get_current_object())
    __pos__ = lambda x: +(x._get_current_object())
    __abs__ = lambda x: abs(x._get_current_object())
    __invert__ = lambda x: ~(x._get_current_object())
    __complex__ = lambda x: complex(x._get_current_object())
    __int__ = lambda x: int(x._get_current_object())
    __long__ = lambda x: long(x._get_current_object())
    __float__ = lambda x: float(x._get_current_object())
    __oct__ = lambda x: oct(x._get_current_object())
    __hex__ = lambda x: hex(x._get_current_object())
    __index__ = lambda x: x._get_current_object().__index__()
    __coerce__ = lambda x, o: x._get_current_object().__coerce__(x, o)
    __enter__ = lambda x: x._get_current_object().__enter__()
    __exit__ = lambda x, *a, **kw: x._get_current_object().__exit__(*a, **kw)
    __radd__ = lambda x, o: o + x._get_current_object()
    __rsub__ = lambda x, o: o - x._get_current_object()
    __rmul__ = lambda x, o: o * x._get_current_object()
    __rdiv__ = lambda x, o: o / x._get_current_object()
    if PY2:
        __rtruediv__ = lambda x, o: x._get_current_object().__rtruediv__(o)
    else:
        __rtruediv__ = __rdiv__
    __rfloordiv__ = lambda x, o: o // x._get_current_object()
    __rmod__ = lambda x, o: o % x._get_current_object()
    __rdivmod__ = lambda x, o: x._get_current_object().__rdivmod__(o)
```

上面的三个类用了很多魔术方法来实现，在这里先不细究魔术方法的原理，有空补上一篇讲解魔术方法的文章。
综合的来看globals.py和local.py的内容，还是从wsgi_app()开始，请求上下文大致逻辑如下：
1. 每次调用app.__call__，即是将当前的请求和环境变量赋予wsgi_app()
2. wsgi()函数如下：
```python
    def wsgi_app(self, environ, start_response):
        ctx = self.request_context(environ)
        ctx.push()
        error = None
        try:
            try:
                response = self.full_dispatch_request()
            except Exception as e:
                error = e
                response = self.handle_exception(e)
            return response(environ, start_response)
        finally:
            if self.should_ignore_error(error):
                error = None
            ctx.auto_pop(error)
```
3. 进栈操作
先看第一行代码：
```python
ctx = self.request_context(environ)
```
self.request_context()
```python
    def request_context(self, environ):
        return RequestContext(self, environ)
```
这里是创建一个请求上下文的实例对象，将当前的环境传给RequestContext()。从RequestContext()的代码实现可以知道是基于globals.py文件中的请求上下文实现的。
```python
_request_ctx_stack = LocalStack()
request = LocalProxy(partial(_lookup_req_object, 'request'))
session = LocalProxy(partial(_lookup_req_object, 'session'))
```
_request_ctx_stack是多线程或者协程隔离的栈结构，request每次都会调用_lookup_req_object栈头部的数据来获取request context。

回过头看wsgi_app()的代码
```python
ctx.push()
```
主要是调用了从RequestContext()类的push()方法:
```python
    def push(self):
        top = _request_ctx_stack.top
        if top is not None and top.preserved:
            top.pop(top._preserved_exc)

        app_ctx = _app_ctx_stack.top
        if app_ctx is None or app_ctx.app != self.app:
            app_ctx = self.app.app_context()
            app_ctx.push()
            self._implicit_app_ctx_stack.append(app_ctx)
        else:
            self._implicit_app_ctx_stack.append(None)

        if hasattr(sys, 'exc_clear'):
            sys.exc_clear()

        _request_ctx_stack.push(self)
        
        self.session = self.app.open_session(self.request)
        if self.session is None:
            self.session = self.app.make_null_session()
```
这里是请求上下文的进栈操作，获取LocalStack()栈结构的最新一条信息,即是Local()的top()属性方法。

4. 出栈操作
然后继续看wsgi_app()函数，try...except...这一段代码是路由匹配和处理逻辑返回的实现，上一篇已经讲过，不细说。
主要是看finally逻辑的代码实现。
```python
ctx.auto_pop(error)
```
```python
    def auto_pop(self, exc):
        if self.request.environ.get('flask._preserve_context') or \
           (exc is not None and self.app.preserve_context_on_exception):
            self.preserved = True
            self._preserved_exc = exc
        else:
            self.pop(exc)
```
```python
    def pop(self, exc=_sentinel):
        app_ctx = self._implicit_app_ctx_stack.pop()

        try:
            clear_request = False
            if not self._implicit_app_ctx_stack:
                self.preserved = False
                self._preserved_exc = None
                if exc is _sentinel:
                    exc = sys.exc_info()[1]
                self.app.do_teardown_request(exc)

                if hasattr(sys, 'exc_clear'):
                    sys.exc_clear()

                request_close = getattr(self.request, 'close', None)
                if request_close is not None:
                    request_close()
                clear_request = True
        finally:
            rv = _request_ctx_stack.pop()

            if clear_request:
                rv.request.environ['werkzeug.request'] = None

            if app_ctx is not None:
                app_ctx.pop(exc)

            assert rv is self, 'Popped wrong request context.  ' \
                '(%r instead of %r)' % (rv, self)
```
非常明显，这是一个请求上下文的出栈操作，删除LocalStack()栈结构的最新一条信息调用了Local()的pop()方法来实现。

到这里，上下文处理逻辑就比较清晰了，每次新的请求，flask先创建当前线程或者进程需要处理的应用上下文和请求上下文对象，保存到对应隔离的栈里面。
每个上下文都保存了当前请求的信息，在初始化后，实现路由匹配，视图函数处理对应逻辑时，可直接从栈上获取这些上下文信息，处理完毕后，将应用上下文和请求上下文出栈，做清理工作。

## 四. 抽象
一个request请求进来后，上下文处理逻辑如下：
```python
wsgi_app()[flask/app.py]-->request_context()[flask/app.py]-->RequestContext.__init__()[flask/ctx.py]-->
push()[flask/ctx.py]-->_request_ctx_stack[flask/globals.py]--LocalStack.push()[werkzeug/local.py]-->
pop()[flask/ctx.py]-->LocalStack.pop()[werkzeug/local.py]
```

参考：
[flask 上下文的实现](https://segmentfault.com/a/1190000004223296)
[上下文（application context 和 request context）](http://cizixs.com/2017/01/13/flask-insight-context)
