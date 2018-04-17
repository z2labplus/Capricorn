## Python源码剖析学习笔记

### 一. Python源码剖析——编译Python
#### Python总体架构
![](./images/Python总体架构.png)



在最高的层次上，Python的整体架构可以分为三个主要的部分，如上。

图左，是Python提供的大量的模块/库以及用户自定义的模块。比如在执行`import os`时，这个`os`是Python内建的模块，用户还可以通过自定义模块来扩展Python系统。

图右，是Python的运行时环境，包括对象/类型系统（Object/Type structures），内存分配器（Memory Alloator）和运行时状态信息（Current State of Python）。

- 对象/类型系统则包含了Python中存在的各种内建对象，比如int/list/dict，以及用户自定义的各种类型和对象。


- 内存分配器则全权负责Python中创建对象时，对内存的申请工作，实际上它就是Python运行时与C中`malloc`的一层接口。
- 运行时状态维护了解释器在执行字节码时不同的状态之间的切换的动作，可视它为一个巨大而复杂的有穷状态机。

图中，是Python的核心——解释器（interpreter），或者称为虚拟机。在解释器中，箭头的方向指示了Python运行过程中的数据流方向。

- Scanner对应词法分析，讲文件输入的Python源代码或者从命令行输入的一行行Python代码切分为一个个的token
- Parser对应语法分析，在Scanner的分析结果上进行语法分析，建立抽象语法树（AST）
- Compiler是根据建立的AST生成指令集合——Python字节码
- Code Evaluator执行以上的字节码，因此，Code Evaluator又可称为虚拟机

以上，在解释器与右边的对象/类型系统，内存分配器之间的箭头表示“使用”关系；与运行时状态之间的箭头表示“修改”关系，即是Python在运行的过程中会不断修改当前解释器所处的状态，在不同状态之间切换。

#### Python源代码的组织

![](./images/Python目录结构.png)

`Include`：该目录包含了Python提供的所有头文件，如果用户需要自己用C或C++来编写自定义模块来扩展Python，那么需要这里提供的头文件。

`Lib`：该目录包含了Python自带的所有标准库，Lib中的库都是用Python语言写的。

`Modules`：该目录包含了所有用C语言编写的模块，`Modules`中的模块时那些对速度要求非常严格的模块，而有一些对速度没有太严格要求的模块，则用Python语言编写。

`Parser`：该目录包含了Python解释器中的Scanner和Parser，即对Python源代码进行词法分析和语法分析的部分。除了这些，Parser目录下还包含了一些有用的工具，这些工具可以根据Python语言的语法自动生成Python语言的词法和语法分析器。

`Objects`：该目录包含了Python的所有内建对象，包含整数/list/dict等。同时，该目录还包括了Python在运行时需要的所有的内部使用对象的实现。

`Python`：该目录包含了Python解释器中的Compiler和执行引擎部分，是Python运行核心所在。

`PCBuild`：该目录包含了Visual Studio 2003使用的工程文件，研究Python源代码就从这里开始。

`PCBuild8`：该目录包含了Visual Studio 2005使用的工程文件。



### 二. Python对象初探

#### Python内的对象

##### 1. 对象的概念

- 对于人的思维来说，对象是一个形象的概念，而对于计算机，对象却是一个抽象的概念，它所知道的只是字节。
- 通常来说，对象是数据以及基于这些数据的操作的集合。
- 在计算机中，一个对象实际上就是一片被分配的内存空间，这些内存有可能是连续的，也有可能是离散的，在更高层次上，这些内存可以当作一个整体来考虑，这个整体就是对象。
- 在这片内存中，存储着一系列的数据以及可以对这些数据进行修改或者读取操作的一系列代码。

##### 2. 对象的特点

- 在Python中，对象就是为C中的结构体在堆上申请的一块内存。
- 在Python中，所有的内建的类型对象都是被静态初始化的。（一般来说，对象是不能被静态初始化的，并且不能在栈空间上生存。唯一的例外就是类型对象）
- 在Python中，一个对象一旦被创建，它在内存中的大小就是不变的了。（这就意味着那些需要容纳可变长度数据的对象只能在对象内维护一个指向一块可变大小的内存区域的指针）

##### 3. 对象的分类

Python的对象从概念上可以大致分为五类，这种分类不一定正确，不过可以提供另外一个角度看待Python中的对象。

- Fundamental 对象：类型对象

- Numeric 对象：数值对象

- Sequence 对象：容纳其他对象的序列集合对象

- Mapping 对象：类似于C++中的map的关联对象

- Internal 对象：Python的虚拟机在运行时内部使用的对象

  ![](./images/Python对象分类.jpeg)

##### 4. 对象机制的基石

在Python中，所有的东西都是对象，而所有的对象都拥有一些相同的内容（这句话的另外意思是，每一个Python对象除了必须有这个PyObject内容外，还占有额外的内存，放置其他内容），这些内容在PyObject中定义，出现在每一个Python对象所占有的内存的最开始的字节中，PyObject是整个Python对象机制的核心。

PyObject定义如下：

![](./images/PyObject定义.png)

- 在PyObject定义中，整型变量`ob_refcnt`与Python的内存管理机制有关，它实现了基于引用计数的垃圾回收机制。
- 在`ob_refcnt`之外，还有一个`ob_type`指向结构体`_typeobject`的指针，这个结构体对应着Python内部的一种特殊对象，它是用来指定一个对象类型的类型对象。

在Python中，对象机制的核心非常简单，一个是引用计数，一个是类型信息。

##### 5. 定长对象和变长对象

不包括可变长数据的对象称为定长对象（例如整数对象）。

包括可变长数据的对象称为变长对象（例如字符串对象）。

区别在于定长对象的不同对象占用的内存大小是一样的，变长对象的不同对象占用的内存大小是不一样的。

##### 6. 可变对象和不可变对象

可变对象是一旦创建后还可改变，但是地址不会发生改变，即该变量指向原来的对象。（例如int，string，float，tuple）

不可变对象是一旦创建后不可改变，如果更改，则变量会指向一个新的对象。（例如list，dict）

#### 类型对象

##### 1. 对象的元信息

结构体`_typeobject`：

![](./images/PyTypeObject定义.png)

在`_typeobject`的定义中包含了很多信息，主要分为四类：

- 类型名，`tp_name`
- 创建该类型对象时分配的内存空间大小信息，`tp_basicsize`和`tp_itemsize`
- 与该类型对象相关联的操作信息，`tp_print`等函数指针
- 描述该类型对象的类型信息

##### 2. 对象的创建

一般来说，Python创建对象时会有两种方法，一种是通过`Python C API`来创建，一种是通过类型对象`PyInt_Type`。

`Python C API`分成两类：

一类称为范型的API，或者称为AOL（Abstract Object Layer）。这类API都具有PyObject_***的形式，可应用在任何Python对象上。

另一类是与类型相关的API，或者称为COL（Concrete Object Layer）。这类API通常只能作用在某一种类型的对象上，对于每一种内建对象，Python都提供了这样的一组API。

无论是使用哪一种`Python C API`，Python内部都是直接分配内存的。

![](./images/PyInt_Type创建整数对象.png)

##### 3. 对象的行为

在PytypeObject中定义了大量的函数指针，这些函数指针最终会指向某个函数，或者指向NULL。这些函数可以视为类型对象中所定义的操作，而这些操作直接决定着一个对象在运行时所表现出的行为。

##### 4. 类型的类型

Python的类型对象`PyTypeObject`也是一个对象，是由`PyType_Type`创建的。

`PyType_Type`是Python类型机制中一个至关重要的对象，所有用户自定义class所对应的`PyTypeObject`对象都是由这个对象创建。

`PyType_Type`是所有class的class，在Python中被称为`metaclass`。

#### 多态性

在Python中创建对象，比如`PyIntObject`对象时，会分配内存，进行初始化。Python内部会用一个`PyObject *`变量来保存和维护这个对象，而不是`PyIntObject *`，其他对象也与此类似。

因此，在Python内部各个函数之间传递的是一种范型指针`PyObject *`，这个指针所指对象的`ob_type`域动态进行判断，通过这个域，Python实现了多态机制。

#### 对象的引用计数

Python通过对一个对象的引用计数的管理来维护对象在内存中的存在与否。Python的每个对象都有`ob_refcnt`变量，这个变量维护着对象的引用计数，从而决定着该对象的创建与消亡。

在Python中，主要是通过`Py_INCREF(op)`和`Py_DECREF(op)`两个宏来增加和减少一个对象的引用计数。当一个对象的引用计数减少到0之后，`Py_DECREF`将调用该对象的析构函数释放该对象所占有的内存和系统资源。

此处调用析构函数并不意味着最终会调用`free`函数来释放内存空间，如果这样做的话，频繁申请内存和释放内存，会导致Python的执行效率大打折扣。

一般来说，Python中大量采用了内存对象池的技术，调用析构函数时，通常是将该对象所占有的空间归还给内存池中，避免了频繁地申请和释放内存。

> Tips: 在Python的各种对象中，类型对象是超越引用计数规则的，永远不会被析构，每一个对象中指向类型对象的指针不会被视为对该类型对象的引用。



### 三. Python中的整数对象

#### 初识PyIntObject对象

##### 1. PyIntObject对象的定义 

![](./images/PyIntObject定义.png)

- Python中的整数对象`PyIntObject`实际上是C中原生类型long的一个简单包装。
- Python中的对象的相关元信息实际上都是保存在对应的类型对象中的，对于`PyIntObject`，类型对象是`PyInt_Type`。

#### PyIntObjecy对象的创建和维护

##### 1. 对象创建的三种途径

- 从long值生成`PyIntObject`对象
- 从字符串生成`PyIntObject`对象
- 从Py_UNICODE对象生成`PyIntObject`对象

##### 2. 小整数对象

在Python对象中，所有的对象都是生活在堆上，如果没有特殊的机制的话，那么Python将一次又一次使用malloc在堆上申请空间和释放空间，基于此种情况，对于小整数引入了对象池技术。

在Python2.5中，将小整数集合的范围默认设定为[-5, 257]，可修改NSMALLNEGINTS和NSMALLPOSINTS的值，重新编译Python，从而将这个范围向两端伸展或收缩。

##### 3. 大整数对象

对于小整数，在小整数对象池中完全地缓存了其`PyIntObject`对象，而对其他整数，Python运行环境将提供一块内存空间，这些内存空间将由大整数轮流使用。

在Python中，有一个`PyIntBlock`结构，在这个结构的基础上，实现了一个单向列表。

##### 4. 添加和删除

`PyIntObject`对象的创建通过两部完成：

- 如果小整数对象池被激活，则尝试小整数对象池
- 如果不能使用小整数对象池，则使用通用的整数对象池



### 四. Python中的字符串对象

#### PyStringObject和PyString_Type

##### 1. PyStringObject对象的定义 

![](./images/PyStringObject对象的定义.png)

- 对于`PyStringObject`，类型对象是`PyString_Type`。
- ​

#### 创建PyStringObject对象

#### 字符串对象的intern机制

#### 字符串缓冲池

#### PyStringObject效率相关问题



### 五. Python中的List对象

### 六. Python中的Dict对象

### 七. 最简单的Python模拟

### 八. Python的编译结果：code对象与pyc文件

### 九. Python的虚拟机框架

### 十. Python虚拟机中的一般表达式

### 十一. Python虚拟机中的控制流

### 十二. Python虚拟机中的函数机制

### 十三. Python虚拟机中的类机制

### 十四. Python运行环境初始化

### 十五. Python模块的动态加载机制

### 十六. Python多线程机制

### 十七. Python的内存管理机制