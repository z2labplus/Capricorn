---
layout: blog
title: 'Flask源码剖析（十一）：模板'
date: 2017-07-29 12:47:51
categories: flask
tags: flask
lead_text: '模板'
---

## 一. 模板
Flask默认使用Jinja2作为模板引擎，当然也可以选择使用其他模板引擎。
模板是一个包含响应文本的文件，其中包含用占位变量表示的动态部分, 具体值只在收到具体的请求后，通过上下文才能知道。
模板引擎实现对模板的渲染，就是根据上下文，对模板中的占位变量，用真实值替换，形成最终的响应文件。

## 二. 代码
Flask连接Jinja2模板引擎的代码放在flask/templating.py文件中，如下：
```python
# -*- coding: utf-8 -*-
from jinja2 import BaseLoader, Environment as BaseEnvironment, \
     TemplateNotFound

from .globals import _request_ctx_stack, _app_ctx_stack
from .signals import template_rendered, before_render_template


def _default_template_ctx_processor():
    reqctx = _request_ctx_stack.top
    appctx = _app_ctx_stack.top
    rv = {}
    if appctx is not None:
        rv['g'] = appctx.g
    if reqctx is not None:
        rv['request'] = reqctx.request
        rv['session'] = reqctx.session
    return rv


class Environment(BaseEnvironment):
    def __init__(self, app, **options):
        if 'loader' not in options:
            options['loader'] = app.create_global_jinja_loader()
        BaseEnvironment.__init__(self, **options)
        self.app = app

class DispatchingJinjaLoader(BaseLoader):

    def __init__(self, app):
        self.app = app

    def get_source(self, environment, template):
        if self.app.config['EXPLAIN_TEMPLATE_LOADING']:
            return self._get_source_explained(environment, template)
        return self._get_source_fast(environment, template)

    def _get_source_explained(self, environment, template):
        attempts = []
        trv = None

        for srcobj, loader in self._iter_loaders(template):
            try:
                rv = loader.get_source(environment, template)
                if trv is None:
                    trv = rv
            except TemplateNotFound:
                rv = None
            attempts.append((loader, srcobj, rv))

        from .debughelpers import explain_template_loading_attempts
        explain_template_loading_attempts(self.app, template, attempts)

        if trv is not None:
            return trv
        raise TemplateNotFound(template)

    def _get_source_fast(self, environment, template):
        for srcobj, loader in self._iter_loaders(template):
            try:
                return loader.get_source(environment, template)
            except TemplateNotFound:
                continue
        raise TemplateNotFound(template)

    def _iter_loaders(self, template):
        loader = self.app.jinja_loader
        if loader is not None:
            yield self.app, loader

        for blueprint in self.app.iter_blueprints():
            loader = blueprint.jinja_loader
            if loader is not None:
                yield blueprint, loader

    def list_templates(self):
        result = set()
        loader = self.app.jinja_loader
        if loader is not None:
            result.update(loader.list_templates())

        for blueprint in self.app.iter_blueprints():
            loader = blueprint.jinja_loader
            if loader is not None:
                for template in loader.list_templates():
                    result.add(template)

        return list(result)


def _render(template, context, app):
    before_render_template.send(app, template=template, context=context)
    rv = template.render(context)
    template_rendered.send(app, template=template, context=context)
    return rv


def render_template(template_name_or_list, **context):
    ctx = _app_ctx_stack.top
    ctx.app.update_template_context(context)
    return _render(ctx.app.jinja_env.get_or_select_template(template_name_or_list),
                   context, ctx.app)


def render_template_string(source, **context):
    ctx = _app_ctx_stack.top
    ctx.app.update_template_context(context)
    return _render(ctx.app.jinja_env.from_string(source),
                   context, ctx.app)
```

## 三. 解析
flask/templating.py文件提供外部调用的函数和类有_default_template_ctx_processor/Environment/DispatchingJinjaLoader/render_template/render_template_string,
此中，flask/app.py文件中调用的是_default_template_ctx_processor/Environment/DispatchingJinjaLoader，来看看是怎么调用的。
```python
from .templating import DispatchingJinjaLoader, Environment, \
     _default_template_ctx_processor
     
class Flask(_PackageBoundObject):

    self.template_context_processors = {
            None: [_default_template_ctx_processor]
        }
        
    jinja_environment = Environment
    
    @locked_cached_property
    def jinja_env(self):
        return self.create_jinja_environment()
        
    def create_jinja_environment(self):
        options = dict(self.jinja_options)
        if 'autoescape' not in options:
            options['autoescape'] = self.select_jinja_autoescape
        if 'auto_reload' not in options:
            if self.config['TEMPLATES_AUTO_RELOAD'] is not None:
                options['auto_reload'] = self.config['TEMPLATES_AUTO_RELOAD']
            else:
                options['auto_reload'] = self.debug
        rv = self.jinja_environment(self, **options)
        rv.globals.update(
            url_for=url_for,
            get_flashed_messages=get_flashed_messages,
            config=self.config,
            request=request,
            session=session,
            g=g
        )
        rv.filters['tojson'] = json.tojson_filter
        return rv
        
        

    def create_global_jinja_loader(self):
        return DispatchingJinjaLoader(self)
```
很明显，当一个flask程序运行时，将创建jijin2的全局变量，关键字是url_for/get_flashed_messages/config/request/session/g。
接下来看看视图函数中是怎么调用render_template/render_template_string，这里着重说明下render_template。
1. 从_app_ctx_stack上下文中获取最新的上下文实例化对象ctx
2. 更新ctx中的模板上下文数据
```python
    def update_template_context(self, context):
        funcs = self.template_context_processors[None]
        reqctx = _request_ctx_stack.top
        if reqctx is not None:
            bp = reqctx.request.blueprint
            if bp is not None and bp in self.template_context_processors:
                funcs = chain(funcs, self.template_context_processors[bp])
        orig_ctx = context.copy()
        for func in funcs:
            context.update(func())
        context.update(orig_ctx)
```
3. 返回渲染后的模板，来看看模板的数据是怎么获取的
```python
ctx.app.jinja_env.get_or_select_template(template_name_or_list)
```
从上面可以知道，jinja_env是locked_cached_property装饰器装饰的
那么，locked_cached_property是怎么定义的呢
```python
class locked_cached_property(object):
    """A decorator that converts a function into a lazy property.  The
    function wrapped is called the first time to retrieve the result
    and then that calculated result is used the next time you access
    the value.  Works like the one in Werkzeug but has a lock for
    thread safety.
    """

    def __init__(self, func, name=None, doc=None):
        self.__name__ = name or func.__name__
        self.__module__ = func.__module__
        self.__doc__ = doc or func.__doc__
        self.func = func
        self.lock = RLock()

    def __get__(self, obj, type=None):
        if obj is None:
            return self
        with self.lock:
            value = obj.__dict__.get(self.__name__, _missing)
            if value is _missing:
                value = self.func(obj)
                obj.__dict__[self.__name__] = value
            return value
```
jinja_env是个描述符，访问jinja_env属性的时候，会调用get方法，obj设置为ctx.app。
obj.dict中并没有self.name，到value is _missing，执行原始函数value = ctx.app.create_jinja_environment()，并设置该属性到ctx.app中。

## 四. 应用
```python
from flask import Flask, render_template

app = Flask(__name__)

@app.route('/')
def hello():
    name = 'hello world'
    return render_template('hello_world.html', name=name)
```