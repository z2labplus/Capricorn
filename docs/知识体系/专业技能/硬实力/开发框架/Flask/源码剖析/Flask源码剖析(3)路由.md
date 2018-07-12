---
layout: blog
title: 'Flask源码剖析（三）：路由'
date: 2017-06-03 20:36:54
categories: flask
tags: flask
lead_text: '路由'
---

## 一. 路由含义
在Flask框架中，路由是指用户请求的URL与视图函数之间的映射,根据HTTP请求的URL在路由表中匹配预定义的URL规则，找到对应的视图函数， 并将视图函数的执行结果返回WSGI服务器。

## 二. 路由规则
### flask.Flask.route()装饰器
```python
@app.route('/')
def hello_world():
    return 'Hello World!'
```
### flask.Flask.add_url_rule()函数
```python
def hello():
    return "hello, world!"

app.add_url_rule('/', 'hello', hello)
```
### 访问基于werkzeug路由系统的flask.Flask.url_map
```python
url_map = Map([
    Rule('/', endpoint='hello')
])

```
## 三. 路由逻辑
注册路由时，Flask做了什么？先从最常使用的route()装饰器开始。
```python
def route(self, rule, **options):
    """Like :meth:`Flask.route` but for a blueprint.  The endpoint for the
    :func:`url_for` function is prefixed with the name of the blueprint.
    """
    def decorator(f):
        endpoint = options.pop("endpoint", f.__name__)
        self.add_url_rule(rule, endpoint, f, **options)
        return f
    return decorator
```
可知，route()装饰器是对add_url_rule()函数的封装。第一种方法等价于第二种方式，但第一种方法实现的更加优雅。
下面，继续深究add_url_rule()函数实现了什么功能。
```python
@setupmethod
def add_url_rule(self, rule, endpoint=None, view_func=None, **options):
  
    if endpoint is None:
        endpoint = _endpoint_from_view_func(view_func)
    options['endpoint'] = endpoint
    methods = options.pop('methods', None)

    if methods is None:
        methods = getattr(view_func, 'methods', None) or ('GET',)
    methods = set(methods)

    required_methods = set(getattr(view_func, 'required_methods', ()))
    
    provide_automatic_options = getattr(view_func,
        'provide_automatic_options', None)

    if provide_automatic_options is None:
        if 'OPTIONS' not in methods:
            provide_automatic_options = True
            required_methods.add('OPTIONS')
        else:
            provide_automatic_options = False

    methods |= required_methods

    options['defaults'] = options.get('defaults') or None

    rule = self.url_rule_class(rule, methods=methods, **options)
    rule.provide_automatic_options = provide_automatic_options

    self.url_map.add(rule)
    if view_func is not None:
        old_func = self.view_functions.get(endpoint)
        if old_func is not None and old_func != view_func:
            raise AssertionError('View function mapping is overwriting an '
                                 'existing endpoint function: %s' % endpoint)
        self.view_functions[endpoint] = view_func
```
这个函数大致实现了几种判断：
1. 当endpoint为None时，endpoint为默认的视图函数的名字，将endpoint添加到options字典中
2. 当methods为None时,methods默认为('GET',)
add_url_rule()执行完毕，填充了self.url_map和self.view_functions,具体的实现逻辑看werkzeug的路由系统。
接着，简单看看werkzeug是怎么实现路由功能的,先不从代码层面上分析，只看实现。
```python
    In [1]: from werkzeug.routing import Map, Rule
    
    In [2]: url_map = Map([
       ...: Rule('/', endpoint='hello'),
       ...: Rule('/index/', endpoint='index'),
       ...: Rule('/index/<int:id>', endpoint='index/show')
       ...: ])
    
    In [3]: url = url_map.bind('test.com', '/')
    
    In [4]: url.match("/", "GET")
    Out[4]: ('hello', {})
    
    In [5]: url.match('/index/123', 'GET')
    Out[5]: ('index/show', {'id': 123})
    
    In [6]: url.match('/index')
    ---------------------------------------------------------------------------
    RequestRedirect                           Traceback (most recent call last)
    <ipython-input-6-cddc343ab7ef> in <module>()
    ----> 1 url.match('/index')
    
    /usr/local/lib/python2.7/dist-packages/werkzeug/routing.pyc in match(self, path_info, method, return_rule, query_args)
       1442                 raise RequestRedirect(self.make_redirect_url(
       1443                     url_quote(path_info, self.map.charset,
    -> 1444                               safe='/:|+') + '/', query_args))
       1445             except RequestAliasRedirect as e:
       1446                 raise RequestRedirect(self.make_alias_redirect_url(
    
    RequestRedirect: 301: Moved Permanently
    
    In [7]: url.match('/404')
    ---------------------------------------------------------------------------
    NotFound                                  Traceback (most recent call last)
    <ipython-input-7-a59a1491d8b0> in <module>()
    ----> 1 url.match('/404')
    
    /usr/local/lib/python2.7/dist-packages/werkzeug/routing.pyc in match(self, path_info, method, return_rule, query_args)
       1481         if have_match_for:
       1482             raise MethodNotAllowed(valid_methods=list(have_match_for))
    -> 1483         raise NotFound()
       1484 
       1485     def test(self, path_info=None, method=None):
    
    NotFound: 404: Not Found
```
上面这段代码主要演示了werkzeug的核心路由功能：
1. 添加路由规则
2. 绑定路由表
3. 匹配url
4. 正常情况下返回endpoint和参数字典，或异常情况下返回重定向，404状态

可知，werkzeug的路由过程是url到endpoint的转换，对视图函数和endpoint之间的对应关系无感知的，只是根据url来查找对应的endpoint。
## 四. 路由实现
在简单应用这一篇中，已经将当Flask运行时，一个request请求进来，会进行怎样的数据处理的过程抽象了出来。
回到路由匹配的逻辑：
```python
    def dispatch_request(self):
        req = _request_ctx_stack.top.request
        if req.routing_exception is not None:
            self.raise_routing_exception(req)
        rule = req.url_rule
        if getattr(rule, 'provide_automatic_options', False) \
           and req.method == 'OPTIONS':
            return self.make_default_options_response()
        return self.view_functions[rule.endpoint](**req.view_args)
```
这个函数的逻辑很简单：
1. 找到请求对象request，获取endpoint
2. view_functions找到对应endpoint的view_func，把请求参数传递过去，进行处理并返回

要理解这个函数,有几点需要去探究一下的。
_request_ctx_stack是定义在global.py里面的堆栈，具体实现是werkzeug/local.py文件中，顾名思义,堆栈是一个后进先出的集合，先不深究。
request是一个RequestContext对象，具体实现在flask/ctx.py文件中,是一个上下文管理器，和路由相关的逻辑如下：
```python
    class Flask(_PackageBoundObject):
        def create_url_adapter(self, request):
            if request is not None:
                return self.url_map.bind_to_environ(request.environ,
                    server_name=self.config['SERVER_NAME'])
            if self.config['SERVER_NAME'] is not None:
                return self.url_map.bind(
                    self.config['SERVER_NAME'],
                    script_name=self.config['APPLICATION_ROOT'] or '/',
                    url_scheme=self.config['PREFERRED_URL_SCHEME'])

```

```python
    class RequestContext(object):
        def __init__(self, app, environ, request=None):
            self.app = app
            if request is None:
                request = app.request_class(environ)
            self.request = request
            self.url_adapter = app.create_url_adapter(self.request)
    
            self.match_request()
            
        def match_request(self):
            try:
                url_rule, self.request.view_args = \
                    self.url_adapter.match(return_rule=True)
                self.request.url_rule = url_rule
            except HTTPException as e:
                self.request.routing_exception = e
                
        def push(self):
            top = _request_ctx_stack.top
            if top is not None and top.preserved:
                top.pop(top._preserved_exc)
    
            app_ctx = _app_ctx_stack.top
            if app_ctx is None or app_ctx.app != self.app:
                app_ctx = self.app.app_context()
                app_ctx.push()
                self._implicit_app_ctx_stack.append(app_ctx)
            else:
                self._implicit_app_ctx_stack.append(None)
    
            if hasattr(sys, 'exc_clear'):
                sys.exc_clear()
    
            _request_ctx_stack.push(self)
    
            self.session = self.app.open_session(self.request)
            if self.session is None:
                self.session = self.app.make_null_session()
                
        def __enter__(self):
            self.push()
            return self
    
        def __exit__(self, exc_type, exc_value, tb):
            self.auto_pop(exc_value)
    
            if BROKEN_PYPY_CTXMGR_EXIT and exc_type is not None:
                reraise(exc_type, exc_value, tb)
```

```python
    def match(self, path):
        if not self.build_only:
            m = self._regex.search(path)
            if m is not None:
                groups = m.groupdict()
                if self.strict_slashes and not self.is_leaf and \
                   not groups.pop('__suffix__'):
                    raise RequestSlash()
                elif not self.strict_slashes:
                    del groups['__suffix__']
    
                result = {}
                for name, value in iteritems(groups):
                    try:
                        value = self._converters[name].to_python(value)
                    except ValidationError:
                        return
                    result[str(name)] = value
                if self.defaults:
                    result.update(self.defaults)
    
                if self.alias and self.map.redirect_defaults:
                    raise RequestAliasRedirect(result)
                    
                return result
```
从以上的路由实现，不难得出以下调用链：
1. RequestContext对象初始化时，调用app.create_url_adapter()方法，将url_map绑定到WSGI的环境变量
2. 调用match_request()方法，match_request调用了werkzeug的match方法，进行匹配url工作，返回rule(详情看路由逻辑的第三种方式)
3. 调用dispatch_request()方法，获取匹配的endpoint值(rule.endpoint)，找到对应的view_func，把请求参数传递过去，进行处理并返回

## 五. 抽象
当一个request请求进来时，Flask抽象后的函数调用链应如下：
```bash
wsgi_app()[flask/app.py]-->request_context()[flask/app.py]-->RequestContext.__init__()[flask/ctx.py]-->
create_url_adapter()[flask/app.py]-->Map.bind()[werkzeug/routing.py]-->match_request()[flask/ctx.py]-->
Map.match()[werkzeug/routing.py]-->dispatch_request()[flask/app.py]
```