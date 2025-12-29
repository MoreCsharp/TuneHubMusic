# TuneHub Music 项目说明

## 项目概述

TuneHub Music 是一个基于 Go + Gin 框架的音乐搜索和下载应用。

- **语言**: Go 1.21+
- **框架**: Gin Web Framework
- **数据库**: SQLite (modernc.org/sqlite)
- **端口**: 8080
- **模块名**: yinyue

## 目录结构

```
TuneHubMusic/
├── main.go                 # 程序入口，初始化日志、配置、存储、路由
├── embed.go                # 嵌入静态资源 (static/, templates/)
├── go.mod / go.sum         # Go 模块依赖
├── ICON.PNG                # 应用图标
├── config/
│   └── config.go           # 配置管理
├── controllers/
│   ├── hello.go            # 测试接口
│   └── music.go            # 核心业务逻辑（搜索、下载、音乐库管理）
├── middleware/
│   └── cors.go             # CORS 跨域中间件
├── models/
│   └── response.go         # 响应模型
├── routes/
│   └── router.go           # 路由配置
├── storage/
│   └── storage.go          # 数据持久化（SQLite 数据库）
├── static/                 # 静态资源
│   ├── css/style.css       # 样式文件
│   ├── js/main.js          # 前端逻辑
│   └── favicon.png         # 网站图标
├── templates/
│   └── index.html          # 前端页面模板
├── data/                   # 数据目录
│   ├── app_data.db         # SQLite 数据库文件
│   └── app.log             # 应用日志
```

## 核心功能

### 1. 音乐搜索
- 调用外部 API: `https://music-dl.sayqz.com/api/`
- 支持多音源搜索

### 2. 音乐下载
- 异步下载任务队列
- 支持 MP3 (320k) 和 FLAC 格式
- **真实下载进度跟踪**（基于 Content-Length）

### 3. 音乐库管理
- 已下载歌曲管理（含专辑信息）
- 文件存在性验证
- 歌单导入功能

### 4. 数据持久化 (SQLite)
- 数据库文件: `./data/app_data.db`
- 使用纯 Go 实现的 SQLite 库 (modernc.org/sqlite)，无需 CGO

#### 数据库表结构

**settings** - 设置表
| 字段 | 类型 | 说明 |
|------|------|------|
| key | TEXT | 设置键 (PRIMARY KEY) |
| value | TEXT | 设置值 |

**library** - 音乐库表
| 字段 | 类型 | 说明 |
|------|------|------|
| id | TEXT | 歌曲ID |
| source | TEXT | 来源 |
| name | TEXT | 歌曲名 |
| artist | TEXT | 艺术家 |
| album | TEXT | 专辑 |
| filename | TEXT | 文件名 |
| path | TEXT | 文件路径 |
| time | TEXT | 下载时间 |

**playlists** - 歌单表
| 字段 | 类型 | 说明 |
|------|------|------|
| id | TEXT | 歌单ID |
| source | TEXT | 来源 |
| name | TEXT | 歌单名 |
| author | TEXT | 作者 |

**playlist_songs** - 歌单歌曲表
| 字段 | 类型 | 说明 |
|------|------|------|
| playlist_id | TEXT | 歌单ID |
| playlist_source | TEXT | 歌单来源 |
| song_id | TEXT | 歌曲ID |
| name | TEXT | 歌曲名 |
| artist | TEXT | 艺术家 |
| album | TEXT | 专辑 |
| types | TEXT | 可用音质 (JSON)

## API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/` | 主页 |
| GET | `/ping` | 健康检查 |
| GET | `/api/v1/search` | 搜索音乐 |
| GET | `/api/v1/url` | 获取音乐URL |
| GET | `/api/v1/download` | 下载音乐 (参数: source, id, name, artist, album, br) |
| GET | `/api/v1/downloads` | 下载任务列表 |
| GET | `/api/v1/library` | 获取音乐库 |
| POST | `/api/v1/library/refresh` | 刷新音乐库 |
| GET | `/api/v1/downloaded` | 检查是否已下载 |
| GET | `/api/v1/settings` | 获取设置 |
| POST | `/api/v1/settings` | 更新设置 |
| GET | `/api/v1/toplists` | 排行榜列表 |
| GET | `/api/v1/toplist` | 排行榜歌曲 |
| GET | `/api/v1/playlists` | 已导入歌单 |
| GET | `/api/v1/playlist/import` | 导入歌单 |
| DELETE | `/api/v1/playlist` | 删除歌单 |

## 编译运行

```bash
# 编译
go build -o tunehub-music .

# 运行
./tunehub-music
```

访问 http://localhost:8080 即可使用。
