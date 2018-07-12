---
layout: blog
title: 'Flask源码剖析（六）：响应'
date: 2017-06-22 18:46:29
categories: flask
tags: flask
lead_text: '响应'
---

## 一. 含义
与Request类相对应，Flask框架使用Response类对HTTP请求的响应。
平时我们不会跟response直接打交道，Flask会自动生成响应返回到客户端，只需要关注view的实现即可。

## 二. 响应
先来看看一个正常的response是什么格式，用简单应用实现的程序为例。
```bash
curl -i "http://127.0.0.1:5000/"
```
返回：
```bash
HTTP/1.0 200 OK
Content-Type: text/html; charset=utf-8
Content-Length: 12
Server: Werkzeug/0.10.4 Python/2.7.6
Date: Fri, 28 Jul 2017 08:12:55 GMT

Hello World!
```

HTTP响应分为三个部分：
1. 响应行（HTTP版本、状态码和原因短语）
2. 响应头（键/值对的列表，响应的时间、此次响应返回的文件长度、文件类型等）
3. 响应体（服务器返回给客户端的文件、数据等）

## 三. 实现
还是从flask.app.wsgi_app()这个函数入手
```python
class Flask(_PackageBoundObject):

    response_class = Response
    
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

来看看full_dispatch_request()是怎么处理response的
```python
    def full_dispatch_request(self):
        self.try_trigger_before_first_request_functions()
        try:
            request_started.send(self)
            rv = self.preprocess_request()
            if rv is None:
                rv = self.dispatch_request()
        except Exception as e:
            rv = self.handle_user_exception(e)
        return self.finalize_request(rv)
```
经过路由匹配，视图处理相关的逻辑后，full_dispatch_request()最终是调用self.finalize_request(rv)
```python
    def finalize_request(self, rv, from_error_handler=False):
        response = self.make_response(rv)
        try:
            response = self.process_response(response)
            request_finished.send(self, response=response)
        except Exception:
            if not from_error_handler:
                raise
            self.logger.exception('Request finalizing failed with an '
                                  'error while handling an error')
        return response
```
这个函数主要是调用make_response()更改http响应的相关信息，包括响应行，响应头，响应体，来生成一个response实例对象。
process_response()处理response对象，执行对应路由的函数来处理response对象信息，然后发结束请求的信号。

```python
    def make_response(self, rv):
        status_or_headers = headers = None
        if isinstance(rv, tuple):
            rv, status_or_headers, headers = rv + (None,) * (3 - len(rv))

        if rv is None:
            raise ValueError('View function did not return a response')

        if isinstance(status_or_headers, (dict, list)):
            headers, status_or_headers = status_or_headers, None

        if not isinstance(rv, self.response_class):
            if isinstance(rv, (text_type, bytes, bytearray)):
                rv = self.response_class(rv, headers=headers,
                                         status=status_or_headers)
                headers = status_or_headers = None
            else:
                rv = self.response_class.force_type(rv, request.environ)

        if status_or_headers is not None:
            if isinstance(status_or_headers, string_types):
                rv.status = status_or_headers
            else:
                rv.status_code = status_or_headers
        if headers:
            rv.headers.extend(headers)

        return rv
```

1. 如果rv是tuple，尝试用(response, status, headers)或者(response, headers)去解析
2. 如果rv是字符串，自动设置响应行和响应头，rv为响应体
3. 如果rv是response_class实例，直接使用

接着来看一下response对象是怎么实例化的
response_class是flask.wrappers:Response类的对象，封装了wergzeug.wrappers:Response类，在这基础上添加了新的属性。
来看看flask.wrappers:Response类是怎么实现的
```python
class Response(ResponseBase):
    default_mimetype = 'text/html'
```
flask.wrappers:Response类封装了wergzeug.wrappers:Response类，设置默认类型，其他的操作直接调用wergzeug.wrappers:Response类。
```python
class Response(BaseResponse, ETagResponseMixin, ResponseStreamMixin,
               CommonResponseDescriptorsMixin,
               WWWAuthenticateMixin):
    """Full featured response object implementing the following mixins:

    - :class:`ETagResponseMixin` for etag and cache control handling
    - :class:`ResponseStreamMixin` to add support for the `stream` property
    - :class:`CommonResponseDescriptorsMixin` for various HTTP descriptors
    - :class:`WWWAuthenticateMixin` for HTTP authentication support
    """
```
与request对象一样，wergzeug.wrappers:Response类继承了多个类，BaseResponse, ETagResponseMixin, ResponseStreamMixin, CommonResponseDescriptorsMixin， WWWAuthenticateMixin。
这里主要看BaseResponse类的实现
````python
class BaseResponse(object):

    charset = 'utf-8'

    default_status = 200

    default_mimetype = 'text/plain'

    implicit_sequence_conversion = True

    autocorrect_location_header = True

    automatically_set_content_length = True

    def __init__(self, response=None, status=None, headers=None,
                 mimetype=None, content_type=None, direct_passthrough=False):
        if isinstance(headers, Headers):
            self.headers = headers
        elif not headers:
            self.headers = Headers()
        else:
            self.headers = Headers(headers)

        if content_type is None:
            if mimetype is None and 'content-type' not in self.headers:
                mimetype = self.default_mimetype
            if mimetype is not None:
                mimetype = get_content_type(mimetype, self.charset)
            content_type = mimetype
        if content_type is not None:
            self.headers['Content-Type'] = content_type
        if status is None:
            status = self.default_status
        if isinstance(status, integer_types):
            self.status_code = status
        else:
            self.status = status

        self.direct_passthrough = direct_passthrough
        self._on_close = []

        if response is None:
            self.response = []
        elif isinstance(response, (text_type, bytes, bytearray)):
            self.set_data(response)
        else:
            self.response = response

    def call_on_close(self, func):
        self._on_close.append(func)
        return func

    def __repr__(self):
        if self.is_sequence:
            body_info = '%d bytes' % sum(map(len, self.iter_encoded()))
        else:
            body_info = self.is_streamed and 'streamed' or 'likely-streamed'
        return '<%s %s [%s]>' % (
            self.__class__.__name__,
            body_info,
            self.status
        )

    @classmethod
    def force_type(cls, response, environ=None):
        if not isinstance(response, BaseResponse):
            if environ is None:
                raise TypeError('cannot convert WSGI application into '
                                'response objects without an environ')
            response = BaseResponse(*_run_wsgi_app(response, environ))
        response.__class__ = cls
        return response

    @classmethod
    def from_app(cls, app, environ, buffered=False):
        return cls(*_run_wsgi_app(app, environ, buffered))

    def _get_status_code(self):
        return self._status_code
    def _set_status_code(self, code):
        self._status_code = code
        try:
            self._status = '%d %s' % (code, HTTP_STATUS_CODES[code].upper())
        except KeyError:
            self._status = '%d UNKNOWN' % code
    status_code = property(_get_status_code, _set_status_code,
                           doc='The HTTP Status code as number')
    del _get_status_code, _set_status_code

    def _get_status(self):
        return self._status
    def _set_status(self, value):
        self._status = to_native(value)
        try:
            self._status_code = int(self._status.split(None, 1)[0])
        except ValueError:
            self._status_code = 0
            self._status = '0 %s' % self._status
    status = property(_get_status, _set_status, doc='The HTTP Status code')
    del _get_status, _set_status

    def get_data(self, as_text=False):
        self._ensure_sequence()
        rv = b''.join(self.iter_encoded())
        if as_text:
            rv = rv.decode(self.charset)
        return rv

    def set_data(self, value):
        if isinstance(value, text_type):
            value = value.encode(self.charset)
        else:
            value = bytes(value)
        self.response = [value]
        if self.automatically_set_content_length:
            self.headers['Content-Length'] = str(len(value))

    data = property(get_data, set_data, doc='''
        A descriptor that calls :meth:`get_data` and :meth:`set_data`.  This
        should not be used and will eventually get deprecated.
        ''')

    def calculate_content_length(self):
        try:
            self._ensure_sequence()
        except RuntimeError:
            return None
        return sum(len(x) for x in self.response)

    def _ensure_sequence(self, mutable=False):
        if self.is_sequence:
            if mutable and not isinstance(self.response, list):
                self.response = list(self.response)
            return
        if self.direct_passthrough:
            raise RuntimeError('Attempted implicit sequence conversion '
                               'but the response object is in direct '
                               'passthrough mode.')
        if not self.implicit_sequence_conversion:
            raise RuntimeError('The response object required the iterable '
                               'to be a sequence, but the implicit '
                               'conversion was disabled.  Call '
                               'make_sequence() yourself.')
        self.make_sequence()

    def make_sequence(self):
        if not self.is_sequence:
            close = getattr(self.response, 'close', None)
            self.response = list(self.iter_encoded())
            if close is not None:
                self.call_on_close(close)

    def iter_encoded(self):
        if __debug__:
            _warn_if_string(self.response)
        return _iter_encoded(self.response, self.charset)

    def set_cookie(self, key, value='', max_age=None, expires=None,
                   path='/', domain=None, secure=None, httponly=False):
        self.headers.add('Set-Cookie', dump_cookie(key, value, max_age,
                         expires, path, domain, secure, httponly,
                         self.charset))

    def delete_cookie(self, key, path='/', domain=None):
        self.set_cookie(key, expires=0, max_age=0, path=path, domain=domain)

    @property
    def is_streamed(self):
        try:
            len(self.response)
        except (TypeError, AttributeError):
            return True
        return False

    @property
    def is_sequence(self):
        return isinstance(self.response, (tuple, list))

    def close(self):
        if hasattr(self.response, 'close'):
            self.response.close()
        for func in self._on_close:
            func()

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_value, tb):
        self.close()

    def freeze(self):
        self.response = list(self.iter_encoded())
        self.headers['Content-Length'] = str(sum(map(len, self.response)))

    def get_wsgi_headers(self, environ):
        headers = Headers(self.headers)
        location = None
        content_location = None
        content_length = None
        status = self.status_code

        for key, value in headers:
            ikey = key.lower()
            if ikey == u'location':
                location = value
            elif ikey == u'content-location':
                content_location = value
            elif ikey == u'content-length':
                content_length = value

        if location is not None:
            old_location = location
            if isinstance(location, text_type):
                location = iri_to_uri(location, safe_conversion=True)

            if self.autocorrect_location_header:
                current_url = get_current_url(environ, root_only=True)
                if isinstance(current_url, text_type):
                    current_url = iri_to_uri(current_url)
                location = url_join(current_url, location)
            if location != old_location:
                headers['Location'] = location

        if content_location is not None and \
           isinstance(content_location, text_type):
            headers['Content-Location'] = iri_to_uri(content_location)

        if 100 <= status < 200 or status == 204:
            headers['Content-Length'] = content_length = u'0'
        elif status == 304:
            remove_entity_headers(headers)

        if self.automatically_set_content_length and \
           self.is_sequence and content_length is None and status != 304:
            try:
                content_length = sum(len(to_bytes(x, 'ascii'))
                                     for x in self.response)
            except UnicodeError:
                pass
            else:
                headers['Content-Length'] = str(content_length)

        return headers

    def get_app_iter(self, environ):
        status = self.status_code
        if environ['REQUEST_METHOD'] == 'HEAD' or \
           100 <= status < 200 or status in (204, 304):
            iterable = ()
        elif self.direct_passthrough:
            if __debug__:
                _warn_if_string(self.response)
            return self.response
        else:
            iterable = self.iter_encoded()
        return ClosingIterator(iterable, self.close)

    def get_wsgi_response(self, environ):
        headers = self.get_wsgi_headers(environ)
        app_iter = self.get_app_iter(environ)
        return app_iter, self.status, headers.to_wsgi_list()

    def __call__(self, environ, start_response):
        app_iter, status, headers = self.get_wsgi_response(environ)
        start_response(status, headers)
        return app_iter
````
从这个类，我们可以知道几点
1. BaseResponse类是一个上下文管理器
2. BaseResponse类定义了一些默认值
3. BaseResponse类接受可选参数response(响应体),status(响应行的状态码),headers(响应头),mimetype(响应体的头部),content_type(响应头的Content-Type)
4. BaseResponse类提供向外直接读写的接口data = property(get_data, set_data, doc='''...''')
5. BaseResponse类的status和status_code可直接修改
6. BaseResponse类的响应体的编码(Content-Type)和长度(Content-Length)是默认的，不需要手动设置
7. BaseResponse类的响应头的存储是放在werkzeug.datastructures:Headers类实现的

## 四. 抽象
```python
flask.app:Flask.full_dispatch_request()-->flask.app:Flask.finalize_request()-->flask.app:Flask.make_response()-->
flask.app:Flask.process_response()-->flask.signals.request_finished()
```