---
layout: blog
title: 'Flask源码剖析（十二）：配置'
date: 2017-08-06 20:49:45
categories: flask
tags: flask
lead_text: '配置'
---

## 一. 配置
应用程序需要某种形式的配置,可能会需要根据应用环境更改不同的设置。
Flask被设计为需要配置来启动应用。
有一个配置对象用来维持加载的配置值：Flask对象的config属性。

## 二. 代码
config属性主要是在flask/config.py文件中定义的,来看看。
```python
# -*- coding: utf-8 -*-

import os
import types
import errno

from werkzeug.utils import import_string
from ._compat import string_types, iteritems
from . import json


class ConfigAttribute(object):

    def __init__(self, name, get_converter=None):
        self.__name__ = name
        self.get_converter = get_converter

    def __get__(self, obj, type=None):
        if obj is None:
            return self
        rv = obj.config[self.__name__]
        if self.get_converter is not None:
            rv = self.get_converter(rv)
        return rv

    def __set__(self, obj, value):
        obj.config[self.__name__] = value


class Config(dict):

    def __init__(self, root_path, defaults=None):
        dict.__init__(self, defaults or {})
        self.root_path = root_path

    def from_envvar(self, variable_name, silent=False):
        rv = os.environ.get(variable_name)
        if not rv:
            if silent:
                return False
            raise RuntimeError('The environment variable %r is not set '
                               'and as such configuration could not be '
                               'loaded.  Set this variable and make it '
                               'point to a configuration file' %
                               variable_name)
        return self.from_pyfile(rv, silent=silent)

    def from_pyfile(self, filename, silent=False):
        filename = os.path.join(self.root_path, filename)
        d = types.ModuleType('config')
        d.__file__ = filename
        try:
            with open(filename) as config_file:
                exec(compile(config_file.read(), filename, 'exec'), d.__dict__)
        except IOError as e:
            if silent and e.errno in (errno.ENOENT, errno.EISDIR):
                return False
            e.strerror = 'Unable to load configuration file (%s)' % e.strerror
            raise
        self.from_object(d)
        return True

    def from_object(self, obj):
        if isinstance(obj, string_types):
            obj = import_string(obj)
        for key in dir(obj):
            if key.isupper():
                self[key] = getattr(obj, key)

    def from_json(self, filename, silent=False):
        filename = os.path.join(self.root_path, filename)

        try:
            with open(filename) as json_file:
                obj = json.loads(json_file.read())
        except IOError as e:
            if silent and e.errno in (errno.ENOENT, errno.EISDIR):
                return False
            e.strerror = 'Unable to load configuration file (%s)' % e.strerror
            raise
        return self.from_mapping(obj)

    def from_mapping(self, *mapping, **kwargs):
        mappings = []
        if len(mapping) == 1:
            if hasattr(mapping[0], 'items'):
                mappings.append(mapping[0].items())
            else:
                mappings.append(mapping[0])
        elif len(mapping) > 1:
            raise TypeError(
                'expected at most 1 positional argument, got %d' % len(mapping)
            )
        mappings.append(kwargs.items())
        for mapping in mappings:
            for (key, value) in mapping:
                if key.isupper():
                    self[key] = value
        return True

    def get_namespace(self, namespace, lowercase=True, trim_namespace=True):
        rv = {}
        for k, v in iteritems(self):
            if not k.startswith(namespace):
                continue
            if trim_namespace:
                key = k[len(namespace):]
            else:
                key = k
            if lowercase:
                key = key.lower()
            rv[key] = v
        return rv

    def __repr__(self):
        return '<%s %s>' % (self.__class__.__name__, dict.__repr__(self))

```

## 三. 解析
flask/config.py文件定义了两个类ConfigAttribute/Config。
1. ConfigAttribute类
ConfigAttribute类使用了__get__和__set__魔术方法，给config属性赋予方法，其中包括以下几种方法
```python
    debug = ConfigAttribute('DEBUG')

    testing = ConfigAttribute('TESTING')

    secret_key = ConfigAttribute('SECRET_KEY')

    session_cookie_name = ConfigAttribute('SESSION_COOKIE_NAME')

    permanent_session_lifetime = ConfigAttribute('PERMANENT_SESSION_LIFETIME',
        get_converter=_make_timedelta)

    send_file_max_age_default = ConfigAttribute('SEND_FILE_MAX_AGE_DEFAULT',
        get_converter=_make_timedelta)

    use_x_sendfile = ConfigAttribute('USE_X_SENDFILE')

    logger_name = ConfigAttribute('LOGGER_NAME')
```
2. Config类
Config类提供了几种获取配置的函数方法
- from_envvar（环境变量）
- from_pyfile（py文件）
- from_object（模块）
- from_json（json数据）
- from_mapping（哈希对象）
- get_namespace（命名空间）

## 四. 应用
应用可以参考[配置处理](http://www.pythondoc.com/flask/config.html)