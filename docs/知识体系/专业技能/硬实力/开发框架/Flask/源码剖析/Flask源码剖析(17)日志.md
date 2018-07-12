---
layout: blog
title: 'Flask源码剖析（十七）：日志'
date: 2017-09-09 09:01:15
categories: flask
tags: flask
lead_text: '日志'
---

## 一. 日志
对于大部分程序员来说，20%的时间在写代码逻辑，80%的时间在调试。
如何提高debug的效率呢？那么最重要的一点是，快读定位问题，从已有的记录信息中得到程序崩溃的内容，对此，Flask程序提供了一个logging模块来记录这些信息。

## 二. 代码
Flask框架提供记录日志的模块是flask/logging.py，来看看是怎么实现的。
```python
# -*- coding: utf-8 -*-

from __future__ import absolute_import

import sys

from werkzeug.local import LocalProxy
from logging import getLogger, StreamHandler, Formatter, getLoggerClass, \
     DEBUG, ERROR
from .globals import _request_ctx_stack


PROD_LOG_FORMAT = '[%(asctime)s] %(levelname)s in %(module)s: %(message)s'
DEBUG_LOG_FORMAT = (
    '-' * 80 + '\n' +
    '%(levelname)s in %(module)s [%(pathname)s:%(lineno)d]:\n' +
    '%(message)s\n' +
    '-' * 80
)


@LocalProxy
def _proxy_stream():
    ctx = _request_ctx_stack.top
    if ctx is not None:
        return ctx.request.environ['wsgi.errors']
    return sys.stderr


def _should_log_for(app, mode):
    policy = app.config['LOGGER_HANDLER_POLICY']
    if policy == mode or policy == 'always':
        return True
    return False


def create_logger(app):
    Logger = getLoggerClass()

    class DebugLogger(Logger):
        def getEffectiveLevel(self):
            if self.level == 0 and app.debug:
                return DEBUG
            return Logger.getEffectiveLevel(self)

    class DebugHandler(StreamHandler):
        def emit(self, record):
            if app.debug and _should_log_for(app, 'debug'):
                StreamHandler.emit(self, record)

    class ProductionHandler(StreamHandler):
        def emit(self, record):
            if not app.debug and _should_log_for(app, 'production'):
                StreamHandler.emit(self, record)

    debug_handler = DebugHandler()
    debug_handler.setLevel(DEBUG)
    debug_handler.setFormatter(Formatter(DEBUG_LOG_FORMAT))

    prod_handler = ProductionHandler(_proxy_stream)
    prod_handler.setLevel(ERROR)
    prod_handler.setFormatter(Formatter(PROD_LOG_FORMAT))

    logger = getLogger(app.logger_name)
    
    del logger.handlers[:]
    logger.__class__ = DebugLogger
    logger.addHandler(debug_handler)
    logger.addHandler(prod_handler)

    logger.propagate = False

    return logger

```

## 三. 解析
flask/logging.py文件主要在原生logging模块的基础上重新封装一层，提供了create_logger函数，即是创建日志记录器，来看看是怎么实现的。
- 调用getLoggerClass()创建一个Logger类
- DebugLogger类继承Logger，返回日志的级别，分别为NOTSET/DEBUG/INFO/WARN/WARNING/ERROR/FATAL/CRITICAL
- DebugHandler类继承StreamHandler，获取环境变量LOGGER_HANDLER_POLICY，如果LOGGER_HANDLER_POLICY为debug，返回定义好的日志格式DEBUG_LOG_FORMAT
- ProductionHandler类继承StreamHandler，获取环境变量LOGGER_HANDLER_POLICY，如果LOGGER_HANDLER_POLICY为production，返回定义好的日志格式PROD_LOG_FORMAT

## 四. 应用
Flask框架的日志使用可参考[记录应用错误](http://docs.jinkan.org/docs/flask/errorhandling.html)