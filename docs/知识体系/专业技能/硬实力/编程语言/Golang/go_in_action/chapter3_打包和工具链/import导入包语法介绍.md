## import导入包语法介绍

#### 本地 `GOROO` 标准库导入
```go
import(
    "fmt"
)
```

调用:
```go
fmt.Println( "Hello Go!")
```

#### 本地模块导入
- 相对路径

当前文件同一目录的 `view` 目录，不建议这种方式
```go
import   "./view"
```

- 绝对路径

加载 `GOPATH/src/shorturl/view` 模块
```go
import   "shorturl/view"
```

### 远程导入

Go语言的工具链支持从分布式版本控制系统获取源代码，会使用导入路径确定需要获取的代码在网络的什么地方。
```go
import "github.com/spf13/viper"
```

> 这个获取过程 使用 `go get` 命令完成。`go get` 将获取任意指定的 `URL` 的包，或者一个已经导入的包所依赖的其他包。
> 由于`go get`的这种递归特性，这个命令会扫描某个包的源码树，获取能找到的所有依赖包。

### 点导入

这个点操作的含义就是这个包导入之后在你调用这个包的函数时，你可以省略前缀的包名。
```go
import( 
    . "fmt" 
) 
```

即是上面包`fmt`的调用:
```go
fmt.Println( "Hello Go!")
```
可以省略的写成:
```go
Println( "Hello Go!")
```

### 别名导入

别名操作顾名思义可以把包命名成另一个用起来容易记忆的名字:
```go
import( 
    f "fmt" 
) 
```

调用:
```go
f.Println( "Hello Go!")
```

### 下划线导入

用户可能需要导入一个包，但是不需要引用这个包的标识符，但Go语言不能导入不使用的包，在这种情况，可以使用空白标识符`_`来重命名这个导入。
当导入一个包时，它所有的 `init()` 函数就会被执行。
```go
import ( 
    _ “github.com/ziutek/mymysql/godrv” 
) 
```