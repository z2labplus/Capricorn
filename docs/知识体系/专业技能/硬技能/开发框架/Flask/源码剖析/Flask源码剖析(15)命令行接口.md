---
layout: blog
title: 'Flask源码剖析（十五）：命令行接口'
date: 2017-08-25 22:15:12
categories: flask
tags: flask
lead_text: '命令行接口'
---

## 一. 命令行接口
Flask提供了一个脚本，对flask程序上的Flask.cli所有实例的所有命令进行访问，以及一些内置的命令。
当然，还可以注册更多的命令。

## 二. 代码
关于命令行接口的所有代码都放在flask/cli.py文件中，如下：
```python
# -*- coding: utf-8 -*-
import os
import sys
from threading import Lock, Thread
from functools import update_wrapper

import click

from ._compat import iteritems, reraise
from .helpers import get_debug_flag
from . import __version__

class NoAppException(click.UsageError):


def find_best_app(module):
    from . import Flask

    for attr_name in 'app', 'application':
        app = getattr(module, attr_name, None)
        if app is not None and isinstance(app, Flask):
            return app

    matches = [v for k, v in iteritems(module.__dict__)
               if isinstance(v, Flask)]

    if len(matches) == 1:
        return matches[0]
    raise NoAppException('Failed to find application in module "%s".  Are '
                         'you sure it contains a Flask application?  Maybe '
                         'you wrapped it in a WSGI middleware or you are '
                         'using a factory function.' % module.__name__)


def prepare_exec_for_file(filename):
    module = []

    if os.path.split(filename)[1] == '__init__.py':
        filename = os.path.dirname(filename)
    elif filename.endswith('.py'):
        filename = filename[:-3]
    else:
        raise NoAppException('The file provided (%s) does exist but is not a '
                             'valid Python file.  This means that it cannot '
                             'be used as application.  Please change the '
                             'extension to .py' % filename)
    filename = os.path.realpath(filename)

    dirpath = filename
    while 1:
        dirpath, extra = os.path.split(dirpath)
        module.append(extra)
        if not os.path.isfile(os.path.join(dirpath, '__init__.py')):
            break

    sys.path.insert(0, dirpath)
    return '.'.join(module[::-1])


def locate_app(app_id):
    __traceback_hide__ = True
    if ':' in app_id:
        module, app_obj = app_id.split(':', 1)
    else:
        module = app_id
        app_obj = None

    try:
        __import__(module)
    except ImportError:
        raise NoAppException('The file/path provided (%s) does not appear to '
                             'exist.  Please verify the path is correct.  If '
                             'app is not on PYTHONPATH, ensure the extension '
                             'is .py' % module)
    mod = sys.modules[module]
    if app_obj is None:
        app = find_best_app(mod)
    else:
        app = getattr(mod, app_obj, None)
        if app is None:
            raise RuntimeError('Failed to find application in module "%s"'
                               % module)

    return app


def find_default_import_path():
    app = os.environ.get('FLASK_APP')
    if app is None:
        return
    if os.path.isfile(app):
        return prepare_exec_for_file(app)
    return app


def get_version(ctx, param, value):
    if not value or ctx.resilient_parsing:
        return
    message = 'Flask %(version)s\nPython %(python_version)s'
    click.echo(message % {
        'version': __version__,
        'python_version': sys.version,
    }, color=ctx.color)
    ctx.exit()

version_option = click.Option(['--version'],
                              help='Show the flask version',
                              expose_value=False,
                              callback=get_version,
                              is_flag=True, is_eager=True)

class DispatchingApp(object):

    def __init__(self, loader, use_eager_loading=False):
        self.loader = loader
        self._app = None
        self._lock = Lock()
        self._bg_loading_exc_info = None
        if use_eager_loading:
            self._load_unlocked()
        else:
            self._load_in_background()

    def _load_in_background(self):
        def _load_app():
            __traceback_hide__ = True
            with self._lock:
                try:
                    self._load_unlocked()
                except Exception:
                    self._bg_loading_exc_info = sys.exc_info()
        t = Thread(target=_load_app, args=())
        t.start()

    def _flush_bg_loading_exception(self):
        __traceback_hide__ = True
        exc_info = self._bg_loading_exc_info
        if exc_info is not None:
            self._bg_loading_exc_info = None
            reraise(*exc_info)

    def _load_unlocked(self):
        __traceback_hide__ = True
        self._app = rv = self.loader()
        self._bg_loading_exc_info = None
        return rv

    def __call__(self, environ, start_response):
        __traceback_hide__ = True
        if self._app is not None:
            return self._app(environ, start_response)
        self._flush_bg_loading_exception()
        with self._lock:
            if self._app is not None:
                rv = self._app
            else:
                rv = self._load_unlocked()
            return rv(environ, start_response)


class ScriptInfo(object):

    def __init__(self, app_import_path=None, create_app=None):
        if create_app is None:
            if app_import_path is None:
                app_import_path = find_default_import_path()
            self.app_import_path = app_import_path
        else:
            app_import_path = None

        self.app_import_path = app_import_path
        self.create_app = create_app
        self.data = {}
        self._loaded_app = None

    def load_app(self):
        __traceback_hide__ = True
        if self._loaded_app is not None:
            return self._loaded_app
        if self.create_app is not None:
            rv = self.create_app(self)
        else:
            if not self.app_import_path:
                raise NoAppException(
                    'Could not locate Flask application. You did not provide '
                    'the FLASK_APP environment variable.\n\nFor more '
                    'information see '
                    'http://flask.pocoo.org/docs/latest/quickstart/')
            rv = locate_app(self.app_import_path)
        debug = get_debug_flag()
        if debug is not None:
            rv.debug = debug
        self._loaded_app = rv
        return rv


pass_script_info = click.make_pass_decorator(ScriptInfo, ensure=True)


def with_appcontext(f):
    @click.pass_context
    def decorator(__ctx, *args, **kwargs):
        with __ctx.ensure_object(ScriptInfo).load_app().app_context():
            return __ctx.invoke(f, *args, **kwargs)
    return update_wrapper(decorator, f)


class AppGroup(click.Group):

    def command(self, *args, **kwargs):
        wrap_for_ctx = kwargs.pop('with_appcontext', True)
        def decorator(f):
            if wrap_for_ctx:
                f = with_appcontext(f)
            return click.Group.command(self, *args, **kwargs)(f)
        return decorator

    def group(self, *args, **kwargs):
        kwargs.setdefault('cls', AppGroup)
        return click.Group.group(self, *args, **kwargs)


class FlaskGroup(AppGroup):

    def __init__(self, add_default_commands=True, create_app=None,
                 add_version_option=True, **extra):
        params = list(extra.pop('params', None) or ())

        if add_version_option:
            params.append(version_option)

        AppGroup.__init__(self, params=params, **extra)
        self.create_app = create_app

        if add_default_commands:
            self.add_command(run_command)
            self.add_command(shell_command)

        self._loaded_plugin_commands = False

    def _load_plugin_commands(self):
        if self._loaded_plugin_commands:
            return
        try:
            import pkg_resources
        except ImportError:
            self._loaded_plugin_commands = True
            return

        for ep in pkg_resources.iter_entry_points('flask.commands'):
            self.add_command(ep.load(), ep.name)
        self._loaded_plugin_commands = True

    def get_command(self, ctx, name):
        self._load_plugin_commands()

        rv = AppGroup.get_command(self, ctx, name)
        if rv is not None:
            return rv

        info = ctx.ensure_object(ScriptInfo)
        try:
            rv = info.load_app().cli.get_command(ctx, name)
            if rv is not None:
                return rv
        except NoAppException:
            pass

    def list_commands(self, ctx):
        self._load_plugin_commands()

        rv = set(click.Group.list_commands(self, ctx))
        info = ctx.ensure_object(ScriptInfo)
        try:
            rv.update(info.load_app().cli.list_commands(ctx))
        except Exception:
            pass
        return sorted(rv)

    def main(self, *args, **kwargs):
        obj = kwargs.get('obj')
        if obj is None:
            obj = ScriptInfo(create_app=self.create_app)
        kwargs['obj'] = obj
        kwargs.setdefault('auto_envvar_prefix', 'FLASK')
        return AppGroup.main(self, *args, **kwargs)


@click.command('run', short_help='Runs a development server.')
@click.option('--host', '-h', default='127.0.0.1',
              help='The interface to bind to.')
@click.option('--port', '-p', default=5000,
              help='The port to bind to.')
@click.option('--reload/--no-reload', default=None,
              help='Enable or disable the reloader.  By default the reloader '
              'is active if debug is enabled.')
@click.option('--debugger/--no-debugger', default=None,
              help='Enable or disable the debugger.  By default the debugger '
              'is active if debug is enabled.')
@click.option('--eager-loading/--lazy-loader', default=None,
              help='Enable or disable eager loading.  By default eager '
              'loading is enabled if the reloader is disabled.')
@click.option('--with-threads/--without-threads', default=False,
              help='Enable or disable multithreading.')
@pass_script_info
def run_command(info, host, port, reload, debugger, eager_loading,
                with_threads):
    from werkzeug.serving import run_simple

    debug = get_debug_flag()
    if reload is None:
        reload = bool(debug)
    if debugger is None:
        debugger = bool(debug)
    if eager_loading is None:
        eager_loading = not reload

    app = DispatchingApp(info.load_app, use_eager_loading=eager_loading)

    if os.environ.get('WERKZEUG_RUN_MAIN') != 'true':
        if info.app_import_path is not None:
            print(' * Serving Flask app "%s"' % info.app_import_path)
        if debug is not None:
            print(' * Forcing debug mode %s' % (debug and 'on' or 'off'))

    run_simple(host, port, app, use_reloader=reload,
               use_debugger=debugger, threaded=with_threads)


@click.command('shell', short_help='Runs a shell in the app context.')
@with_appcontext
def shell_command():
    import code
    from flask.globals import _app_ctx_stack
    app = _app_ctx_stack.top.app
    banner = 'Python %s on %s\nApp: %s%s\nInstance: %s' % (
        sys.version,
        sys.platform,
        app.import_name,
        app.debug and ' [debug]' or '',
        app.instance_path,
    )
    ctx = {}

    startup = os.environ.get('PYTHONSTARTUP')
    if startup and os.path.isfile(startup):
        with open(startup, 'r') as f:
            eval(compile(f.read(), startup, 'exec'), ctx)

    ctx.update(app.make_shell_context())

    code.interact(banner=banner, local=ctx)


cli = FlaskGroup(help="""\
This shell command acts as general utility script for Flask applications.

It loads the application configured (through the FLASK_APP environment
variable) and then provides commands either provided by the application or
Flask itself.

The most useful commands are the "run" and "shell" command.

Example usage:

\b
  %(prefix)s%(cmd)s FLASK_APP=hello.py
  %(prefix)s%(cmd)s FLASK_DEBUG=1
  %(prefix)sflask run
""" % {
    'cmd': os.name == 'posix' and 'export' or 'set',
    'prefix': os.name == 'posix' and '$ ' or '',
})


def main(as_module=False):
    this_module = __package__ + '.cli'
    args = sys.argv[1:]

    if as_module:
        if sys.version_info >= (2, 7):
            name = 'python -m ' + this_module.rsplit('.', 1)[0]
        else:
            name = 'python -m ' + this_module

        sys.argv = ['-m', this_module] + sys.argv[1:]
    else:
        name = None

    cli.main(args=args, prog_name=name)


if __name__ == '__main__':
    main(as_module=True)

```

## 三. 解析
flask/cli.py文件中主要定义了四个类，分别是DispatchingApp/ScriptInfo/AppGroup/FlaskGroup，来看看这四个类是怎么定义的。
1. DispatchingApp
- 先看看DispatchingApp是怎么被调用的
```python
app = DispatchingApp(info.load_app, use_eager_loading=eager_loading)
```
- 第一个参数info.load_app是一个Flask程序的实例，第二个参数use_eager_loading是一个bool值
- 当use_eager_loading为True，调用self._load_unlocked()，直接返回loader，use_eager_loading为False，调用self._load_in_background()，启用线程锁
- 注意DispatchingApp类使用了__call__方法，允许一个类的实例像函数一样被调用，具体使用可参考[Python 魔术方法指南](http://pycoders-weekly-chinese.readthedocs.io/en/latest/issue6/a-guide-to-pythons-magic-methods.html)

2. ScriptInfo
- ScriptInfo类只有初始化函数__init__，载入flask程序实例的函数load_app,当
- 当ScriptInfo类初始化时，create_app=None，则调用find_default_import_path从环境变量中获取FLASK_APP的值，即Flask程序的路径，prepare_exec_for_file函数是判断此路径是否合法
- load_app函数是判断flask程序实例是否存在，如果不存在，则创建一个新的flask程序实例，并调用find_best_app函数判断此实例是否flask程序

3. AppGroup
- AppGroup类集成了click库的Group类，重写了command和group方法
- command函数主要是从当前的环境中获取上下文，从with_appcontext函数可知
- group函数给AppGroup类赋予一个默认的key值cls

4. FlaskGroup
- FlaskGroup类继承了AppGroup类，重写了get_command和list_commands方法
- get_command方法首先载入扩展的flask命令行选项，再从当前的上下文中获取命令，如果不为空，则返回当前上下文中的命令行选项，为空则重新载入当前的flask程序
- list_commands方法也是首先载入扩展的flask命令行选项，再从当前的上下文中获取命令，然后重新加载当前的flask程序，获取到命令行选项并更新字典，返回排序后的命令行

## 四. 应用
命令行接口的使用可参考[命令行接口](http://python.usyiyi.cn/documents/flask_011_ch/cli.html)。
