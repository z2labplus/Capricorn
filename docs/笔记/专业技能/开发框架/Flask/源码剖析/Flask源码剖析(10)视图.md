---
layout: blog
title: 'Flask源码剖析（十）：视图'
date: 2017-07-20 18:54:01
categories: flask
tags: flask
lead_text: '视图'
---

## 一. 视图
之前实现的程序都是基于函数通用视图，这里说的视图指的是Flask提供的视图类，这就是即插视图。
> 灵感来自Django的基于类而不是函数的通用视图。 
> 其主要目的是让你可以对已实现的部分进行替换，并且这个方式可以定制即插视图。
> -- [即插视图](http://docs.jinkan.org/docs/flask/views.html)

## 二. 代码
```python
# -*- coding: utf-8 -*-
from .globals import request
from ._compat import with_metaclass


http_method_funcs = frozenset(['get', 'post', 'head', 'options',
                               'delete', 'put', 'trace', 'patch'])


class View(object):
    methods = None

    decorators = ()

    def dispatch_request(self):
        raise NotImplementedError()

    @classmethod
    def as_view(cls, name, *class_args, **class_kwargs):
        def view(*args, **kwargs):
            self = view.view_class(*class_args, **class_kwargs)
            return self.dispatch_request(*args, **kwargs)

        if cls.decorators:
            view.__name__ = name
            view.__module__ = cls.__module__
            for decorator in cls.decorators:
                view = decorator(view)

        view.view_class = cls
        view.__name__ = name
        view.__doc__ = cls.__doc__
        view.__module__ = cls.__module__
        view.methods = cls.methods
        return view


class MethodViewType(type):

    def __new__(cls, name, bases, d):
        rv = type.__new__(cls, name, bases, d)
        if 'methods' not in d:
            methods = set(rv.methods or [])
            for key in d:
                if key in http_method_funcs:
                    methods.add(key.upper())
            if methods:
                rv.methods = sorted(methods)
        return rv


class MethodView(with_metaclass(MethodViewType, View)):
    def dispatch_request(self, *args, **kwargs):
        meth = getattr(self, request.method.lower(), None)
        if meth is None and request.method == 'HEAD':
            meth = getattr(self, 'get', None)
        assert meth is not None, 'Unimplemented method %r' % request.method
        return meth(*args, **kwargs)

```
## 三. 解析
上面的代码中定义了三个类:View/MethodViewType/MethodView，分别看看这三个类分别是怎么定义的
1. View类定义了一个实例方法dispatch_request和类方法as_view，dispatch_request方法并没有实现任何逻辑，as_view方法接受了参数，重新封装了dispatch_request方法，返回一个view实例对象。
2. MethodViewType类继承了type类，为rv.methods属性赋值，值为['get', 'post', 'head', 'options', 'delete', 'put', 'trace', 'patch']
3. MethodView类继承了MethodViewType类和View类，重写了dispatch_request方法

## 四. 使用
具体的使用方法可以参考[即插视图](http://docs.jinkan.org/docs/flask/views.html),此处就不再另外举例了。