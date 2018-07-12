---
layout: blog
title: 'Flask源码剖析（十三）：扩展'
date: 2017-08-11 17:35:09
categories: flask
tags: flask
lead_text: '扩展'
---

## 一. 扩展
Flask通常需要一些第三方库来工作，这些第三库通常可以分离出来。

## 二. 代码
关于载入扩展的代码存放于flask/ext/\_\_init__.py和flask/exthook.py文件中，分别看看这两个文件
1. flask/ext/\_\_init__.py
```python
# -*- coding: utf-8 -*-

def setup():
    from ..exthook import ExtensionImporter
    importer = ExtensionImporter(['flask_%s', 'flaskext.%s'], __name__)
    importer.install()


setup()
del setup
```
这段代码很简单，允许扩展名为flask_name或者flaskext.name的命名包，并安装此扩展包。

2. flask/exthook.py
```python
# -*- coding: utf-8 -*-
import sys
import os
import warnings
from ._compat import reraise


class ExtDeprecationWarning(DeprecationWarning):
    pass

warnings.simplefilter('always', ExtDeprecationWarning)


class ExtensionImporter(object):

    def __init__(self, module_choices, wrapper_module):
        self.module_choices = module_choices
        self.wrapper_module = wrapper_module
        self.prefix = wrapper_module + '.'
        self.prefix_cutoff = wrapper_module.count('.') + 1

    def __eq__(self, other):
        return self.__class__.__module__ == other.__class__.__module__ and \
               self.__class__.__name__ == other.__class__.__name__ and \
               self.wrapper_module == other.wrapper_module and \
               self.module_choices == other.module_choices

    def __ne__(self, other):
        return not self.__eq__(other)

    def install(self):
        sys.meta_path[:] = [x for x in sys.meta_path if self != x] + [self]

    def find_module(self, fullname, path=None):
        if fullname.startswith(self.prefix) and \
           fullname != 'flask.ext.ExtDeprecationWarning':
            return self
            
    def load_module(self, fullname):

        if fullname in sys.modules:
            return sys.modules[fullname]

        modname = fullname.split('.', self.prefix_cutoff)[self.prefix_cutoff]

        warnings.warn(
            "Importing flask.ext.{x} is deprecated, use flask_{x} instead."
            .format(x=modname), ExtDeprecationWarning, stacklevel=2
        )

        for path in self.module_choices:
            realname = path % modname
            try:
                __import__(realname)
            except ImportError:
                exc_type, exc_value, tb = sys.exc_info()

                sys.modules.pop(fullname, None)

                if self.is_important_traceback(realname, tb):
                    reraise(exc_type, exc_value, tb.tb_next)
                continue
            module = sys.modules[fullname] = sys.modules[realname]
            if '.' not in modname:
                setattr(sys.modules[self.wrapper_module], modname, module)

            if realname.startswith('flaskext.'):
                warnings.warn(
                    "Detected extension named flaskext.{x}, please rename it "
                    "to flask_{x}. The old form is deprecated."
                    .format(x=modname), ExtDeprecationWarning
                )

            return module
        raise ImportError('No module named %s' % fullname)

    def is_important_traceback(self, important_module, tb):
        while tb is not None:
            if self.is_important_frame(important_module, tb):
                return True
            tb = tb.tb_next
        return False

    def is_important_frame(self, important_module, tb):
        g = tb.tb_frame.f_globals
        if '__name__' not in g:
            return False

        module_name = g['__name__']

        if module_name == important_module:
            return True

        filename = os.path.abspath(tb.tb_frame.f_code.co_filename)
        test_string = os.path.sep + important_module.replace('.', os.path.sep)
        return test_string + '.py' in filename or \
               test_string + os.path.sep + '__init__.py' in filename
```
ExtensionImporter类主要是重导向flaskext.foo，用flask_foo替代，不需要升级旧的第三方库。

## 三. 解析
flask的扩展模块的导入实现非常简洁，主要是以下几个步骤，当导入from falsk.ext.foo import F时
1. 首先执行的是ext/\_\_.init__.py文件，实例化对象importer将调用install函数，向sys.meta_path添加模块装载类
2. 当import时会调用其find_module，如果返回非None,会调用load_module加载