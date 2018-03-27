---
layout: blog
title: 'Flask源码剖析（五）：请求'
date: 2017-06-15 21:37:41
categories: flask
tags: flask
lead_text: '请求'
---

## 一. 含义
先来看一下request的定义：
```python
def _lookup_req_object(name):
    top = _request_ctx_stack.top
    if top is None:
        raise RuntimeError(_request_ctx_err_msg)
    return getattr(top, name)
...

request = LocalProxy(partial(_lookup_req_object, 'request'))
```
实质上以上函数可以看作
```python
request = LocalProxy(_request_ctx_stack.top.request) = LocalProxy (_request_ctx_stack._local[stack][-1].request)
```
栈顶元素内，name叫做request对象的值，这个值包含了HTTP请求头等信息，包括像HTTP请求头的信息，可以提供给全局使用。

## 二. 请求
request对象在RequestContext实例初始化的时候，就已经创建了，实例被压入上下文栈顶之后，
只是通过LocalProxy形成了新的代理后的request，但内容是初始化时创建的。
```python
class RequestContext(object):

    def __init__(self, app, environ, request=None):
        self.app = app
        if request is None:
            request = app.request_class(environ)
        self.request = request
        self.url_adapter = app.create_url_adapter(self.request)
        self.flashes = None
        self.session = None
    ...
```
RequestContext类的实例有request=app.request_class属性。

## 三. 实现
request_class是flask.wrappers:Request类的对象，封装了wergzeug.wrappers:Request类，在这基础上添加了新的属性。
来看看flask.wrappers:Request类是怎么实现的
```python
# -*- coding: utf-8 -*-

from werkzeug.wrappers import Request as RequestBase, Response as ResponseBase
from werkzeug.exceptions import BadRequest

from . import json
from .globals import _request_ctx_stack

_missing = object()


def _get_data(req, cache):
    getter = getattr(req, 'get_data', None)
    if getter is not None:
        return getter(cache=cache)
    return req.data


class Request(RequestBase):
    url_rule = None

    view_args = None

    routing_exception = None

    _is_old_module = False

    @property
    def max_content_length(self):
        ctx = _request_ctx_stack.top
        if ctx is not None:
            return ctx.app.config['MAX_CONTENT_LENGTH']

    @property
    def endpoint(self):
        if self.url_rule is not None:
            return self.url_rule.endpoint

    @property
    def module(self):
        from warnings import warn
        warn(DeprecationWarning('modules were deprecated in favor of '
                                'blueprints.  Use request.blueprint '
                                'instead.'), stacklevel=2)
        if self._is_old_module:
            return self.blueprint

    @property
    def blueprint(self):
        if self.url_rule and '.' in self.url_rule.endpoint:
            return self.url_rule.endpoint.rsplit('.', 1)[0]

    @property
    def json(self):
        from warnings import warn
        warn(DeprecationWarning('json is deprecated.  '
                                'Use get_json() instead.'), stacklevel=2)
        return self.get_json()

    @property
    def is_json(self):
        mt = self.mimetype
        if mt == 'application/json':
            return True
        if mt.startswith('application/') and mt.endswith('+json'):
            return True
        return False

    def get_json(self, force=False, silent=False, cache=True):
        rv = getattr(self, '_cached_json', _missing)
        
        if cache and rv is not _missing:
            return rv

        if not (force or self.is_json):
            return None

        request_charset = self.mimetype_params.get('charset')
        try:
            data = _get_data(self, cache)
            if request_charset is not None:
                rv = json.loads(data, encoding=request_charset)
            else:
                rv = json.loads(data)
        except ValueError as e:
            if silent:
                rv = None
            else:
                rv = self.on_json_loading_failed(e)
        if cache:
            self._cached_json = rv
        return rv

    def on_json_loading_failed(self, e):
        ctx = _request_ctx_stack.top
        if ctx is not None and ctx.app.config.get('DEBUG', False):
            raise BadRequest('Failed to decode JSON object: {0}'.format(e))
        raise BadRequest()

    def _load_form_data(self):
        RequestBase._load_form_data(self)
        
        ctx = _request_ctx_stack.top
        if ctx is not None and ctx.app.debug and \
           self.mimetype != 'multipart/form-data' and not self.files:
            from .debughelpers import attach_enctype_error_multidict
            attach_enctype_error_multidict(self)
```
这个类很简单，即是封装了wergzeug.wrappers:Request类，增加了endpoint,json,is_json等属性，和on_json_loading_failed，_load_form_data等方法。

接着，看看wergzeug.wrappers:Request类
```python
class Request(BaseRequest, AcceptMixin, ETagRequestMixin,
              UserAgentMixin, AuthorizationMixin,
              CommonRequestDescriptorsMixin):
    """Full featured request object implementing the following mixins:

    - :class:`AcceptMixin` for accept header parsing
    - :class:`ETagRequestMixin` for etag and cache control handling
    - :class:`UserAgentMixin` for user agent introspection
    - :class:`AuthorizationMixin` for http auth handling
    - :class:`CommonRequestDescriptorsMixin` for common headers
    """
```
wergzeug.wrappers:Request类继承了多个类，BaseRequest，AcceptMixin，ETagRequestMixin，UserAgentMixin，AuthorizationMixin，CommonRequestDescriptorsMixin。
这里主要看BaseRequest类的实现
```python
class BaseRequest(object):
    charset = 'utf-8'

    encoding_errors = 'replace'

    max_content_length = None

    max_form_memory_size = None

    parameter_storage_class = ImmutableMultiDict

    list_storage_class = ImmutableList

    dict_storage_class = ImmutableTypeConversionDict

    form_data_parser_class = FormDataParser

    trusted_hosts = None

    disable_data_descriptor = False

    def __init__(self, environ, populate_request=True, shallow=False):
        self.environ = environ
        if populate_request and not shallow:
            self.environ['werkzeug.request'] = self
        self.shallow = shallow

    def __repr__(self):
        args = []
        try:
            args.append("'%s'" % to_native(self.url, self.url_charset))
            args.append('[%s]' % self.method)
        except Exception:
            args.append('(invalid WSGI environ)')

        return '<%s %s>' % (
            self.__class__.__name__,
            ' '.join(args)
        )

    @property
    def url_charset(self):
        return self.charset

    @classmethod
    def from_values(cls, *args, **kwargs):
        from werkzeug.test import EnvironBuilder
        charset = kwargs.pop('charset', cls.charset)
        kwargs['charset'] = charset
        builder = EnvironBuilder(*args, **kwargs)
        try:
            return builder.get_request(cls)
        finally:
            builder.close()

    @classmethod
    def application(cls, f):
        def application(*args):
            request = cls(args[-2])
            with request:
                return f(*args[:-2] + (request,))(*args[-2:])
        return update_wrapper(application, f)

    def _get_file_stream(self, total_content_length, content_type, filename=None,
                        content_length=None):
        return default_stream_factory(total_content_length, content_type,
                                      filename, content_length)

    @property
    def want_form_data_parsed(self):
        return bool(self.environ.get('CONTENT_TYPE'))

    def make_form_data_parser(self):
        return self.form_data_parser_class(self._get_file_stream,
                                           self.charset,
                                           self.encoding_errors,
                                           self.max_form_memory_size,
                                           self.max_content_length,
                                           self.parameter_storage_class)

    def _load_form_data(self):
        if 'form' in self.__dict__:
            return

        _assert_not_shallow(self)

        if self.want_form_data_parsed:
            content_type = self.environ.get('CONTENT_TYPE', '')
            content_length = get_content_length(self.environ)
            mimetype, options = parse_options_header(content_type)
            parser = self.make_form_data_parser()
            data = parser.parse(self._get_stream_for_parsing(),
                                mimetype, content_length, options)
        else:
            data = (self.stream, self.parameter_storage_class(),
                    self.parameter_storage_class())

        d = self.__dict__
        d['stream'], d['form'], d['files'] = data

    def _get_stream_for_parsing(self):
        cached_data = getattr(self, '_cached_data', None)
        if cached_data is not None:
            return BytesIO(cached_data)
        return self.stream

    def close(self):
        files = self.__dict__.get('files')
        for key, value in iter_multi_items(files or ()):
            value.close()

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_value, tb):
        self.close()

    @cached_property
    def stream(self):
        _assert_not_shallow(self)
        return get_input_stream(self.environ)

    input_stream = environ_property('wsgi.input', 'The WSGI input stream.\n'
        'In general it\'s a bad idea to use this one because you can easily '
        'read past the boundary.  Use the :attr:`stream` instead.')

    @cached_property
    def args(self):
        return url_decode(wsgi_get_bytes(self.environ.get('QUERY_STRING', '')),
                          self.url_charset, errors=self.encoding_errors,
                          cls=self.parameter_storage_class)

    @cached_property
    def data(self):
        if self.disable_data_descriptor:
            raise AttributeError('data descriptor is disabled')
        return self.get_data(parse_form_data=True)

    def get_data(self, cache=True, as_text=False, parse_form_data=False):
        rv = getattr(self, '_cached_data', None)
        if rv is None:
            if parse_form_data:
                self._load_form_data()
            rv = self.stream.read()
            if cache:
                self._cached_data = rv
        if as_text:
            rv = rv.decode(self.charset, self.encoding_errors)
        return rv

    @cached_property
    def form(self):
        self._load_form_data()
        return self.form

    @cached_property
    def values(self):
        args = []
        for d in self.args, self.form:
            if not isinstance(d, MultiDict):
                d = MultiDict(d)
            args.append(d)
        return CombinedMultiDict(args)

    @cached_property
    def files(self):
        self._load_form_data()
        return self.files

    @cached_property
    def cookies(self):
        return parse_cookie(self.environ, self.charset,
                            self.encoding_errors,
                            cls=self.dict_storage_class)

    @cached_property
    def headers(self):
        return EnvironHeaders(self.environ)

    @cached_property
    def path(self):
        raw_path = wsgi_decoding_dance(self.environ.get('PATH_INFO') or '',
                                       self.charset, self.encoding_errors)
        return '/' + raw_path.lstrip('/')

    @cached_property
    def full_path(self):
        return self.path + u'?' + to_unicode(self.query_string, self.url_charset)

    @cached_property
    def script_root(self):
        raw_path = wsgi_decoding_dance(self.environ.get('SCRIPT_NAME') or '',
                                       self.charset, self.encoding_errors)
        return raw_path.rstrip('/')

    @cached_property
    def url(self):
        return get_current_url(self.environ,
                               trusted_hosts=self.trusted_hosts)

    @cached_property
    def base_url(self):
        return get_current_url(self.environ, strip_querystring=True,
                               trusted_hosts=self.trusted_hosts)

    @cached_property
    def url_root(self):
        return get_current_url(self.environ, True,
                               trusted_hosts=self.trusted_hosts)

    @cached_property
    def host_url(self):
        return get_current_url(self.environ, host_only=True,
                               trusted_hosts=self.trusted_hosts)

    @cached_property
    def host(self):
        return get_host(self.environ, trusted_hosts=self.trusted_hosts)

    query_string = environ_property('QUERY_STRING', '', read_only=True,
        load_func=wsgi_get_bytes, doc=
        '''The URL parameters as raw bytestring.''')
    method = environ_property('REQUEST_METHOD', 'GET', read_only=True,
        load_func=lambda x: x.upper(), doc=
        '''The transmission method. (For example ``'GET'`` or ``'POST'``).''')

    @cached_property
    def access_route(self):
        if 'HTTP_X_FORWARDED_FOR' in self.environ:
            addr = self.environ['HTTP_X_FORWARDED_FOR'].split(',')
            return self.list_storage_class([x.strip() for x in addr])
        elif 'REMOTE_ADDR' in self.environ:
            return self.list_storage_class([self.environ['REMOTE_ADDR']])
        return self.list_storage_class()

    @property
    def remote_addr(self):
        return self.environ.get('REMOTE_ADDR')

    remote_user = environ_property('REMOTE_USER', doc='''
        If the server supports user authentication, and the script is
        protected, this attribute contains the username the user has
        authenticated as.''')

    scheme = environ_property('wsgi.url_scheme', doc='''
        URL scheme (http or https).

        .. versionadded:: 0.7''')

    is_xhr = property(lambda x: x.environ.get('HTTP_X_REQUESTED_WITH', '')
                      .lower() == 'xmlhttprequest', doc='''
        True if the request was triggered via a JavaScript XMLHttpRequest.
        This only works with libraries that support the `X-Requested-With`
        header and set it to "XMLHttpRequest".  Libraries that do that are
        prototype, jQuery and Mochikit and probably some more.''')
    is_secure = property(lambda x: x.environ['wsgi.url_scheme'] == 'https',
                         doc='`True` if the request is secure.')
    is_multithread = environ_property('wsgi.multithread', doc='''
        boolean that is `True` if the application is served by
        a multithreaded WSGI server.''')
    is_multiprocess = environ_property('wsgi.multiprocess', doc='''
        boolean that is `True` if the application is served by
        a WSGI server that spawns multiple processes.''')
    is_run_once = environ_property('wsgi.run_once', doc='''
        boolean that is `True` if the application will be executed only
        once in a process lifetime.  This is the case for CGI for example,
        but it's not guaranteed that the exeuction only happens one time.''')
```
这个类的内容很多，但也只是很简单的数据转换，实例化需要的唯一变量是environ，获取environ的值，实现了很多属性和方法来封装environ的各个值，主要用了cached_property装饰器。

## 四. 抽象
```python
flask.ctx:RequestContext-->flask.wrappers:Request-->wergzeug.wrappers:Request-->
wergzeug.wrappers:(BaseRequest, AcceptMixin, ETagRequestMixin,UserAgentMixin, AuthorizationMixin,CommonRequestDescriptorsMixin)
```
