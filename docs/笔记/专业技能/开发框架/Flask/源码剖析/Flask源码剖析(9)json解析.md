---
layout: blog
title: 'Flask源码剖析（九）：json解析'
date: 2017-07-13 19:24:35
categories: flask
tags: flask
lead_text: 'json解析'
---

## 一. 什么是json
> JSON(JavaScript Object Notation, JS 对象标记) 是一种轻量级的数据交换格式。
> 它基于 ECMAScript (w3c制定的js规范)的一个子集，采用完全独立于编程语言的文本格式来存储和表示数据。
> 简洁和清晰的层次结构使得 JSON 成为理想的数据交换语言。 
> 易于人阅读和编写，同时也易于机器解析和生成，并有效地提升网络传输效率。
> --[百度百科](https://baike.baidu.com/item/JSON/2462549?fr=aladdin)

## 二. 代码
Flask在处理json数据时，对json数据加了一层封装，返回一个response对象。
看一个简单的例子：
```python
# -*- coding: utf-8 -*-

from flask import Flask, jsonify
app = Flask(__name__)

@app.route('/json')
def hello():
    return jsonify({'msg': 'Hello World', 'code': 200})

if __name__ == '__main__':
    app.run()
```

Flask的json实现都在flask/json.py文件中，来看看是怎么实现的。
```python
# -*- coding: utf-8 -*-
import io
import uuid
from datetime import date
from .globals import current_app, request
from ._compat import text_type, PY2

from werkzeug.http import http_date
from jinja2 import Markup

from itsdangerous import json as _json


_slash_escape = '\\/' not in _json.dumps('/')


__all__ = ['dump', 'dumps', 'load', 'loads', 'htmlsafe_dump',
           'htmlsafe_dumps', 'JSONDecoder', 'JSONEncoder',
           'jsonify']


def _wrap_reader_for_text(fp, encoding):
    if isinstance(fp.read(0), bytes):
        fp = io.TextIOWrapper(io.BufferedReader(fp), encoding)
    return fp


def _wrap_writer_for_text(fp, encoding):
    try:
        fp.write('')
    except TypeError:
        fp = io.TextIOWrapper(fp, encoding)
    return fp


class JSONEncoder(_json.JSONEncoder):

    def default(self, o):
        if isinstance(o, date):
            return http_date(o.timetuple())
        if isinstance(o, uuid.UUID):
            return str(o)
        if hasattr(o, '__html__'):
            return text_type(o.__html__())
        return _json.JSONEncoder.default(self, o)


class JSONDecoder(_json.JSONDecoder):


def _dump_arg_defaults(kwargs):
    if current_app:
        kwargs.setdefault('cls', current_app.json_encoder)
        if not current_app.config['JSON_AS_ASCII']:
            kwargs.setdefault('ensure_ascii', False)
        kwargs.setdefault('sort_keys', current_app.config['JSON_SORT_KEYS'])
    else:
        kwargs.setdefault('sort_keys', True)
        kwargs.setdefault('cls', JSONEncoder)


def _load_arg_defaults(kwargs):
    if current_app:
        kwargs.setdefault('cls', current_app.json_decoder)
    else:
        kwargs.setdefault('cls', JSONDecoder)


def dumps(obj, **kwargs):
    _dump_arg_defaults(kwargs)
    encoding = kwargs.pop('encoding', None)
    rv = _json.dumps(obj, **kwargs)
    if encoding is not None and isinstance(rv, text_type):
        rv = rv.encode(encoding)
    return rv


def dump(obj, fp, **kwargs):
    _dump_arg_defaults(kwargs)
    encoding = kwargs.pop('encoding', None)
    if encoding is not None:
        fp = _wrap_writer_for_text(fp, encoding)
    _json.dump(obj, fp, **kwargs)


def loads(s, **kwargs):
    _load_arg_defaults(kwargs)
    if isinstance(s, bytes):
        s = s.decode(kwargs.pop('encoding', None) or 'utf-8')
    return _json.loads(s, **kwargs)


def load(fp, **kwargs):
    _load_arg_defaults(kwargs)
    if not PY2:
        fp = _wrap_reader_for_text(fp, kwargs.pop('encoding', None) or 'utf-8')
    return _json.load(fp, **kwargs)


def htmlsafe_dumps(obj, **kwargs):
    rv = dumps(obj, **kwargs) \
        .replace(u'<', u'\\u003c') \
        .replace(u'>', u'\\u003e') \
        .replace(u'&', u'\\u0026') \
        .replace(u"'", u'\\u0027')
    if not _slash_escape:
        rv = rv.replace('\\/', '/')
    return rv


def htmlsafe_dump(obj, fp, **kwargs):
    fp.write(text_type(htmlsafe_dumps(obj, **kwargs)))


def jsonify(*args, **kwargs):
    indent = None
    separators = (',', ':')

    if current_app.config['JSONIFY_PRETTYPRINT_REGULAR'] and not request.is_xhr:
        indent = 2
        separators = (', ', ': ')

    if args and kwargs:
        raise TypeError('jsonify() behavior undefined when passed both args and kwargs')
    elif len(args) == 1:  # single args are passed directly to dumps()
        data = args[0]
    else:
        data = args or kwargs

    return current_app.response_class(
        (dumps(data, indent=indent, separators=separators), '\n'),
        mimetype=current_app.config['JSONIFY_MIMETYPE']
    )


def tojson_filter(obj, **kwargs):
    return Markup(htmlsafe_dumps(obj, **kwargs))

```
Flask的json实现主要是使用了itsdangerous中的json对象，来自于simplejson或者是内置的json，先不深入研究itsdangerous的代码。

## 三. 解析
1. 对于JSONEncoder，在_json.JSONEncoder的基础上做了一点小修改，增加了对date，UUID和内置__html__属性（如flask.Markup）等的处理
```python
class JSONEncoder(_json.JSONEncoder):

    def default(self, o):
        if isinstance(o, date):
            return http_date(o.timetuple())
        if isinstance(o, uuid.UUID):
            return str(o)
        if hasattr(o, '__html__'):
            return text_type(o.__html__())
        return _json.JSONEncoder.default(self, o)
```
2. 对于JSONDecoder，继承了_json.JSONDecoder没有做修改
3. 对于基本的dump/dumps/load/loads等方法，Flask提供了两个方法来处理unicode编码转换，此外，对于字节流，Flask进行了封装返回unicode文本流，再交给json处理
```python
def _wrap_reader_for_text(fp, encoding):
    if isinstance(fp.read(0), bytes):
        fp = io.TextIOWrapper(io.BufferedReader(fp), encoding)
    return fp


def _wrap_writer_for_text(fp, encoding):
    try:
        fp.write('')
    except TypeError:
        fp = io.TextIOWrapper(fp, encoding)
    return fp
```
4. 对于html，Flask提供了特殊字符转换处理
```python
def htmlsafe_dumps(obj, **kwargs):
    rv = dumps(obj, **kwargs) \
        .replace(u'<', u'\\u003c') \
        .replace(u'>', u'\\u003e') \
        .replace(u'&', u'\\u0026') \
        .replace(u"'", u'\\u0027')
    if not _slash_escape:
        rv = rv.replace('\\/', '/')
    return rv


def htmlsafe_dump(obj, fp, **kwargs):
    fp.write(text_type(htmlsafe_dumps(obj, **kwargs)))
```
5. jsonify函数在dumps函数返回结果的基础上，调用response_class对象对结果进行了包装，给返回的数据头部的mimetype赋值json类型（默认为application/json）