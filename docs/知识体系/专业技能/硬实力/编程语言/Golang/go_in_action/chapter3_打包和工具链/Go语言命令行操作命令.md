## Go语言命令行操作命令

### Go版本号
```go
go version go1.10.1 darwin/amd64
```

### Go命令操作工具

```go
Go is a tool for managing Go source code.

Usage:

	go command [arguments]

The commands are:

	build       compile packages and dependencies
	clean       remove object files and cached files
	doc         show documentation for package or symbol
	env         print Go environment information
	bug         start a bug report
	fix         update packages to use new APIs
	fmt         gofmt (reformat) package sources
	generate    generate Go files by processing source
	get         download and install packages and dependencies
	install     compile and install packages and dependencies
	list        list packages
	run         compile and run Go program
	test        test packages
	tool        run specified go tool
	version     print Go version
	vet         report likely mistakes in packages

Use "go help [command]" for more information about a command.

Additional help topics:

	c           calling between Go and C
	buildmode   build modes
	cache       build and test caching
	filetype    file types
	gopath      GOPATH environment variable
	environment environment variables
	importpath  import path syntax
	packages    package lists
	testflag    testing flags
	testfunc    testing functions

Use "go help [topic]" for more information about that topic.
```

#### go build

用于测试编译。在包的编译过程中，若有必要，会同时编译与之相关联的包。

- 如果是普通包，当你执行go build之后，它不会产生任何文件。如果你需要在$GOPATH/pkg下生成相应的文件，得执行go install。
- 如果是main包，当你执行go build之后，它就会在当前目录下生成一个可执行文件。如果你需要在$GOPATH/bin下生成相应的文件，需要执行go install，或者使用go build -o 路径/a.exe。
- go build命令默认会编译当前目录下的所有go文件，如果某个项目文件夹下有多个文件，而你只想编译某个文件，就可在go build之后加上文件名
- go build会忽略目录下以“_”或“.”开头的go文件。
- 如果你的源代码针对不同的操作系统需要不同的处理，那么你可以根据不同的操作系统后缀来命名文件。go build的时候会选择性地编译以系统名结尾的文件（Linux、Darwin、Windows、Freebsd）。例如Linux系统下面编译只会选择array_linux.go文件，其它系统命名后缀文件全部忽略。

#### go clean

用来移除当前源码包里面编译生成的文件。

这些文件包括：
```go
_obj/            旧的object目录，由Makefiles遗留
_test/           旧的test目录，由Makefiles遗留
_testmain.go     旧的gotest文件，由Makefiles遗留
test.out         旧的test记录，由Makefiles遗留
build.out        旧的test记录，由Makefiles遗留
*.[568ao]        object文件，由Makefiles遗留
DIR(.exe)        由go build产生
DIR.test(.exe)   由go test -c产生
MAINFILE(.exe)   由go build MAINFILE.go产生
```

#### go doc

强大的文档工具

- 查看相应package的文档，例如 `builtin` 包，那么执行 `go doc builtin`
- 查看某一个包里面的函数，那么执行 `godoc fmt Printf`，或者 `godoc -src fmt Printf`
- 查看golang.org的本地copy版本，执行 `godoc -http=:端口号`，如果设置了`GOPATH`，在pkg分类下，不但会列出标准包的文档，还会列出你本地GOPATH中所有项目的相关文档

#### go env

查看当前go的环境变量

#### go bug

#### go fix

用来修复以前老版本的代码到新版本，例如go1之前老版本的代码转化到go1

#### go fmt

格式化写好的代码文件

#### go generate

#### go get

用来动态获取远程代码包，目前支持的有 `BitBucket` 、 `GitHub` 、 `Google Code` 和 `Launchpad` 。
下载源码包的go工具会自动根据不同的域名调用不同的源码工具，对应关系如下：
```go
BitBucket (Mercurial Git)
GitHub (Git)
Google Code Project Hosting (Git, Mercurial, Subversion)
Launchpad (Bazaar)
```

为了 `go get` 能正常工作，必须确保安装了合适的源码管理工具，并同时把这些命令加入你的PATH中。`go get` 支持自定义域名的功能，具体参见 `go help remote` 。

#### go install

在内部实际上分成了两步操作：第一步是生成结果文件(可执行文件或者.a包)，第二步会把编译好的结果移到 `$GOPATH/pkg` 或者 `$GOPATH/bin`。

#### go list

列出当前全部安装的package

#### go run

编译并运行Go程序

#### go test

执行这个命令，会自动读取源码目录下面名为 `*_test.go` 的文件，生成并运行测试用的可执行文件。
输出的信息类似：
```go
ok   archive/tar   0.011s
FAIL archive/zip   0.022s
ok   compress/gzip 0.033s
...
```

#### go tool

#### go version

查看go当前的版本

#### go vet

`vet` 命令会帮开发人员检测代码的常见错误。
捕获的类型错误：

- `Printf` 类函数调用时，类型匹配错误的参数。 
- 定义常用的方法时，方法签名的错误。
- 错误的结构标签。
- 没有指定字段名的结构字面量。
