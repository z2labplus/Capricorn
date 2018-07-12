---
layout: blog
title: 'Flask源码剖析（八）：信号'
date: 2017-07-02 23:00:56
categories: flask
tags: flask
lead_text: '信号'
---

## 一. 含义
> 什么是信号？信号通过发送发生在核心框架的其它地方或 Flask 扩展的动作时的通知来帮助你解耦应用。简而言之，信号允许特定的发送端通知订阅者发生了什么。
> 来自于Flask文档：[信号](http://docs.jinkan.org/docs/flask/signals.html#flask)

## 二. 信号
Flask的关于信号的逻辑主要是放在signals.py文件中,先看看代码：
```python
# -*- coding: utf-8 -*-
signals_available = False
try:
    from blinker import Namespace
    signals_available = True
except ImportError:
    class Namespace(object):
        def signal(self, name, doc=None):
            return _FakeSignal(name, doc)

    class _FakeSignal(object):

        def __init__(self, name, doc=None):
            self.name = name
            self.__doc__ = doc
        def _fail(self, *args, **kwargs):
            raise RuntimeError('signalling support is unavailable '
                               'because the blinker library is '
                               'not installed.')
        send = lambda *a, **kw: None
        connect = disconnect = has_receivers_for = receivers_for = \
            temporarily_connected_to = connected_to = _fail
        del _fail

_signals = Namespace()


template_rendered = _signals.signal('template-rendered')
before_render_template = _signals.signal('before-render-template')
request_started = _signals.signal('request-started')
request_finished = _signals.signal('request-finished')
request_tearing_down = _signals.signal('request-tearing-down')
got_request_exception = _signals.signal('got-request-exception')
appcontext_tearing_down = _signals.signal('appcontext-tearing-down')
appcontext_pushed = _signals.signal('appcontext-pushed')
appcontext_popped = _signals.signal('appcontext-popped')
message_flashed = _signals.signal('message-flashed')
```
很明显，Flask的信号依赖于blinker库，如果没有安装blinker库，则创建一个不会报错的虚假信号类。
Flask提供了一系列的核心信号，template_rendered/before_render_template/request_started/request_finished/request_tearing_down/got_request_exception/appcontext_pushed/appcontext_popped/message_flashed，
当程序一启动时就创建了上面的信号。

## 三. 信号调用
来看一个简单的程序，从中研究Flask是怎么创建/发送/订阅信号的。
```python
# -*- coding: utf-8 -*-

from flask import Flask, request_started
app = Flask(__name__)

def log_request(sender, **extra):
    print 'request start test'
    sender.logger.debug('Request context is set up')
    flask.request_finished

@app.route('/')
def hello_world():
    request_started.connect(log_request, app)
    return 'Hello World!'

if __name__ == '__main__':
    app.run()

```
执行以下代码，获取到信号返回：
````python
curl http://127.0.0.1:5000/
````
输出
```python
request start test
127.0.0.1 - - [17/Aug/2017 00:08:54] "GET / HTTP/1.1" 200 -

```
还是从wsgi_app()入口开始，在full_dispatch_request()函数中有这么一段代码，如下：
```python
request_started.send(self)
```
在finalize_request()函数中有一段代码，如下：
```python
request_finished.send(self, response=response)
```
可知，Flask在一个请求进来时，就会发送一个请求开始的信号，在结束请求的操作时，则会发送一个请求结束的信号。
回归到小程序上，程序里是这样订阅信号的
```python
request_started.connect(log_request, app)
```
## 四. 实现
从上文可以知道，Flask的信号机制实现依赖与blinker库，下面来粗略看看这个库是怎么运作的。
```python
xiezhigang@ flask_site_packages 16$ tree -I "*.pyc" blinker
blinker
├── base.py
├── __init__.py
├── _saferef.py
└── _utilities.py

0 directories, 4 files
xiezhigang@ flask_site_packages 17$ cloc blinker
       9 text files.
       9 unique files.                              
       1 file ignored.

http://cloc.sourceforge.net v 1.60  T=0.03 s (262.7 files/s, 38250.6 lines/s)
-------------------------------------------------------------------------------
Language                     files          blank        comment           code
-------------------------------------------------------------------------------
Python                           4            179            283            412
XML                              4              0              0            291
-------------------------------------------------------------------------------
SUM:                             8            179            283            703
-------------------------------------------------------------------------------
```
blinker库总共4个文件，有效代码400多行，是一个很小的库。下面看看blinker的NameSpace类是怎么定义的。
```python
class NamedSignal(Signal):
    """A named generic notification emitter."""

    def __init__(self, name, doc=None):
        Signal.__init__(self, doc)

        #: The name of this signal.
        self.name = name

    def __repr__(self):
        base = Signal.__repr__(self)
        return "%s; %r>" % (base[:-1], self.name)
        
class Namespace(dict):
    """A mapping of signal names to signals."""

    def signal(self, name, doc=None):
        """Return the :class:`NamedSignal` *name*, creating it if required.

        Repeated calls to this function will return the same signal object.

        """
        try:
            return self[name]
        except KeyError:
            return self.setdefault(name, NamedSignal(name, doc))
```
Namespace()这个类是继续了字典类（dict）,如果存在该name，则返回该name对应的key值，key值是一个Signal类实例化对象。
下面看看Signal类定义了哪些方法。
```python
ANY = symbol('ANY')
ANY.__doc__ = 'Token for "any sender".'
ANY_ID = 0


class Signal(object):
    ANY = ANY

    @lazy_property
    def receiver_connected(self):
        return Signal(doc="Emitted after a receiver connects.")

    @lazy_property
    def receiver_disconnected(self):
        return Signal(doc="Emitted after a receiver disconnects.")

    def __init__(self, doc=None):
        if doc:
            self.__doc__ = doc
        self.receivers = {}
        self._by_receiver = defaultdict(set)
        self._by_sender = defaultdict(set)
        self._weak_senders = {}

    def connect(self, receiver, sender=ANY, weak=True):
        receiver_id = hashable_identity(receiver)
        if weak:
            receiver_ref = reference(receiver, self._cleanup_receiver)
            receiver_ref.receiver_id = receiver_id
        else:
            receiver_ref = receiver
        if sender is ANY:
            sender_id = ANY_ID
        else:
            sender_id = hashable_identity(sender)

        self.receivers.setdefault(receiver_id, receiver_ref)
        self._by_sender[sender_id].add(receiver_id)
        self._by_receiver[receiver_id].add(sender_id)
        del receiver_ref

        if sender is not ANY and sender_id not in self._weak_senders:
            # wire together a cleanup for weakref-able senders
            try:
                sender_ref = reference(sender, self._cleanup_sender)
                sender_ref.sender_id = sender_id
            except TypeError:
                pass
            else:
                self._weak_senders.setdefault(sender_id, sender_ref)
                del sender_ref

        if ('receiver_connected' in self.__dict__ and
            self.receiver_connected.receivers):
            try:
                self.receiver_connected.send(self,
                                             receiver=receiver,
                                             sender=sender,
                                             weak=weak)
            except:
                self.disconnect(receiver, sender)
                raise
        if receiver_connected.receivers and self is not receiver_connected:
            try:
                receiver_connected.send(self,
                                        receiver_arg=receiver,
                                        sender_arg=sender,
                                        weak_arg=weak)
            except:
                self.disconnect(receiver, sender)
                raise
        return receiver

    def connect_via(self, sender, weak=False):
        def decorator(fn):
            self.connect(fn, sender, weak)
            return fn
        return decorator

    @contextmanager
    def connected_to(self, receiver, sender=ANY):
        self.connect(receiver, sender=sender, weak=False)
        try:
            yield None
        except:
            self.disconnect(receiver)
            raise
        else:
            self.disconnect(receiver)

    def temporarily_connected_to(self, receiver, sender=ANY):
        warn("temporarily_connected_to is deprecated; "
             "use connected_to instead.",
             DeprecationWarning)
        return self.connected_to(receiver, sender)

    def send(self, *sender, **kwargs):
        if len(sender) == 0:
            sender = None
        elif len(sender) > 1:
            raise TypeError('send() accepts only one positional argument, '
                            '%s given' % len(sender))
        else:
            sender = sender[0]
        if not self.receivers:
            return []
        else:
            return [(receiver, receiver(sender, **kwargs))
                    for receiver in self.receivers_for(sender)]

    def has_receivers_for(self, sender):
        if not self.receivers:
            return False
        if self._by_sender[ANY_ID]:
            return True
        if sender is ANY:
            return False
        return hashable_identity(sender) in self._by_sender

    def receivers_for(self, sender):
        if self.receivers:
            sender_id = hashable_identity(sender)
            if sender_id in self._by_sender:
                ids = (self._by_sender[ANY_ID] |
                       self._by_sender[sender_id])
            else:
                ids = self._by_sender[ANY_ID].copy()
            for receiver_id in ids:
                receiver = self.receivers.get(receiver_id)
                if receiver is None:
                    continue
                if isinstance(receiver, WeakTypes):
                    strong = receiver()
                    if strong is None:
                        self._disconnect(receiver_id, ANY_ID)
                        continue
                    receiver = strong
                yield receiver

    def disconnect(self, receiver, sender=ANY):
        if sender is ANY:
            sender_id = ANY_ID
        else:
            sender_id = hashable_identity(sender)
        receiver_id = hashable_identity(receiver)
        self._disconnect(receiver_id, sender_id)

        if ('receiver_disconnected' in self.__dict__ and
            self.receiver_disconnected.receivers):
            self.receiver_disconnected.send(self,
                                            receiver=receiver,
                                            sender=sender)

    def _disconnect(self, receiver_id, sender_id):
        if sender_id == ANY_ID:
            if self._by_receiver.pop(receiver_id, False):
                for bucket in self._by_sender.values():
                    bucket.discard(receiver_id)
            self.receivers.pop(receiver_id, None)
        else:
            self._by_sender[sender_id].discard(receiver_id)
            self._by_receiver[receiver_id].discard(sender_id)

    def _cleanup_receiver(self, receiver_ref):
        self._disconnect(receiver_ref.receiver_id, ANY_ID)

    def _cleanup_sender(self, sender_ref):
        sender_id = sender_ref.sender_id
        assert sender_id != ANY_ID
        self._weak_senders.pop(sender_id, None)
        for receiver_id in self._by_sender.pop(sender_id, ()):
            self._by_receiver[receiver_id].discard(sender_id)

    def _cleanup_bookkeeping(self):
        for mapping in (self._by_sender, self._by_receiver):
            for _id, bucket in list(mapping.items()):
                if not bucket:
                    mapping.pop(_id, None)

    def _clear_state(self):
        self._weak_senders.clear()
        self.receivers.clear()
        self._by_sender.clear()
        self._by_receiver.clear()

```
Signal类实现了一个信号发射器功能，包含创建/发送/订阅/断开/清除等方法。