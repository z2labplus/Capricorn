---
layout: blog
title: 'Flask源码剖析（二）：简单应用'
date: 2017-05-29 22:01:17
categories: flask
tags: flask
lead_text: '简单应用'
---

## 一. 最小的应用
```
# -*- coding: utf-8 -*-
from flask import Flask
app = Flask(__name__)

@app.route('/')
def hello_world():
    import ipdb
    ipdb.set_trace()
    return 'Hello World!'

if __name__ == '__main__':
    app.run()
```
## 二. 分析工具
### ipdb

> [ipdb](https://github.com/gotcha/ipdb) exports functions to access the IPython debugger, 
which features tab completion, syntax highlighting, better tracebacks, better introspection with the same interface as the pdb module.

#### 分析
```
# -*- coding: utf-8 -*-
from flask import Flask
app = Flask(__name__)

@app.route('/')
def hello_world():
    import ipdb
    ipdb.set_trace()
    return 'Hello World!'

if __name__ == '__main__':
    app.run()
```
在终端执行：
```python
(flask_env) xiezhigang@ pro 1$ ipdb hello.py 

> /home/xiezhigang/PycharmProjects/flask_env/pro/hello.py(1)<module>()
----> 1 from flask import Flask
      2 app = Flask(__name__)
      3 

ipdb> return
 * Running on http://127.0.0.1:5000/ (Press CTRL+C to quit)
```
> Tips: 在python2.7，可以使用python -m ipdb hello.py 

另起一个终端访问：
```python
(flask_env) xiezhigang@ pro 19$ curl 127.0.0.1:5000
```
此时，第一个终端返回：
```python
> /home/xiezhigang/PycharmProjects/flask_env/pro/hello.py(8)hello_world()
      7     ipdb.set_trace()
----> 8     return 'Hello World!'
      9 
```
然后，在第一个终端查看断点的函数调用过程
```python
ipdb> w
  /home/xiezhigang/PycharmProjects/flask_env/bin/ipdb(11)<module>()
      9 if __name__ == '__main__':
     10     sys.argv[0] = re.sub(r'(-script\.pyw?|\.exe)?$', '', sys.argv[0])
---> 11     sys.exit(main())

  /home/xiezhigang/PycharmProjects/flask_env/local/lib/python2.7/site-packages/ipdb/__main__.py(198)main()
    197         try:
--> 198             pdb._runscript(mainpyfile)
    199             if pdb._user_requested_quit:

  /usr/lib/python2.7/pdb.py(1233)_runscript()
   1232         statement = 'execfile(%r)' % filename
-> 1233         self.run(statement)
   1234 

  /usr/lib/python2.7/bdb.py(400)run()
    399         try:
--> 400             exec cmd in globals, locals
    401         except BdbQuit:

  <string>(1)<module>()

  /home/xiezhigang/PycharmProjects/flask_env/pro/hello.py(11)<module>()
      9 
     10 if __name__ == '__main__':
---> 11     app.run()

  /home/xiezhigang/PycharmProjects/flask_env/local/lib/python2.7/site-packages/flask/app.py(841)run()
    840         try:
--> 841             run_simple(host, port, self, **options)
    842         finally:

  /home/xiezhigang/PycharmProjects/flask_env/local/lib/python2.7/site-packages/werkzeug/serving.py(736)run_simple()
    735     else:
--> 736         inner()
    737 

  /home/xiezhigang/PycharmProjects/flask_env/local/lib/python2.7/site-packages/werkzeug/serving.py(699)inner()
    698             log_startup(srv.socket)
--> 699         srv.serve_forever()
    700 

  /home/xiezhigang/PycharmProjects/flask_env/local/lib/python2.7/site-packages/werkzeug/serving.py(536)serve_forever()
    535         try:
--> 536             HTTPServer.serve_forever(self)
    537         except KeyboardInterrupt:

  /usr/lib/python2.7/SocketServer.py(238)serve_forever()
    237                 if self in r:
--> 238                     self._handle_request_noblock()
    239         finally:

  /usr/lib/python2.7/SocketServer.py(295)_handle_request_noblock()
    294             try:
--> 295                 self.process_request(request, client_address)
    296             except:

  /usr/lib/python2.7/SocketServer.py(321)process_request()
    320         """
--> 321         self.finish_request(request, client_address)
    322         self.shutdown_request(request)

  /usr/lib/python2.7/SocketServer.py(334)finish_request()
    333         """Finish one request by instantiating RequestHandlerClass."""
--> 334         self.RequestHandlerClass(request, client_address, self)
    335 

  /usr/lib/python2.7/SocketServer.py(649)__init__()
    648         try:
--> 649             self.handle()
    650         finally:

  /home/xiezhigang/PycharmProjects/flask_env/local/lib/python2.7/site-packages/werkzeug/serving.py(232)handle()
    231         try:
--> 232             rv = BaseHTTPRequestHandler.handle(self)
    233         except (socket.error, socket.timeout) as e:

  /usr/lib/python2.7/BaseHTTPServer.py(340)handle()
    339 
--> 340         self.handle_one_request()
    341         while not self.close_connection:

  /home/xiezhigang/PycharmProjects/flask_env/local/lib/python2.7/site-packages/werkzeug/serving.py(267)handle_one_request()
    266         elif self.parse_request():
--> 267             return self.run_wsgi()
    268 

  /home/xiezhigang/PycharmProjects/flask_env/local/lib/python2.7/site-packages/werkzeug/serving.py(209)run_wsgi()
    208         try:
--> 209             execute(self.server.app)
    210         except (socket.error, socket.timeout) as e:

  /home/xiezhigang/PycharmProjects/flask_env/local/lib/python2.7/site-packages/werkzeug/serving.py(197)execute()
    196         def execute(app):
--> 197             application_iter = app(environ, start_response)
    198             try:

  /home/xiezhigang/PycharmProjects/flask_env/local/lib/python2.7/site-packages/flask/app.py(1997)__call__()
   1996         """Shortcut for :attr:`wsgi_app`."""
-> 1997         return self.wsgi_app(environ, start_response)
   1998 

  /home/xiezhigang/PycharmProjects/flask_env/local/lib/python2.7/site-packages/flask/app.py(1982)wsgi_app()
   1981             try:
-> 1982                 response = self.full_dispatch_request()
   1983             except Exception as e:

  /home/xiezhigang/PycharmProjects/flask_env/local/lib/python2.7/site-packages/flask/app.py(1612)full_dispatch_request()
   1611             if rv is None:
-> 1612                 rv = self.dispatch_request()
   1613         except Exception as e:

  /home/xiezhigang/PycharmProjects/flask_env/local/lib/python2.7/site-packages/flask/app.py(1598)dispatch_request()
   1597         # otherwise dispatch to the handler for that endpoint
-> 1598         return self.view_functions[rule.endpoint](**req.view_args)
   1599 

> /home/xiezhigang/PycharmProjects/flask_env/pro/hello.py(8)hello_world()
      7     ipdb.set_trace()
----> 8     return 'Hello World!'
      9 
```
### Pycharm

> [PyCharm](http://www.jetbrains.com/pycharm/) provides smart code completion, code inspections, on-the-fly error highlighting and quick-fixes, along with automated code refactorings and rich navigation capabilities.

这里使用的Pycharm的版本是2017.1，早在Pycharm4.5就提供了一个Profile功能。
* 在Pycharm的右上角有一排按钮，找一下即可，或者在文件上右击，点击选项“Profile xxx”。
* 默认支持**cProfile**,提供两种视图**Statistics**和**Call Graph**。
* 在Call Graph视图中，右击视图中的方框可以跳转到对应的函数。

下面看看上面这个简单应用的调用过程：
![trace_hello](http://7xp6ii.com1.z0.glb.clouddn.com/hello.png)

## 三. 调用顺序
由ipdb输出的日志，可得知执行hello.py的函数调用链如下：

```bash
run()[flask/app.py] --> run_simple()[werkzeug/serving.py] --> inner()[werkzeug/serving.py]
--> serve_forever()[werkzeug/serving.py] --> serve_forever()[python2.7/SocketServer.py] 
--> _handle_request_noblock()[python2.7/SocketServer.py] 
--> process_request()[python2.7/SocketServer.py] --> handle()[werkzeug/serving.py] 
--> handle()[python2.7/BaseHTTPServer.py] --> handle_one_request()[werkzeug/serving.py] 
--> run_wsgi()[werkzeug/serving.py] --> __call__()[flask/app.py]--> wsgi_app()[flask/app.py] 
--> full_dispatch_request()[flask/app.py] --> dispatch_request()[flask/app.py] 
--> hello_world()[pro/hello.py];
```
由pychamr的Profile功能，可得知执行hello.py的函数调用链如下：
```bash
run()-->run_simple()-->innner()-->serve_forever()-->serve_forever()-->_entry_retry()-->select.select-->fileno
```
## 四. 抽象
从上面的函数调用链，让我们回归到flask本身，抽象werkzeug和werkzeug调用的底层函数，那么抽象后的函数调用链应如下：
```bash
run()[flask/app.py] --> run_simple()[werkzeug/serving.py]--> full_dispatch_request()[flask/app.py] 
--> dispatch_request()[flask/app.py] --> hello_world()[pro/hello.py]
```

## 五. 启动流程
应用启动*app.run()*， 代码如下：

```python
def run(self, host=None, port=None, debug=None, **options):
    """comments have been removed.
    """
    from werkzeug.serving import run_simple
    if host is None:
        host = '127.0.0.1'
    if port is None:
        server_name = self.config['SERVER_NAME']
        if server_name and ':' in server_name:
            port = int(server_name.rsplit(':', 1)[1])
        else:
            port = 5000
    if debug is not None:
        self.debug = bool(debug)
    options.setdefault('use_reloader', self.debug)
    options.setdefault('use_debugger', self.debug)
    try:
        run_simple(host, port, self, **options)
    finally:
        # reset the first request information if the development server
        # reset normally.  This makes it possible to restart the server
        # without reloader and that stuff from an interactive shell.
        self._got_first_request = False
```
这段代码非常简单，无非是处理一下参数，然后调用werkzeug的run_simple函数来处理创建的Flask的appication，注意：run_simple的第三个参数是self。
此处只研究Flask，werkzeug的内在逻辑就不深入探讨了。

```python
    def full_dispatch_request(self):
        """Dispatches the request and on top of that performs request
        pre and postprocessing as well as HTTP exception catching and
        error handling.

        .. versionadded:: 0.7
        """
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
这段代码的核心在于处理请求hooks处理的逻辑和dispatch_request()，以及错误处理和返回响应。

那我们来看看这些hooks函数都做了哪些操作。
```python
    def try_trigger_before_first_request_functions(self):
        """Called before each request and will ensure that it triggers
        the :attr:`before_first_request_funcs` and only exactly once per
        application instance (which means process usually).

        :internal:
        """
        if self._got_first_request:
            return
        with self._before_request_lock:
            if self._got_first_request:
                return
            for func in self.before_first_request_funcs:
                func()
            self._got_first_request = True
```
这段代码的核心在于self._before_request_lock和self.before_first_request_funcs，self._before_request_lock是一个线程锁，保证了一次性只有一个app实例，self.before_first_request_funcs是一个字典，把一系列第一次请求处理前的函数注册在self.before_first_request_funcs中，这个hook是通过before_first_request定义。

执行完第一次请求前的hook函数后，开始发出一个信号（signal，下面的文章会研究一下这个模块），有一个请求进来了。
```python
    def preprocess_request(self):
        """Called before the actual request dispatching and will
        call each :meth:`before_request` decorated function, passing no
        arguments.
        If any of these functions returns a value, it's handled as
        if it was the return value from the view and further
        request handling is stopped.

        This also triggers the :meth:`url_value_preprocessor` functions before
        the actual :meth:`before_request` functions are called.
        """
        bp = _request_ctx_stack.top.request.blueprint

        funcs = self.url_value_preprocessors.get(None, ())
        if bp is not None and bp in self.url_value_preprocessors:
            funcs = chain(funcs, self.url_value_preprocessors[bp])
        for func in funcs:
            func(request.endpoint, request.view_args)

        funcs = self.before_request_funcs.get(None, ())
        if bp is not None and bp in self.before_request_funcs:
            funcs = chain(funcs, self.before_request_funcs[bp])
        for func in funcs:
            rv = func()
            if rv is not None:
                return rv
```
这个hook是在每个请求处理前处理的，通过before_request定义，主要是找到app中定义的bp（蓝图），并注册到self.before_request_funcs。

然后开始正常的处理请求函数
```python
    def dispatch_request(self):
        """Does the request dispatching.  Matches the URL and returns the
        return value of the view or error handler.  This does not have to
        be a response object.  In order to convert the return value to a
        proper response object, call :func:`make_response`.

        .. versionchanged:: 0.7
           This no longer does the exception handling, this code was
           moved to the new :meth:`full_dispatch_request`.
        """
        req = _request_ctx_stack.top.request
        if req.routing_exception is not None:
            self.raise_routing_exception(req)
        rule = req.url_rule
        # if we provide automatic options for this URL and the
        # request came with the OPTIONS method, reply automatically
        if getattr(rule, 'provide_automatic_options', False) \
           and req.method == 'OPTIONS':
            return self.make_default_options_response()
        # otherwise dispatch to the handler for that endpoint
        return self.view_functions[rule.endpoint](**req.view_args)

```
这段代码主要是找到处理函数，返回处理结果，这是路由的全过程。

此外，正常请求处理之后的hook函数--self.finalize_request(rv)，这个函数是通过after_request定义的，还有异常处理的函数。

当一个请求结束时，会执行pop操作，并触发do_teardown_request和teardown_request，因此不管请求是否异常，都会执行teardown_request。
此时一个简单的Flask的hello_world程序正常输出了“Hello World”。