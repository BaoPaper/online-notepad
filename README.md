<!-- PROJECT SHIELDS -->

[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![MIT License][license-shield]][license-url]

<!-- PROJECT LOGO -->
<br />

<p align="center">
  <a href="https://github.com/BaoPaper/online-notepad/">
    <img src="public/image/logo.png" alt="Logo" width="80" height="80">
  </a>

  <h3 align="center">在线记事本 / online-notepad</h3>
  <p align="center">
    📝 一个基于 Express.js 的简洁在线记事本
    <br />
    <br />
    <a href="https://github.com/BaoPaper/online-notepad/issues">报告 Bug</a>
    ·
    <a href="https://github.com/BaoPaper/online-notepad/issues">提出新特性</a>
  </p>
</p>

## 功能特性

- **无限笔记** - 支持创建任意数量的笔记
- **Markdown 预览** - 实时渲染（Marked.js）
- **附件上传** - 最大 10MB，自动生成 Markdown 链接
- **自动保存** - 输入即保存
- **JWT 认证** - 密码保护 + 记住我（更长的 JWT 有效期）
- **响应式界面** - 侧边栏导航，移动端可用

## 项目结构

```
online-notepad/
├── app.js              # Express 服务器与接口
├── package.json        # 依赖与脚本
├── notes/              # 笔记存储
│   ├── meta.json       # 元数据（id, title, createdAt）
│   └── {uuid}.txt      # 笔记内容
├── uploads/            # 附件存储
├── public/
│   ├── css/style.css
│   ├── js/marked.min.js
│   └── image/
└── views/
    ├── note.ejs        # 主界面
    └── login.ejs       # 登录页
```

## 安装与运行

1) 克隆仓库

```sh
git clone https://github.com/BaoPaper/online-notepad.git
cd online-notepad
```

2) 安装依赖

```sh
npm install
```

3) 设置环境变量

```sh
PASSWORD=your_password
JWT_SECRET=your_jwt_secret
```

4) 运行服务

```sh
npm start
```

## API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/notes` | 获取笔记列表 |
| POST | `/api/notes` | 新建笔记 |
| DELETE | `/api/notes/:id` | 删除笔记 |
| PUT | `/api/notes/:id/title` | 重命名笔记 |
| POST | `/save/:id` | 保存笔记内容 |
| POST | `/api/upload` | 上传附件 |

## 技术栈

- [Node.js](https://github.com/nodejs/node)
- [Express.js](https://expressjs.com)
- [jsonwebtoken](https://github.com/auth0/node-jsonwebtoken)
- [multer](https://github.com/expressjs/multer) - 文件上传
- [uuid](https://github.com/uuidjs/uuid) - 唯一 ID
- [Marked.js](https://github.com/markedjs/marked) - Markdown 解析

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `PASSWORD` | 登录密码 | `password` |
| `JWT_SECRET` | JWT 签名密钥 | 必填 |
| `JWT_SESSION_EXPIRES_IN` | 未勾选“记住我”的有效期 | `1d` |
| `JWT_REMEMBER_EXPIRES_IN` | 勾选“记住我”的有效期 | `30d` |
| `PORT` | 服务端口 | `3000` |

## License

MIT License - see [LICENSE.txt](https://github.com/BaoPaper/online-notepad/blob/master/LICENSE.txt)

<!-- links -->
[your-project-path]:BaoPaper/online-notepad
[contributors-shield]: https://img.shields.io/github/contributors/BaoPaper/online-notepad.svg?style=flat-square
[contributors-url]: https://github.com/BaoPaper/online-notepad/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/BaoPaper/online-notepad.svg?style=flat-square
[forks-url]: https://github.com/BaoPaper/online-notepad/network/members
[stars-shield]: https://img.shields.io/github/stars/BaoPaper/online-notepad.svg?style=flat-square
[stars-url]: https://github.com/BaoPaper/online-notepad/stargazers
[issues-shield]: https://img.shields.io/github/issues/BaoPaper/online-notepad.svg?style=flat-square
[issues-url]: https://github.com/BaoPaper/online-notepad/issues
[license-shield]: https://img.shields.io/github/license/BaoPaper/online-notepad.svg?style=flat-square
[license-url]: https://github.com/BaoPaper/online-notepad/blob/master/LICENSE.txt
