const express = require('express');
const jwt = require('jsonwebtoken');
const favicon = require('serve-favicon');
const bodyParser = require('body-parser');
const cookieParser = require('cookie-parser');
const path = require('path');
const fs = require('fs');
const multer = require('multer');
const { v4: uuidv4 } = require('uuid');

const app = express();
const port = 3000;

// 路径常量
const NOTES_DIR = path.join(__dirname, 'notes');
const UPLOADS_DIR = path.join(__dirname, 'uploads');
const META_FILE = path.join(NOTES_DIR, 'meta.json');

// 确保目录存在
if (!fs.existsSync(NOTES_DIR)) fs.mkdirSync(NOTES_DIR);
if (!fs.existsSync(UPLOADS_DIR)) fs.mkdirSync(UPLOADS_DIR);

// Meta 数据读写
function readMeta() {
  if (!fs.existsSync(META_FILE)) return { notes: [] };
  return JSON.parse(fs.readFileSync(META_FILE, 'utf8'));
}

function writeMeta(meta) {
  fs.writeFileSync(META_FILE, JSON.stringify(meta, null, 2));
}

// 迁移旧笔记 (1.txt ~ 8.txt)
function migrateOldNotes() {
  const meta = readMeta();
  if (meta.migrated) return;

  // 获取当前基础时间，避免循环中 Date.now() 产生微小差异导致顺序不可控
  const baseTime = Date.now();

  for (let i = 1; i <= 8; i++) {
    const oldPath = path.join(NOTES_DIR, `${i}.txt`);
    if (fs.existsSync(oldPath)) {
      const content = fs.readFileSync(oldPath, 'utf8');
      if (content.trim()) {
        const id = uuidv4();
        const newPath = path.join(NOTES_DIR, `${id}.txt`);
        fs.writeFileSync(newPath, content);
        meta.notes.push({
          id,
          title: `笔记 ${i}`,
          createdAt: baseTime - i * 1000
        });
      }
      fs.unlinkSync(oldPath);
    }
  }

  meta.migrated = true;
  writeMeta(meta);
}

// 启动时迁移
migrateOldNotes();

// 文件上传配置
const storage = multer.diskStorage({
  destination: UPLOADS_DIR,
  filename: (req, file, cb) => {
    const ext = path.extname(file.originalname);
    cb(null, `${uuidv4()}${ext}`);
  }
});
const upload = multer({
  storage,
  limits: { fileSize: 10 * 1024 * 1024 } // 10MB
});

// 中间件
app.use(bodyParser.urlencoded({ extended: true }));
app.use(bodyParser.json());
app.use(cookieParser());
app.use(express.static(path.join(__dirname, 'public')));
app.use('/uploads', express.static(UPLOADS_DIR));
app.use(favicon(path.join(__dirname, 'public', 'image', 'favicon.png')));


// 统一密码
const correctPassword = process.env.PASSWORD || 'password';
const JWT_SECRET = process.env.JWT_SECRET;
const JWT_SESSION_EXPIRES_IN = process.env.JWT_SESSION_EXPIRES_IN || '1d';
const JWT_REMEMBER_EXPIRES_IN = process.env.JWT_REMEMBER_EXPIRES_IN || '30d';

if (!JWT_SECRET) {
  throw new Error('JWT_SECRET is required');
}

function signAuthToken(rememberMe) {
  return jwt.sign(
    { role: 'user' },
    JWT_SECRET,
    { expiresIn: rememberMe ? JWT_REMEMBER_EXPIRES_IN : JWT_SESSION_EXPIRES_IN }
  );
}

function verifyAuthToken(token) {
  if (!token) return null;
  try {
    return jwt.verify(token, JWT_SECRET);
  } catch (err) {
    return null;
  }
}


// Cookie 自动登录中间件

// 认证检查
function requireAuth(req, res, next) {
  const payload = verifyAuthToken(req.cookies.auth_token);
  if (!payload) {
    return req.xhr || req.headers.accept?.includes('json')
      ? res.status(401).json({ success: false, message: 'Unauthorized' })
      : res.redirect('/login');
  }
  req.user = payload;
  next();
}

// 登录页面
app.get('/login', (req, res) => {
  const payload = verifyAuthToken(req.cookies.auth_token);
  if (payload) return res.redirect('/');
  res.render('login');
});

app.post('/login', (req, res) => {
  const { password, rememberMe } = req.body;
  if (password === correctPassword) {
    const remember = rememberMe === 'on';
    const token = signAuthToken(remember);
    const cookieOptions = {
      httpOnly: true,
      sameSite: 'strict',
      secure: false
    };
    if (remember) {
      cookieOptions.maxAge = 30 * 24 * 60 * 60 * 1000;
    }
    res.cookie('auth_token', token, cookieOptions);
    res.redirect('/');
  } else {
    res.send('密码错误，请重新输入。<br><a href="/login">返回登录页</a>');
  }
});

app.get('/logout', (req, res) => {
  res.clearCookie('auth_token');
  res.redirect('/login');
});

// 首页 - 重定向到第一个笔记或创建新笔记
app.get('/', requireAuth, (req, res) => {
  const meta = readMeta();
  if (meta.notes.length > 0) {
    // 按创建时间倒序，取最新的
    const sorted = [...meta.notes].sort((a, b) => b.createdAt - a.createdAt);
    res.redirect(`/note/${sorted[0].id}`);
  } else {
    // 创建第一个笔记
    const id = uuidv4();
    meta.notes.push({ id, title: '新建笔记', createdAt: Date.now() });
    writeMeta(meta);
    fs.writeFileSync(path.join(NOTES_DIR, `${id}.txt`), '');
    res.redirect(`/note/${id}`);
  }
});

// 笔记页面
app.get('/note/:id', requireAuth, (req, res) => {
  const { id } = req.params;
  const meta = readMeta();
  const note = meta.notes.find(n => n.id === id);
  if (!note) return res.status(404).send('笔记不存在<br><a href="/">返回首页</a>');

  const filePath = path.join(NOTES_DIR, `${id}.txt`);
  const content = fs.existsSync(filePath) ? fs.readFileSync(filePath, 'utf8') : '';
  const sortedNotes = [...meta.notes].sort((a, b) => b.createdAt - a.createdAt);
  res.render('note', { note: content, notes: sortedNotes, currentId: id, currentTitle: note.title });
});

// API: 获取单个笔记内容
app.get('/api/notes/:id/content', requireAuth, (req, res) => {
  const { id } = req.params;
  const meta = readMeta();
  const note = meta.notes.find(n => n.id === id);
  if (!note) return res.status(404).json({ success: false, message: 'Not found' });

  const filePath = path.join(NOTES_DIR, `${id}.txt`);
  const content = fs.existsSync(filePath) ? fs.readFileSync(filePath, 'utf8') : '';

  res.json({
    success: true,
    id: note.id,
    title: note.title,
    content,
    createdAt: note.createdAt
  });
});

// API: 获取笔记列表
app.get('/api/notes', requireAuth, (req, res) => {
  const meta = readMeta();
  const sorted = [...meta.notes].sort((a, b) => b.createdAt - a.createdAt);
  res.json(sorted);
});

// API: 创建笔记
app.post('/api/notes', requireAuth, (req, res) => {
  const meta = readMeta();
  const id = uuidv4();
  const note = { id, title: req.body.title || '新建笔记', createdAt: Date.now() };
  meta.notes.push(note);
  writeMeta(meta);
  fs.writeFileSync(path.join(NOTES_DIR, `${id}.txt`), '');
  res.json(note);
});

// 从内容中提取引用的附件文件名
function extractAttachments(content) {
  const regex = /\/uploads\/([a-f0-9-]+\.[a-z0-9]+)/gi;
  const matches = [];
  let match;
  while ((match = regex.exec(content)) !== null) {
    matches.push(match[1]);
  }
  return matches;
}

// 删除附件文件
function deleteAttachment(filename) {
  const filePath = path.join(UPLOADS_DIR, filename);
  if (fs.existsSync(filePath)) fs.unlinkSync(filePath);
}

// API: 删除笔记
app.delete('/api/notes/:id', requireAuth, (req, res) => {
  const { id } = req.params;
  const meta = readMeta();
  const note = meta.notes.find(n => n.id === id);
  if (!note) return res.status(404).json({ success: false, message: 'Not found' });

  // 删除笔记关联的附件
  if (note.attachments) {
    note.attachments.forEach(deleteAttachment);
  }

  meta.notes = meta.notes.filter(n => n.id !== id);
  writeMeta(meta);

  const filePath = path.join(NOTES_DIR, `${id}.txt`);
  if (fs.existsSync(filePath)) fs.unlinkSync(filePath);

  res.json({ success: true });
});

// API: 重命名笔记
app.put('/api/notes/:id/title', requireAuth, (req, res) => {
  const { id } = req.params;
  const { title } = req.body;
  const meta = readMeta();
  const note = meta.notes.find(n => n.id === id);
  if (!note) return res.status(404).json({ success: false, message: 'Not found' });

  note.title = title || '无标题';
  writeMeta(meta);
  res.json({ success: true, title: note.title });
});

// API: 保存笔记内容
app.post('/save/:id', requireAuth, (req, res) => {
  const { id } = req.params;
  const meta = readMeta();
  const note = meta.notes.find(n => n.id === id);
  if (!note) {
    return res.status(404).json({ success: false, message: 'Not found' });
  }

  const content = req.body.note || '';
  const filePath = path.join(NOTES_DIR, `${id}.txt`);
  fs.writeFileSync(filePath, content);

  // 清理不再引用的附件
  const currentAttachments = extractAttachments(content);
  const oldAttachments = note.attachments || [];
  oldAttachments.forEach(filename => {
    if (!currentAttachments.includes(filename)) {
      deleteAttachment(filename);
    }
  });
  note.attachments = currentAttachments;
  writeMeta(meta);

  res.json({ success: true });
});

// API: 上传附件
app.post('/api/upload', requireAuth, upload.single('file'), (req, res) => {
  if (!req.file) return res.status(400).json({ success: false, message: 'No file' });

  const { noteId } = req.body;
  const isImage = /\.(jpg|jpeg|png|gif|webp|svg|bmp)$/i.test(req.file.originalname);
  const url = `/uploads/${req.file.filename}`;
  const markdown = isImage
    ? `![${req.file.originalname}](${url})`
    : `[${req.file.originalname}](${url})`;

  // 关联附件到笔记
  if (noteId) {
    const meta = readMeta();
    const note = meta.notes.find(n => n.id === noteId);
    if (note) {
      note.attachments = note.attachments || [];
      note.attachments.push(req.file.filename);
      writeMeta(meta);
    }
  }

  res.json({
    success: true,
    url,
    filename: req.file.filename,
    originalname: req.file.originalname,
    markdown
  });
});

app.set('view engine', 'ejs');
app.set('views', path.join(__dirname, 'views'));

app.listen(port, () => {
  console.log(`Server is running on port ${port}`);
});
