# HitokotoGo

一个基于 Go 实现的轻量级一言（Hitokoto）服务，提供随机句子 API 与统计页面，支持启动时自动更新句子数据。

## 功能特性

- 纯 Go 标准库实现，无第三方 Web 框架
- 随机返回一言句子
- 支持按分类（动画、漫画、游戏、文学等 12 个分类）筛选
- 统计页面：展示各分类句子数量
- 启动时自动检查并更新句子数据
- 支持 Redis 可用性检测，不可用时回退至本地文件缓存
- 支持 Docker 部署（已配置中国大陆镜像加速）

## 项目结构

```
HitokotoGo/
├── entity/               # 数据模型定义
│   ├── Sentences.go      # 句子结构体
│   ├── Categories.go     # 分类结构体
│   └── Version.go        # 版本结构体
├── libs/                 # 工具库
│   ├── load.go           # 句子/分类数据加载
│   ├── store.go          # 内存句子缓存（含分类索引与 uuid 索引）
│   ├── random.go         # 随机数工具
│   ├── redis.go          # Redis 连接与按分类缓存
│   ├── httpget.go        # 带超时的 HTTP 下载
│   ├── version.go        # 本地/远程版本读取
│   ├── Sentences.go      # 句子下载更新逻辑
│   └── autoupdate.go     # 定时自动更新
├── frontend/             # 前端静态页面
│   ├── index.html        # 首页（随机展示句子）
│   └── stats.html        # 统计页面
├── data/                 # 句子数据缓存（自动生成）
│   ├── categories.json   # 分类定义
│   ├── version.json      # 版本信息
│   └── sentences/        # 各分类句子文件（a.json ~ l.json）
├── .env                  # 环境变量配置（需自行创建）
├── .env.exp              # 环境变量模板
├── go.mod / go.sum       # Go 模块依赖
├── Dockerfile            # Docker 构建文件（含国内镜像）
├── docker-compose.yml    # Docker Compose 编排
└── main.go               # 主程序入口
```

## 快速开始

### 环境要求

- Go 1.26 或更高版本

### 本地运行

```bash
# 1. 克隆项目
git clone https://git.1v.fit/xiaoman1221/HitokotoGo.git
cd HitokotoGo

# 2. 复制环境变量配置
cp .env.exp .env

# 3. 运行
go run .
```

打开浏览器访问 `http://localhost:8080`

### Docker 部署

```bash
# 使用 Docker Compose（推荐）
docker-compose up -d

# 或直接构建
docker build -t hitokotogo .
docker run -d -p 8080:8080 -v ./data:/app/data hitokotogo
```

#### 中国大陆用户特别说明

Dockerfile 已默认配置以下加速：

| 加速项 | 镜像源 |
|--------|--------|
| Go 模块下载 | `goproxy.cn` |
| Alpine 软件源 | `mirrors.ustc.edu.cn` |

Docker 镜像拉取（如 `redis:7-alpine`）需在 Docker daemon 中配置 registry-mirror：
- 阿里云加速器：登录 `https://cr.console.aliyun.com` 获取专属地址
- 中科大镜像源：`https://docker.mirrors.ustc.edu.cn`

## API 接口

### 随机获取句子

**接口地址：** `GET /v2`

**请求参数：**
- `c` (可选): 句子分类 key，不传则从全部分类中随机

**示例：**
```bash
# 随机获取一句
curl http://localhost:8080/v2

# 获取动画分类句子
curl http://localhost:8080/v2?c=a
```

**返回示例：**
```json
{
  "id": 1,
  "uuid": "xxx",
  "hitokoto": "世界那么大，我想去看看",
  "type": "a",
  "from": "网络",
  "from_who": null,
  "creator": "匿名",
  "creator_uid": 0,
  "reviewer": 0,
  "commit_from": "web",
  "created_at": "2020-01-01T00:00:00Z",
  "length": 13
}
```

### 获取统计数据

**接口地址：** `GET /stats/data`

**返回示例：**
```json
{
  "total": 1234,
  "categories": [
    { "key": "a", "name": "动画", "desc": "Anime - 动画", "count": 120 },
    { "key": "b", "name": "漫画", "desc": "Comic - 漫画", "count": 85 }
  ]
}
```

### 统计页面

浏览器访问 `http://localhost:8080/stats` 可查看可视化统计页面。

## 句子分类

| Key | 名称 | 说明 |
|-----|------|------|
| `a` | 动画 | Anime |
| `b` | 漫画 | Comic |
| `c` | 游戏 | Game |
| `d` | 文学 | Literature |
| `e` | 原创 | Original |
| `f` | 网络 | Internet |
| `g` | 其他 | Other |
| `h` | 影视 | Video |
| `i` | 诗词 | Poem |
| `j` | 网易云 | NetEase Music |
| `k` | 哲学 | Philosophy |
| `l` | 抖机灵 | Funny |

## 环境变量

| 变量名 | 说明 | 示例值 |
|--------|------|--------|
| `HOST` | 服务监听地址 | `0.0.0.0` |
| `PORT` | 服务监听端口 | `8080` |
| `REDIS_HOST` | Redis 主机地址 | `127.0.0.1` |
| `REDIS_PORT` | Redis 端口 | `6379` |
| `REDIS_PASSWORD` | Redis 密码 | `""` |
| `REDIS_DB` | Redis 数据库编号 | `0` |
| `SENTENCES_URL` | 句子数据源地址 | `https://sentences-bundle.hitokoto.cn` |
| `REFRESH_INTERVAL` | 首页自动刷新间隔（毫秒） | `5000` |
| `BACKGROUND_REFRESH` | 每次刷新是否更换背景图 | `false` |
| `AUTO_UPDATE` | 是否每小时自动检查句子包更新 | `true` |
| `BACKGROUND_API` | 背景图 API 地址 | `https://t.alcy.cc/pc` |

## 技术栈

- **语言**: Go 1.26
- **Web 框架**: Go 标准库 `net/http`
- **依赖管理**: Go Modules
- **主要依赖**:
  - `github.com/joho/godotenv` - 环境变量管理
  - `github.com/redis/go-redis/v9` - Redis 客户端

## 许可证

本项目遵循MIT协议。

## 致谢

- 句子数据来源于 [hitokoto.cn](https://hitokoto.cn)
