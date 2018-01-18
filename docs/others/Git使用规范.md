---
layout: blog
title: 'Git使用规范'
date: 2017-04-25 01:03:16
categories: others
tags: git
lead_text: 'Git实践及规范'
---
## 背景
Ruby on Rails作者汉森说，灵活性被过分高估——约束才是解放。
无规矩不成方圆。世事向来如此，不在一定规则之内，十之八九不能成事。
假使一个团队内部没有约束，形成一套做事的规范，必定走向混乱。

## 协作流程
目前比较广泛使用的协作流程有三种，Git Flow/Github Flow/Gitlab Flow，所处团队使用的是Git Flow/Github Flow，就聊聊这个。
那么，当谈论协作流程时，我们在应该讨论些什么？可以参考这篇文章，[Gitflow Workflow](https://www.atlassian.com/git/tutorials/comparing-workflows/gitflow-workflow)。

## 操作
### 分支
#### 主分支
##### 主分支（Master）
代码库有且只有一个主分支，用来发布正式版本。

#### 开发分支
##### 开发分支（Develop）
- 开发分支用来生成代码的最新开发中的版本
- 需要对外发布时，则在Master分支上，对Develop分支进行合并

#### 临时分支
##### 功能分支（feature）
- 为了开发特定功能从Develop分支checkout下来的，开发完成后再并入Develop
- 命名为feature/x
- 合并成功后删除此分支或可保留一周再删除

##### 预发布分支（release）
- 预发布分支是从Develop分支checkout下来，在正式发布版本之前，需要对预发布的版本进行一个全面测试
- 命令为release/x
- 预发布结束后，确认没问题后需要分别合并到Master分支和Develop分支

##### 热修复分支（hotfix）
- 线上出现紧急bug需要发布一个分支修复，从Master分支checkout出来
- 命名为hotfix/x
- 修复结束后再合并进Master和Develop分支

##### 修复分支（fix）
- 软件发布后难免有bug，需要发布一个分支进行bug修复，从Develop分支checkout出来
- 命名为fix/x
- 修复结束后再合并进Develop分支

> P.S. Master分支和Develop分支都是在主仓库上，feature/release/fix/hotfix分支是开发者从主仓库fork下来的仓库上创建的，分支命名可以参考[语义化版本 2.0.0](https://semver.org/lang/zh-CN/)。

### 提交
- 不要一次提交就推送，可多次提交再推送
- 提交合并时的粒度是一个功能点/bug fix
- 第一行是不超过50个字的提要，空一行罗列出改动原因、主要变动、以及需要注意的问题，最后一行提供对应的记录网址，包括bug/功能点
```bash
Present-tense summary under 50 characters

* More information about commit (under 72 characters).
* More information about commit (under 72 characters).
......

http://taiga.bu6.io/project/p_c-appxiang-guan-ye-wu/us/274
```

### 推送

### 拉取
### 合并

## 开发流程
## 其他补充
[25 Tips for Intermediate Git Users](https://www.andyjeffries.co.uk/25-tips-for-intermediate-git-users/)
## 参考
- [Git分支管理策略](http://www.ruanyifeng.com/blog/2012/07/git.html)
- [团队中的 Git 实践](https://ourai.ws/posts/working-with-git-in-team/?hmsr=toutiao.io&utm_medium=toutiao.io&utm_source=toutiao.io)

[参考](http://www.codeceo.com/article/git-standard.html)
[参考](http://developer.51cto.com/art/201512/502921.htm)
[参考](https://www.jianshu.com/p/08eef2108647)
[参考](https://zhuanlan.zhihu.com/p/23478654)
[参考](https://gist.github.com/belm/6989341)
[参考](https://ourai.ws/posts/working-with-git-in-team/?hmsr=toutiao.io&utm_medium=toutiao.io&utm_source=toutiao.io)
[参考](https://semver.org/lang/zh-CN/)
[参考](https://mp.weixin.qq.com/s?__biz=MzAwNDYwNzU2MQ==&mid=401622986&idx=1&sn=470717939914b956ac372667ed23863c&scene=2&srcid=0114ZcTNyAMH8CLwTKlj6CTN&from=timeline&isappinstalled=0#wechat_redirect)
[参考](http://www.codeceo.com/article/git-standard.html)
[参考](https://www.jianshu.com/p/3c68937586c8?utm_campaign=maleskine&utm_content=note&utm_medium=seo_notes&utm_source=recommendation)
[参考](http://developer.51cto.com/art/201512/502921.htm)

