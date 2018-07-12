---
layout: blog
title: 'Flask源码剖析（十九）：兼容'
date: 2017-09-17 20:21:09
categories: flask
tags: flask
lead_text: '兼容'
---

## 一. 兼容
Flask框架兼容与py2和py3版本，来看看是怎么实现的。

## 二. 代码
Flask框架提供python不同版本的兼容的实现放在flask/_compat.py文件中，如下
```python
# -*- coding: utf-8 -*-
import sys

PY2 = sys.version_info[0] == 2
_identity = lambda x: x


if not PY2:
    text_type = str
    string_types = (str,)
    integer_types = (int,)

    iterkeys = lambda d: iter(d.keys())
    itervalues = lambda d: iter(d.values())
    iteritems = lambda d: iter(d.items())

    from io import StringIO

    def reraise(tp, value, tb=None):
        if value.__traceback__ is not tb:
            raise value.with_traceback(tb)
        raise value

    implements_to_string = _identity

else:
    text_type = unicode
    string_types = (str, unicode)
    integer_types = (int, long)

    iterkeys = lambda d: d.iterkeys()
    itervalues = lambda d: d.itervalues()
    iteritems = lambda d: d.iteritems()

    from cStringIO import StringIO

    exec('def reraise(tp, value, tb=None):\n raise tp, value, tb')

    def implements_to_string(cls):
        cls.__unicode__ = cls.__str__
        cls.__str__ = lambda x: x.__unicode__().encode('utf-8')
        return cls


def with_metaclass(meta, *bases):
    class metaclass(type):
        def __new__(cls, name, this_bases, d):
            return meta(name, bases, d)
    return type.__new__(metaclass, 'temporary_class', (), {})

BROKEN_PYPY_CTXMGR_EXIT = False
if hasattr(sys, 'pypy_version_info'):
    class _Mgr(object):
        def __enter__(self):
            return self
        def __exit__(self, *args):
            if hasattr(sys, 'exc_clear'):
                # Python 3 (PyPy3) doesn't have exc_clear
                sys.exc_clear()
    try:
        try:
            with _Mgr():
                raise AssertionError()
        except:
            raise
    except TypeError:
        BROKEN_PYPY_CTXMGR_EXIT = True
    except AssertionError:
        pass

```

## 三. 解析
Flask框架的py2/py3兼容主要是依赖于six模块，_compat.py文件主要是提供了utf-8编码和字符串格式的兼容，另外还有Ubuntu 14.04对PyPy 2.2.1的支持问题。
- py2的文件需要注明是utf-8编码,py3默认支持utf-8
- py3不再支持iterkeys()，使用iter()将keys()的返回值转换成为一个迭代器
- py2有为非浮点数准备的int和long类型, 在py3里，只有一种整数类型int

具体的py2和py3的语法区别可参考[python2 与 python3 语法区别](http://blog.csdn.net/samxx8/article/details/21535901)
