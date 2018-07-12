---
layout: blog
title: 'vim配置'
date: 2016-05-20 00:12:21
categories: others
tags: vim
lead_text: 'vim配置'
---

### 配置
vim的配置在~/.vimrc文件中
```
"基础配置"
set nocompatible "去掉vi的一致性"
set number "显示行号"
set guioptions-=r "隐藏滚动条"
set guioptions-=L "隐藏滚动条"
set guioptions-=b "隐藏滚动条"
set showtabline=0 "隐藏顶部标签栏"
set guifont=Monaco:h13 "设置字体"
syntax on "开启语法高亮"
let g:solarized_termcolors=256 "solarized主题设置在终端下的设置"
set background=dark "设置背景色"
set nowrap "设置不折行"
set fileformat=unix "设置以unix的格式保存文件"
set cindent "设置C样式的缩进格式"
set tabstop=4 "设置table长度"
set shiftwidth=4 "设置table长度"
set showmatch "显示匹配的括号"
set scrolloff=5 "距离顶部和底部5行"    
set laststatus=2 "命令行为两行" 
set fenc=utf-8 "文件编码"
set backspace=2
set mouse=a "启用鼠标"
set selection=exclusive
set selectmode=mouse,key
set matchtime=5
set ignorecase "忽略大小写"
set incsearch
set hlsearch "高亮搜索项"
set noexpandtab "不允许扩展table"
set whichwrap+=<,>,h,l
set autoread
set cursorline "突出显示当前行"
set cursorcolumn  "突出显示当前列"
  
"插件安装"
filetype off
set rtp+=~/.vim/bundle/Vundle.vim
call vundle#begin()
Plugin 'VundleVim/Vundle.vim'
Plugin 'Valloric/YouCompleteMe'
Plugin 'Lokaltog/vim-powerline'
Plugin 'scrooloose/nerdtree'
Plugin 'Yggdroot/indentLine'
Plugin 'jiangmiao/auto-pairs'
Plugin 'tell-k/vim-autopep8'
Plugin 'scrooloose/nerdcommenter'
call vundle#end()
filetype plugin indent on

"NERDTree配置"
map <F2> :NERDTreeToggle<CR> "F2开启和关闭树"
let NERDTreeChDirMode=1
let NERDTreeShowBookmarks=1 "显示书签"
let NERDTreeIgnore=['\~$', '\.pyc$', '\.swp$'] "设置忽略文件类型"
let NERDTreeWinSize=25 "窗口大小"

"ycm配置"
let g:ycm_global_ycm_extra_conf = '~/.ycm_extra_conf.py' "默认配置文件路径"
let g:ycm_confirm_extra_conf=0 "打开vim时不再询问是否加载ycm_extra_conf.py配置"
set completeopt=longest,menu
let g:ycm_path_to_python_interpreter='/usr/bin/python' "python解释器路径"
let g:ycm_seed_identifiers_with_syntax=1 "是否开启语义补全"
let g:ycm_complete_in_comments=1 "是否在注释中也开启补全"
let g:ycm_collect_identifiers_from_comments_and_strings = 0 
let g:ycm_min_num_of_chars_for_completion=2 "开始补全的字符数"
let g:ycm_autoclose_preview_window_after_completion=1 "补全后自动关机预览窗口"
let g:ycm_cache_omnifunc=0 "禁止缓存匹配项,每次都重新生成匹配项"
let g:ycm_complete_in_strings = 1 "字符串中也开启补全"
autocmd InsertLeave * if pumvisible() == 0|pclose|endif "离开插入模式后自动关闭预览窗口"
inoremap <expr> <CR>       pumvisible() ? '<C-y>' : '<CR>' "回车即选中当前项"    

"ycm配置 上下左右键行为"
inoremap <expr> <Down>     pumvisible() ? '\<C-n>' : '\<Down>'
inoremap <expr> <Up>       pumvisible() ? '\<C-p>' : '\<Up>'
inoremap <expr> <PageDown> pumvisible() ? '\<PageDown>\<C-p>\<C-n>' : '\<PageDown>'
inoremap <expr> <PageUp>   pumvisible() ? '\<PageUp>\<C-p>\<C-n>' : '\<PageUp>'

"autopep8设置"
let g:indentLine_char='┆' "缩进指示线"
let g:indentLine_enabled = 1 "缩进指示线"
let g:autopep8_disable_show_diff=1

"nerdcommenter配置"
let mapleader=','
"注释/反注释"
map <F4> <leader>ci <CR>

```

### 插件管理
#### 安装
安装Vundle管理vim插件：
```bash
git clone https://github.com/VundleVim/Vundle.vim.git ~/.vim/bundle/Vundle.vim
```

#### 使用
重启vim并输入下面的命令:
```bash
:PluginInstall
```

#### 注意
YouCompleteMe不仅需要安装，还需要手动编译：
```
cd ~/.vim/bundle/YouCompleteMe
./install.py --clang-completer
```

### 参考
[把vim配置成顺手的python轻量级IDE（一）](https://www.jianshu.com/p/f0513d18742a)