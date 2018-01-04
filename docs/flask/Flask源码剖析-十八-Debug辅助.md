---
layout: blog
title: 'Flask源码剖析（十八）：Debug辅助'
date: 2017-09-13 19:32:12
categories: flask
tags: flask
lead_text: 'Debug辅助'
---

## 一. Debug辅助
Flask框架提供了在debug模式下的一些辅助函数，主要是抛出异常的各种提示。

## 二. 代码
debug模式下的辅助函数主要放在flask/debughelpers.py，看看是怎么实现的
```python
# -*- coding: utf-8 -*-
from ._compat import implements_to_string, text_type
from .app import Flask
from .blueprints import Blueprint
from .globals import _request_ctx_stack


class UnexpectedUnicodeError(AssertionError, UnicodeError):
    """Raised in places where we want some better error reporting for
    unexpected unicode or binary data.
    """


@implements_to_string
class DebugFilesKeyError(KeyError, AssertionError):
    """Raised from request.files during debugging.  The idea is that it can
    provide a better error message than just a generic KeyError/BadRequest.
    """

    def __init__(self, request, key):
        form_matches = request.form.getlist(key)
        buf = ['You tried to access the file "%s" in the request.files '
               'dictionary but it does not exist.  The mimetype for the request '
               'is "%s" instead of "multipart/form-data" which means that no '
               'file contents were transmitted.  To fix this error you should '
               'provide enctype="multipart/form-data" in your form.' %
               (key, request.mimetype)]
        if form_matches:
            buf.append('\n\nThe browser instead transmitted some file names. '
                       'This was submitted: %s' % ', '.join('"%s"' % x
                            for x in form_matches))
        self.msg = ''.join(buf)

    def __str__(self):
        return self.msg


class FormDataRoutingRedirect(AssertionError):
    """This exception is raised by Flask in debug mode if it detects a
    redirect caused by the routing system when the request method is not
    GET, HEAD or OPTIONS.  Reasoning: form data will be dropped.
    """

    def __init__(self, request):
        exc = request.routing_exception
        buf = ['A request was sent to this URL (%s) but a redirect was '
               'issued automatically by the routing system to "%s".'
               % (request.url, exc.new_url)]

        # In case just a slash was appended we can be extra helpful
        if request.base_url + '/' == exc.new_url.split('?')[0]:
            buf.append('  The URL was defined with a trailing slash so '
                       'Flask will automatically redirect to the URL '
                       'with the trailing slash if it was accessed '
                       'without one.')

        buf.append('  Make sure to directly send your %s-request to this URL '
                   'since we can\'t make browsers or HTTP clients redirect '
                   'with form data reliably or without user interaction.' %
                   request.method)
        buf.append('\n\nNote: this exception is only raised in debug mode')
        AssertionError.__init__(self, ''.join(buf).encode('utf-8'))


def attach_enctype_error_multidict(request):
    """Since Flask 0.8 we're monkeypatching the files object in case a
    request is detected that does not use multipart form data but the files
    object is accessed.
    """
    oldcls = request.files.__class__
    class newcls(oldcls):
        def __getitem__(self, key):
            try:
                return oldcls.__getitem__(self, key)
            except KeyError:
                if key not in request.form:
                    raise
                raise DebugFilesKeyError(request, key)
    newcls.__name__ = oldcls.__name__
    newcls.__module__ = oldcls.__module__
    request.files.__class__ = newcls


def _dump_loader_info(loader):
    yield 'class: %s.%s' % (type(loader).__module__, type(loader).__name__)
    for key, value in sorted(loader.__dict__.items()):
        if key.startswith('_'):
            continue
        if isinstance(value, (tuple, list)):
            if not all(isinstance(x, (str, text_type)) for x in value):
                continue
            yield '%s:' % key
            for item in value:
                yield '  - %s' % item
            continue
        elif not isinstance(value, (str, text_type, int, float, bool)):
            continue
        yield '%s: %r' % (key, value)


def explain_template_loading_attempts(app, template, attempts):
    """This should help developers understand what failed"""
    info = ['Locating template "%s":' % template]
    total_found = 0
    blueprint = None
    reqctx = _request_ctx_stack.top
    if reqctx is not None and reqctx.request.blueprint is not None:
        blueprint = reqctx.request.blueprint

    for idx, (loader, srcobj, triple) in enumerate(attempts):
        if isinstance(srcobj, Flask):
            src_info = 'application "%s"' % srcobj.import_name
        elif isinstance(srcobj, Blueprint):
            src_info = 'blueprint "%s" (%s)' % (srcobj.name,
                                                srcobj.import_name)
        else:
            src_info = repr(srcobj)

        info.append('% 5d: trying loader of %s' % (
            idx + 1, src_info))

        for line in _dump_loader_info(loader):
            info.append('       %s' % line)

        if triple is None:
            detail = 'no match'
        else:
            detail = 'found (%r)' % (triple[1] or '<string>')
            total_found += 1
        info.append('       -> %s' % detail)

    seems_fishy = False
    if total_found == 0:
        info.append('Error: the template could not be found.')
        seems_fishy = True
    elif total_found > 1:
        info.append('Warning: multiple loaders returned a match for the template.')
        seems_fishy = True

    if blueprint is not None and seems_fishy:
        info.append('  The template was looked up from an endpoint that '
                    'belongs to the blueprint "%s".' % blueprint)
        info.append('  Maybe you did not place a template in the right folder?')
        info.append('  See http://flask.pocoo.org/docs/blueprints/#templates')

    app.logger.info('\n'.join(info))

```
## 三. 解析
flask/debughelpers.py文件中主要提供外部调用的UnexpectedUnicodeError/FormDataRoutingRedirect类，attach_enctype_error_multidict/explain_template_loading_attempts函数
1. UnexpectedUnicodeError
- UnexpectedUnicodeError类继承了AssertionError和UnicodeError，没做其他修饰

2. FormDataRoutingRedirect
- FormDataRoutingRedirect类继承AssertionError，重写了__init__方法,接受request请求上下文为参数，添加了异常抛出说明

3. attach_enctype_error_multidict
- attach_enctype_error_multidict方法接受当前请求上下文为参数，这个方法内定义了一个newcls类，判断请求上下文中的数据是否符合字典，并且包含键值对，反之则返回报错，并输出异常抛出说明

4. explain_template_loading_attempts
- explain_template_loading_attempts方法接受当前程序实例，jinja2模板加载器和attempts列表为参数
- attempts列表是个两层嵌套列表，每个嵌套列表中的第一个值是jinja模板加载器,第二个参数是Flask实例或者Blueprint实例，第三个参数是是否获取到了对应的模板
- explain_template_loading_attempts方法根据三个参数的不同，从而拼接不同的输出说明，调用logger来输出
