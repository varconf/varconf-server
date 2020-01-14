# varconf-server
> 基于go语言构建的分布式配置中心.

![](https://img.shields.io/badge/language-go-cccfff.svg)
[![Build Status](https://travis-ci.org/varconf/varconf-server.svg?branch=master)](https://travis-ci.org/varconf/varconf-server)
[![Go Report Card](https://goreportcard.com/badge/github.com/varconf/varconf-server)](https://goreportcard.com/report/github.com/varconf/varconf-server)

## 说明文档
- [English]()

## 安装部署
### 依赖环境
- 1、Mysql（需要通过varconf.sql初始化数据库）
- 2、Windows 或 Linux 或 MacOS

### 编译部署
#### 下载
```sh
git clone https://github.com/varconf/varconf-server.git
```
#### 编译
```sh
go build -mod=vendor
```
#### 配置
```
在config.json中写入数据库配置文件
```
### docker部署
```
TODO
```

### 操作命令
#### 启动 
>  1.普通模式
```sh
varconf -s start
```
>  2.daemon模式
```sh
varconf -s start -d
```
#### 停止
```sh
varconf -s stop
```

## 功能特性
- 1、简单易用: 接入灵活方便，一分钟上手；
- 2、轻量级部署: 部署简单，不依赖第三方服务，一分钟上手；
- 3、跨语言支持: 基于HTTP接口实现配置拉取（支持Long-Polling模式），易于多语言支持。

## 项目预览
`1.应用页面`

![image](https://github.com/varconf/varconf-doc/blob/master/images/app_list.png)

`2.配置页面`

![image](https://github.com/varconf/varconf-doc/blob/master/images/config_status.png)
