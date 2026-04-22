# HitokotoGo

一个基于 Go 实现的简易一言（Hitokoto）服务，提供随机句子读取接口，并支持启动时自动检查、下载和更新句子数据。

## 项目简介

HitokotoGo 是一个使用 Go 标准库构建的轻量级 HTTP 服务。程序启动后会：

- 加载 `.env` 配置
- 检查 Redis 连接状态
- 检查本地句子数据是否存在或是否需要更新
- 自动从远程句子源下载或刷新数据
- 提供首页和随机句子 API

该项目适合用于：

- 个人练手项目
- 简单 API 服务部署
- 一言/随机文案接口场景
- Docker / Docker Compose 部署示例

## 功能特性

- 使用 Go 编写，依赖简单
- 基于标准库 `net/http` 提供 HTTP 接口
- 启动时自动检测并更新句子数据
- 支持按分类随机返回句子
- 支持 Redis 可用性检测
- Redis 不可用时可回退到文件数据读取
- 支持 Docker 部署
- 支持 Docker Compose 编排 Go 服务与 Redis

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

## 运行环境

- Go 1.26
- 可选：Redis
- 可选：Docker / Docker Compose

## 配置说明

程序启动时会优先读取项目根目录下的 `.env` 文件。

如果 `.env` 不存在，程序会尝试自动创建。但建议你手动准备好 `.env`，避免首次启动时因默认配置异常影响运行。

### 环境变量

| 变量名 | 说明 | 示例值 |
|--------|------|--------|
| `HOST` | 服务监听地址 | `0.0.0.0` |
| `PORT` | 服务监听端口 | `8080` |
| `REDIS_HOST` | Redis 主机地址 | `127.0.0.1` 或 `redis` |
| `REDIS_PORT` | Redis 端口 | `6379` |
| `REDIS_PASSWORD` | Redis 密码 | `""` |
| `REDIS_DB` | Redis 数据库编号 | `0` |
| `SENTENCES_URL` | 句子数据源地址 | `https://sentences-bundle.hitokoto.cn` |

### `.env` 示例

## 本地运行

### 1. 克隆项目

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
