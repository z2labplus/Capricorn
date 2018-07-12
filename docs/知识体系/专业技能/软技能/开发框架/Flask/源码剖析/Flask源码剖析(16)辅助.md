---
layout: blog
title: 'Flask源码剖析（十六）：辅助'
date: 2017-08-31 18:50:33
categories: flask
tags: flask
lead_text: '辅助'
---

## 一. 辅助
Flask框架提供了很多辅助函数，stream_with_context/make_response/url_for等，为其他功能的实现提供了公共函数。

## 二. 代码
这些辅助函数的实现主要放在flask/helpers.py文件，如下：
```python
# -*- coding: utf-8 -*-

import os
import sys
import pkgutil
import posixpath
import mimetypes
from time import time
from zlib import adler32
from threading import RLock
from werkzeug.routing import BuildError
from functools import update_wrapper

try:
    from werkzeug.urls import url_quote
except ImportError:
    from urlparse import quote as url_quote

from werkzeug.datastructures import Headers, Range
from werkzeug.exceptions import BadRequest, NotFound, \
    RequestedRangeNotSatisfiable

try:
    from werkzeug.wsgi import wrap_file
except ImportError:
    from werkzeug.utils import wrap_file

from jinja2 import FileSystemLoader

from .signals import message_flashed
from .globals import session, _request_ctx_stack, _app_ctx_stack, \
     current_app, request
from ._compat import string_types, text_type


_missing = object()


_os_alt_seps = list(sep for sep in [os.path.sep, os.path.altsep]
                    if sep not in (None, '/'))


def get_debug_flag(default=None):
    val = os.environ.get('FLASK_DEBUG')
    if not val:
        return default
    return val not in ('0', 'false', 'no')


def _endpoint_from_view_func(view_func):
    assert view_func is not None, 'expected view func if endpoint ' \
                                  'is not provided.'
    return view_func.__name__


def stream_with_context(generator_or_function):
    try:
        gen = iter(generator_or_function)
    except TypeError:
        def decorator(*args, **kwargs):
            gen = generator_or_function(*args, **kwargs)
            return stream_with_context(gen)
        return update_wrapper(decorator, generator_or_function)

    def generator():
        ctx = _request_ctx_stack.top
        if ctx is None:
            raise RuntimeError('Attempted to stream with context but '
                'there was no context in the first place to keep around.')
        with ctx:
            yield None

            try:
                for item in gen:
                    yield item
            finally:
                if hasattr(gen, 'close'):
                    gen.close()

    wrapped_g = generator()
    next(wrapped_g)
    return wrapped_g


def make_response(*args):
    if not args:
        return current_app.response_class()
    if len(args) == 1:
        args = args[0]
    return current_app.make_response(args)


def url_for(endpoint, **values):
    appctx = _app_ctx_stack.top
    reqctx = _request_ctx_stack.top
    if appctx is None:
        raise RuntimeError('Attempted to generate a URL without the '
                           'application context being pushed. This has to be '
                           'executed when application context is available.')

    if reqctx is not None:
        url_adapter = reqctx.url_adapter
        blueprint_name = request.blueprint
        if not reqctx.request._is_old_module:
            if endpoint[:1] == '.':
                if blueprint_name is not None:
                    endpoint = blueprint_name + endpoint
                else:
                    endpoint = endpoint[1:]
        else:
            # TODO: get rid of this deprecated functionality in 1.0
            if '.' not in endpoint:
                if blueprint_name is not None:
                    endpoint = blueprint_name + '.' + endpoint
            elif endpoint.startswith('.'):
                endpoint = endpoint[1:]
        external = values.pop('_external', False)

    else:
        url_adapter = appctx.url_adapter
        if url_adapter is None:
            raise RuntimeError('Application was not able to create a URL '
                               'adapter for request independent URL generation. '
                               'You might be able to fix this by setting '
                               'the SERVER_NAME config variable.')
        external = values.pop('_external', True)

    anchor = values.pop('_anchor', None)
    method = values.pop('_method', None)
    scheme = values.pop('_scheme', None)
    appctx.app.inject_url_defaults(endpoint, values)

    old_scheme = None
    if scheme is not None:
        if not external:
            raise ValueError('When specifying _scheme, _external must be True')
        old_scheme = url_adapter.url_scheme
        url_adapter.url_scheme = scheme

    try:
        try:
            rv = url_adapter.build(endpoint, values, method=method,
                                   force_external=external)
        finally:
            if old_scheme is not None:
                url_adapter.url_scheme = old_scheme
    except BuildError as error:
        values['_external'] = external
        values['_anchor'] = anchor
        values['_method'] = method
        return appctx.app.handle_url_build_error(error, endpoint, values)

    if anchor is not None:
        rv += '#' + url_quote(anchor)
    return rv


def get_template_attribute(template_name, attribute):
    return getattr(current_app.jinja_env.get_template(template_name).module,
                   attribute)


def flash(message, category='message'):
    flashes = session.get('_flashes', [])
    flashes.append((category, message))
    session['_flashes'] = flashes
    message_flashed.send(current_app._get_current_object(),
                         message=message, category=category)


def get_flashed_messages(with_categories=False, category_filter=[]):
    flashes = _request_ctx_stack.top.flashes
    if flashes is None:
        _request_ctx_stack.top.flashes = flashes = session.pop('_flashes') \
            if '_flashes' in session else []
    if category_filter:
        flashes = list(filter(lambda f: f[0] in category_filter, flashes))
    if not with_categories:
        return [x[1] for x in flashes]
    return flashes


def send_file(filename_or_fp, mimetype=None, as_attachment=False,
              attachment_filename=None, add_etags=True,
              cache_timeout=None, conditional=False, last_modified=None):
    mtime = None
    fsize = None
    if isinstance(filename_or_fp, string_types):
        filename = filename_or_fp
        if not os.path.isabs(filename):
            filename = os.path.join(current_app.root_path, filename)
        file = None
        if attachment_filename is None:
            attachment_filename = os.path.basename(filename)
    else:
        file = filename_or_fp
        filename = None

    if mimetype is None:
        if attachment_filename is not None:
            mimetype = mimetypes.guess_type(attachment_filename)[0] \
                or 'application/octet-stream'

        if mimetype is None:
            raise ValueError(
                'Unable to infer MIME-type because no filename is available. '
                'Please set either `attachment_filename`, pass a filepath to '
                '`filename_or_fp` or set your own MIME-type via `mimetype`.'
            )

    headers = Headers()
    if as_attachment:
        if attachment_filename is None:
            raise TypeError('filename unavailable, required for '
                            'sending as attachment')
        headers.add('Content-Disposition', 'attachment',
                    filename=attachment_filename)

    if current_app.use_x_sendfile and filename:
        if file is not None:
            file.close()
        headers['X-Sendfile'] = filename
        fsize = os.path.getsize(filename)
        headers['Content-Length'] = fsize
        data = None
    else:
        if file is None:
            file = open(filename, 'rb')
            mtime = os.path.getmtime(filename)
            fsize = os.path.getsize(filename)
            headers['Content-Length'] = fsize
        data = wrap_file(request.environ, file)

    rv = current_app.response_class(data, mimetype=mimetype, headers=headers,
                                    direct_passthrough=True)

    if last_modified is not None:
        rv.last_modified = last_modified
    elif mtime is not None:
        rv.last_modified = mtime

    rv.cache_control.public = True
    if cache_timeout is None:
        cache_timeout = current_app.get_send_file_max_age(filename)
    if cache_timeout is not None:
        rv.cache_control.max_age = cache_timeout
        rv.expires = int(time() + cache_timeout)

    if add_etags and filename is not None:
        from warnings import warn

        try:
            rv.set_etag('%s-%s-%s' % (
                os.path.getmtime(filename),
                os.path.getsize(filename),
                adler32(
                    filename.encode('utf-8') if isinstance(filename, text_type)
                    else filename
                ) & 0xffffffff
            ))
        except OSError:
            warn('Access %s failed, maybe it does not exist, so ignore etags in '
                 'headers' % filename, stacklevel=2)

    if conditional:
        if callable(getattr(Range, 'to_content_range_header', None)):
            try:
                rv = rv.make_conditional(request, accept_ranges=True,
                                         complete_length=fsize)
            except RequestedRangeNotSatisfiable:
                file.close()
                raise
        else:
            rv = rv.make_conditional(request)
        if rv.status_code == 304:
            rv.headers.pop('x-sendfile', None)
    return rv


def safe_join(directory, *pathnames):
    for filename in pathnames:
        if filename != '':
            filename = posixpath.normpath(filename)
        for sep in _os_alt_seps:
            if sep in filename:
                raise NotFound()
        if os.path.isabs(filename) or \
           filename == '..' or \
           filename.startswith('../'):
            raise NotFound()
        directory = os.path.join(directory, filename)
    return directory


def send_from_directory(directory, filename, **options):
    filename = safe_join(directory, filename)
    if not os.path.isabs(filename):
        filename = os.path.join(current_app.root_path, filename)
    try:
        if not os.path.isfile(filename):
            raise NotFound()
    except (TypeError, ValueError):
        raise BadRequest()
    options.setdefault('conditional', True)
    return send_file(filename, **options)


def get_root_path(import_name):
    mod = sys.modules.get(import_name)
    if mod is not None and hasattr(mod, '__file__'):
        return os.path.dirname(os.path.abspath(mod.__file__))

    loader = pkgutil.get_loader(import_name)

    if loader is None or import_name == '__main__':
        return os.getcwd()

    if hasattr(loader, 'get_filename'):
        filepath = loader.get_filename(import_name)
    else:
        __import__(import_name)
        mod = sys.modules[import_name]
        filepath = getattr(mod, '__file__', None)

        if filepath is None:
            raise RuntimeError('No root path can be found for the provided '
                               'module "%s".  This can happen because the '
                               'module came from an import hook that does '
                               'not provide file name information or because '
                               'it\'s a namespace package.  In this case '
                               'the root path needs to be explicitly '
                               'provided.' % import_name)

    return os.path.dirname(os.path.abspath(filepath))


def _matching_loader_thinks_module_is_package(loader, mod_name):
    if hasattr(loader, 'is_package'):
        return loader.is_package(mod_name)
    elif (loader.__class__.__module__ == '_frozen_importlib' and
          loader.__class__.__name__ == 'NamespaceLoader'):
        return True
    raise AttributeError(
        ('%s.is_package() method is missing but is required by Flask of '
         'PEP 302 import hooks.  If you do not use import hooks and '
         'you encounter this error please file a bug against Flask.') %
        loader.__class__.__name__)


def find_package(import_name):
    root_mod_name = import_name.split('.')[0]
    loader = pkgutil.get_loader(root_mod_name)
    if loader is None or import_name == '__main__':
        package_path = os.getcwd()
    else:
        if hasattr(loader, 'get_filename'):
            filename = loader.get_filename(root_mod_name)
        elif hasattr(loader, 'archive'):
            filename = loader.archive
        else:
            __import__(import_name)
            filename = sys.modules[import_name].__file__
        package_path = os.path.abspath(os.path.dirname(filename))

        if _matching_loader_thinks_module_is_package(
                loader, root_mod_name):
            package_path = os.path.dirname(package_path)

    site_parent, site_folder = os.path.split(package_path)
    py_prefix = os.path.abspath(sys.prefix)
    if package_path.startswith(py_prefix):
        return py_prefix, package_path
    elif site_folder.lower() == 'site-packages':
        parent, folder = os.path.split(site_parent)
        if folder.lower() == 'lib':
            base_dir = parent
        elif os.path.basename(parent).lower() == 'lib':
            base_dir = os.path.dirname(parent)
        else:
            base_dir = site_parent
        return base_dir, package_path
    return None, package_path


class locked_cached_property(object):

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


class _PackageBoundObject(object):

    def __init__(self, import_name, template_folder=None, root_path=None):
        self.import_name = import_name

        self.template_folder = template_folder

        if root_path is None:
            root_path = get_root_path(self.import_name)

        self.root_path = root_path

        self._static_folder = None
        self._static_url_path = None

    def _get_static_folder(self):
        if self._static_folder is not None:
            return os.path.join(self.root_path, self._static_folder)
    def _set_static_folder(self, value):
        self._static_folder = value
    static_folder = property(_get_static_folder, _set_static_folder, doc='''
    The absolute path to the configured static folder.
    ''')
    del _get_static_folder, _set_static_folder

    def _get_static_url_path(self):
        if self._static_url_path is not None:
            return self._static_url_path
        if self.static_folder is not None:
            return '/' + os.path.basename(self.static_folder)
    def _set_static_url_path(self, value):
        self._static_url_path = value
    static_url_path = property(_get_static_url_path, _set_static_url_path)
    del _get_static_url_path, _set_static_url_path

    @property
    def has_static_folder(self):
        return self.static_folder is not None

    @locked_cached_property
    def jinja_loader(self):
        if self.template_folder is not None:
            return FileSystemLoader(os.path.join(self.root_path,
                                                 self.template_folder))

    def get_send_file_max_age(self, filename):
        return total_seconds(current_app.send_file_max_age_default)

    def send_static_file(self, filename):
        if not self.has_static_folder:
            raise RuntimeError('No static folder for this object')
        cache_timeout = self.get_send_file_max_age(filename)
        return send_from_directory(self.static_folder, filename,
                                   cache_timeout=cache_timeout)

    def open_resource(self, resource, mode='rb'):
        if mode not in ('r', 'rb'):
            raise ValueError('Resources can only be opened for reading')
        return open(os.path.join(self.root_path, resource), mode)

def total_seconds(td):
    return td.days * 60 * 60 * 24 + td.seconds

```

## 三. 解析
flask/helpers.py文件主要提供了get_debug_flag/_endpoint_from_view_func/stream_with_context/make_response/url_for/get_template_attribute/flash/get_flashed_messages/send_from_directory/find_package/total_seconds等辅助函数，locked_cached_property/_PackageBoundObject辅助类。
1. get_debug_flag
- 获取环境变量FLASK_DEBUG，如果没有此环境变量，返回None，如果此环境变量的值不在('0', 'false', 'no')之间，返回True，反之返回False

2. _endpoint_from_view_func
- 返回视图函数的endpoint

3. stream_with_context
- stream_with_context获取当前上下文，如果没有上下文，则返回报错RuntimeError，存在上下文，则遍历当前上下文的内容，并对每个元素进行检查，返回一个生成器

4. make_response
- make_response是可接受参数，重新封装自定义数据，返回response对象

5. url_for
- 这部分源码解析可参考[Flask的url_for重定向问题和相应源码分析](https://jiayi.space/post/flaskde-url_forzhong-ding-xiang-wen-ti-he-xiang-ying-yuan-ma-fen-xi)

6. get_template_attribute
- 获取jiaja2模板引擎渲染的模板中的方法

7. flash
- 给session中的_flask的key加上对应的(category, message)，category可以是error/info/warning等，message为字符串

8. get_flashed_messages
- 从当前上下文中获取flashes，如果flashes值为空，则从当前的session中获取_flashes的值或者空列表，然后按照category进行过滤，返回一个列表，列表中的值是message

9. send_from_directory
- 返回一个文件的绝对路径，并为此文件添加了返回头参数等
- 使用可参考[flask.send_from_directory](https://programtalk.com/python-examples/flask.send_from_directory/)

10. find_package
- 获取导入的模块路径，例如import xxx.xxx.xx，按照.切割，获取第一个值，即模块的最上层
- 调用pkgutil.get_loader()获取该模块的的路径，如果获取失败，则返回当前主程序的路径，获取成功，返回此模块的绝对路径
- 对此模块的绝对路径进行切割，返回路径前缀和绝对路径的值

11. total_seconds
- 获取一个时间对象（timedelta），返回当前时间对象的秒数

12. locked_cached_property
- 可以根据前文来对照一下cached_property和locked_cached_property的异同
- 当访问locked_cached_property的属性时，就会调用__get__方法，如果当前参数obj为None，则返回自身self，反之，打开线程锁，并在 obj.__dict__中寻找是否已经存在对应的值，存在对应的值则直接返回，不存在调用底层的函数 self.func，并把得到的值保存起来，再返回

13. _PackageBoundObject
- _PackageBoundObject类主要实现了静态文件相关的属性，比如获取文件路径/加载jinja2引擎模板/获取静态文件的最大缓存时间等等