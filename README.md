<!-- PROJECT SHIELDS -->

[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![MIT License][license-shield]][license-url]

<!-- PROJECT LOGO -->
<br />

<p align="center">
  <a href="https://github.com/BaoPaper/SimpleNote/">
    <img src="public/image/logo.png" alt="Logo" width="80" height="80">
  </a>

  <h3 align="center">SimpleNote / 简单笔记</h3>
  <p align="center">
    📝 一个基于 Go + Gin 的简洁 Markdown 在线笔记
    <br />
    <br />
    <a href="https://github.com/BaoPaper/SimpleNote/issues">报告 Bug</a>
    ·
    <a href="https://github.com/BaoPaper/SimpleNote/issues">提出新特性</a>
  </p>
</p>

## 功能特性

- **无限笔记** - 支持创建任意数量的笔记
- **Markdown 预览** - 前端实时渲染（Marked.js + DOMPurify）
- **附件上传** - 最大 10MB，自动生成 Markdown 链接
- **自动保存** - 输入即保存
- **JWT 认证** - 密码保护 + 记住我（更长的 JWT 有效期）
- **响应式界面** - 侧边栏导航，移动端可用

## 项目结构

```
SimpleNote/
├── .github/
│   └── workflows/
│       ├── docker-build.yml   # GHCR 镜像构建工作流
│       └── release.yml        # Tag 触发的 Release 工作流
├── internal/
│   └── notepad/
│       ├── app.go             # 应用装配与启动入口
│       ├── auth.go            # JWT 鉴权逻辑
│       ├── errors.go          # 错误页与错误响应
│       ├── handlers_auth.go   # 登录/登出/首页处理
│       ├── handlers_notes.go  # 笔记与上传接口
│       ├── models.go          # 常量与数据模型
│       ├── router.go          # Gin 路由注册
│       ├── store.go           # 笔记存储与迁移逻辑
│       ├── utils.go           # 通用辅助方法
│       ├── auth_test.go       # 鉴权测试
│       ├── router_test.go     # 路由测试
│       ├── store_test.go      # 存储测试
│       └── test_helpers_test.go # 测试辅助方法
├── public/
│   ├── css/
│   │   └── style.css
│   ├── image/
│   ├── js/
│   │   ├── marked.min.js
│   │   └── purify.min.js
│   └── svg/
├── views/
│   ├── note.ejs               # 主界面（Go Template）
│   ├── login.ejs              # 登录页
│   └── error.ejs              # 错误页
├── notes/                     # 笔记存储
│   ├── meta.json              # 元数据（id, title, createdAt）
│   └── {uuid}.txt             # 笔记内容
├── uploads/                   # 附件存储
├── .gitignore
├── docker-compose.yml         # 本地/服务器容器编排
├── Dockerfile                 # 生产镜像构建文件
├── go.mod                     # Go 模块定义
├── go.sum                     # Go 依赖锁定
├── LICENSE
├── main.go                    # 程序入口
└── README.md
```

## 安装与运行

1. 克隆仓库

```sh
git clone https://github.com/BaoPaper/SimpleNote.git
cd SimpleNote
```

要求：`Go 1.25+`

2. 下载依赖

```sh
go mod download
```

3. 设置环境变量

```sh
PASSWORD=your_password
JWT_SECRET=your_jwt_secret
```

4. 运行服务

```sh
go run .
```

5. 可选：构建二进制

```sh
go build -o simplenote .
```

## 无感迁移说明

- 保留原有路由结构：`/login`、`/logout`、`/note/:id`、`/api/*`、`/save/:id`
- 保留原有数据格式：继续使用 `notes/meta.json`、`notes/{uuid}.txt`、`uploads/`
- 保留原有认证约定：继续使用 `auth_token` Cookie、`PASSWORD`、`JWT_SECRET` 与相同的 JWT 过期配置
- 继续由前端 `marked.js` 渲染 Markdown，浏览器侧行为与原项目保持一致
- Docker Compose 仍将宿主机 `./data` 挂载到容器内 `/home/app/notes`，已有部署数据可直接复用

## API 接口

| 方法   | 路径                   | 说明         |
| ------ | ---------------------- | ------------ |
| GET    | `/api/notes`           | 获取笔记列表 |
| POST   | `/api/notes`           | 新建笔记     |
| DELETE | `/api/notes/:id`       | 删除笔记     |
| PUT    | `/api/notes/:id/title` | 重命名笔记   |
| POST   | `/save/:id`            | 保存笔记内容 |
| POST   | `/api/upload`          | 上传附件     |

## 技术栈

- [Go](https://go.dev)
- [Gin](https://gin-gonic.com)
- [golang-jwt/jwt](https://github.com/golang-jwt/jwt) - JWT 认证
- [google/uuid](https://github.com/google/uuid) - 唯一 ID
- [Marked.js](https://github.com/markedjs/marked) - Markdown 解析
- [DOMPurify](https://github.com/cure53/DOMPurify) - 预览内容净化

## 环境变量

| 变量                      | 说明                   | 默认值     |
| ------------------------- | ---------------------- | ---------- |
| `PASSWORD`                | 登录密码               | `password` |
| `JWT_SECRET`              | JWT 签名密钥           | 必填       |
| `JWT_SESSION_EXPIRES_IN`  | 未勾选“记住我”的有效期 | `1d`       |
| `JWT_REMEMBER_EXPIRES_IN` | 勾选“记住我”的有效期   | `30d`      |
| `PORT`                    | 服务端口               | `3000`     |

## 开发与验证

```sh
go test ./...
go build ./...
docker compose up --build
```

## Release 发布

- 推送 `v*` 格式的 Tag（如 `v1.0.0`）后，会自动执行 `.github/workflows/release.yml`
- 工作流会先运行 `go test ./...`，再交叉编译 Linux、macOS、Windows 的可执行包并创建 GitHub Release
- 同一个 Tag 也会触发 `.github/workflows/docker-build.yml`，向 GHCR 推送版本镜像标签
- Release 与 Docker 工作流都不需要额外配置仓库 Secrets 或 Variables，直接使用 GitHub 自动提供的 `GITHUB_TOKEN`
- 只有在实际运行应用时，才需要配置 `PASSWORD`、`JWT_SECRET` 等环境变量

## License

MIT License - see [LICENSE.txt](https://github.com/BaoPaper/SimpleNote/blob/main/LICENSE)

<!-- links -->

[your-project-path]: BaoPaper/SimpleNote
[contributors-shield]: https://img.shields.io/github/contributors/BaoPaper/SimpleNote.svg?style=flat-square
[contributors-url]: https://github.com/BaoPaper/SimpleNote/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/BaoPaper/SimpleNote.svg?style=flat-square
[forks-url]: https://github.com/BaoPaper/SimpleNote/network/members
[stars-shield]: https://img.shields.io/github/stars/BaoPaper/SimpleNote.svg?style=flat-square
[stars-url]: https://github.com/BaoPaper/SimpleNote/stargazers
[issues-shield]: https://img.shields.io/github/issues/BaoPaper/SimpleNote.svg?style=flat-square
[issues-url]: https://github.com/BaoPaper/SimpleNote/issues
[license-shield]: https://img.shields.io/github/license/BaoPaper/SimpleNote.svg?style=flat-square
[license-url]: https://github.com/BaoPaper/SimpleNote/blob/main/LICENSE
