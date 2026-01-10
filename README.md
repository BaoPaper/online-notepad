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
    🗒 一个基于 Express.js 的简洁在线记事本 / A minimalist online notepad based on Express.js
    <br />
    <br />
    <a href="https://github.com/BaoPaper/online-notepad/issues">报告Bug</a>
    ·
    <a href="https://github.com/BaoPaper/online-notepad/issues">提出新特性</a>
  </p>

</p>

## Features

- **Unlimited Notes** - Create as many notes as you need
- **Markdown Preview** - Real-time markdown rendering with Marked.js
- **File Attachments** - Upload files (max 10MB) and insert as markdown links
- **Auto Save** - Content saves automatically as you type
- **Session Auth** - Password protection with "Remember Me" cookie
- **Responsive UI** - Sidebar navigation with mobile support

## Project Structure

```
online-notepad/
├── app.js              # Express server with all routes and APIs
├── package.json        # Dependencies
├── notes/              # Note storage
│   ├── meta.json       # Note metadata (id, title, createdAt)
│   └── {uuid}.txt      # Note content files
├── uploads/            # Uploaded attachments
├── public/
│   ├── css/style.css   # Styles
│   ├── js/marked.min.js
│   └── image/
└── views/
    ├── note.ejs        # Main notepad view
    └── login.ejs       # Login page
```

## Installation

1. Clone the repo

```sh
git clone https://github.com/BaoPaper/online-notepad.git
cd online-notepad
```

2. Install NPM packages
```sh
npm install
```

3. Run the server
```sh
node app.js
```

4. (Optional) Set custom password
```sh
PASSWORD=your_password node app.js
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/notes` | List all notes |
| POST | `/api/notes` | Create new note |
| DELETE | `/api/notes/:id` | Delete note |
| PUT | `/api/notes/:id/title` | Rename note |
| POST | `/save/:id` | Save note content |
| POST | `/api/upload` | Upload attachment |

## Tech Stack

- [Node.js](https://github.com/nodejs/node)
- [Express.js](https://expressjs.com)
- [express-session](https://github.com/expressjs/session)
- [multer](https://github.com/expressjs/multer) - File uploads
- [uuid](https://github.com/uuidjs/uuid) - Unique IDs
- [Marked.js](https://github.com/markedjs/marked) - Markdown parsing

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PASSWORD` | Login password | `password` |
| `PORT` | Server port | `3000` |

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
[issues-url]: https://img.shields.io/github/issues/BaoPaper/online-notepad.svg
[license-shield]: https://img.shields.io/github/license/BaoPaper/online-notepad.svg?style=flat-square
[license-url]: https://github.com/BaoPaper/online-notepad/blob/master/LICENSE.txt
