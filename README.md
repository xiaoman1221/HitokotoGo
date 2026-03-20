# HitokotoGo

一言（Hitokoto）句子服务 Go 语言实现，提供随机句子 API 接口。

## 项目简介

HitokotoGo 是一个基于 Go 语言开发的一言句子服务，从 [hitokoto.cn](https://hitokoto.cn) 获取句子数据，并提供 HTTP API 接口供前端调用。

## 功能特性

- 🚀 快速启动，基于 Go 1.26
- 📦 自动下载和缓存句子数据
- 🔄 支持按分类获取句子
- 🌐 提供 HTTP API 接口
- 🔧 支持环境变量配置

## 项目结构
```
HitokotoGo/ 
├── entity/ # 数据模型定义 
│ └── Sentences.go # 句子相关结构体 
├── sentences/ # 句子数据缓存目录 
├── wwwroot/ # 静态文件目录 
│ └── index.html # 首页 
├── .env # 环境变量配置 
├── go.mod # Go 模块依赖 
├── main.go # 主程序入口 
└── README.md # 项目文档
```
## 快速开始

### 环境要求

- Go 1.26 或更高版本

### 安装步骤

1. **克隆项目**
2. **安装依赖**

5. **访问服务**

打开浏览器访问：`http://localhost:8000`

## API 接口

### 获取随机句子

**接口地址：** `POST /v2`

**请求参数：**
- `key` (可选): 句子分类 key，默认为 `a`

**示例：**

### 句子分类

支持的分类 key 包括：
- `a` - 动画
- `b` - 漫画
- `c` - 游戏
- `d` - 文学
- `e` - 原创
- `f` - 来自网络
- `g` - 其他
- `h` - 影视
- `i` - 诗词
- `j` - 网易云
- `k` - 哲学
- `l` - 历史

## 配置说明

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| HOST | 服务器监听地址 | 0.0.0.0 |
| PORT | 服务器端口 | 8000 |
| BASE_URL | 句子数据源地址 | https://sentences-bundle.hitokoto.cn |

## 开发说明

### 数据结构

项目使用的主要数据结构定义在 `entity/Sentences.go` 中：

- `SentencesSimple`: 句子简化模型
- `SentencesVersion`: 句子版本信息
- `SentencesCategories`: 句子分类信息

### 自动更新

程序首次运行时会自动从 `sentences-bundle.hitokoto.cn` 下载句子数据到 `sentences/` 目录，后续运行会直接读取本地缓存。

## 技术栈

- **语言**: Go 1.26
- **Web 框架**: Go 标准库 net/http
- **依赖管理**: Go Modules
- **第三方库**: 
  - `github.com/joho/godotenv` - 环境变量管理

## 许可证

本项目遵循开源协议。

## 致谢

- 句子数据来源于 [hitokoto.cn](https://hitokoto.cn)

## 问题反馈

如有问题或建议，欢迎提交 Issue。
