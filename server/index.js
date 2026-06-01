import 'dotenv/config'
import express from 'express'
import cors from 'cors'
import session from 'express-session'
import multer from 'multer'
import path from 'path'
import bcrypt from 'bcryptjs'
import { fileURLToPath } from 'url'
import fs from 'fs'
import Database from 'better-sqlite3'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const app = express()
const PORT = process.env.PORT || 3001
const isProd = process.env.NODE_ENV === 'production'

const dataDir = path.join(__dirname, '..', 'data')
const uploadsDir = path.join(__dirname, '..', 'uploads')
fs.mkdirSync(dataDir, { recursive: true })
fs.mkdirSync(uploadsDir, { recursive: true })

const storage = multer.diskStorage({
  destination: (req, file, cb) => cb(null, uploadsDir),
  filename: (req, file, cb) => {
    const unique = Date.now() + '-' + Math.round(Math.random() * 1e9)
    cb(null, unique + path.extname(file.originalname))
  },
})

const upload = multer({
  storage,
  fileFilter: (req, file, cb) => {
    const allowed = /jpeg|jpg|png|gif|webp/
    const ext = allowed.test(path.extname(file.originalname).toLowerCase())
    const mime = allowed.test(file.mimetype)
    cb(ext && mime ? null : new Error('Only image files are allowed'), ext && mime)
  },
  limits: { fileSize: 5 * 1024 * 1024 },
})

// --- Database ---

const dbPath = path.join(dataDir, 'bakery.db')
const db = new Database(dbPath)
db.pragma('journal_mode = WAL')

db.exec(`
  CREATE TABLE IF NOT EXISTS products (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    price REAL NOT NULL,
    ingredients TEXT,
    image TEXT,
    category TEXT DEFAULT 'general',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
  );
  CREATE TABLE IF NOT EXISTS admins (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL
  );
`)

const pass = process.env.ADMIN_PASSWORD || 'admin123'
const adminCount = db.prepare('SELECT COUNT(*) as count FROM admins').get()
if (adminCount.count === 0) {
  const hash = bcrypt.hashSync(pass, 10)
  db.prepare('INSERT INTO admins (username, password) VALUES (?, ?)').run('admin', hash)
} else if (process.env.ADMIN_PASSWORD) {
  const hash = bcrypt.hashSync(pass, 10)
  db.prepare('UPDATE admins SET password = ? WHERE username = ?').run(hash, 'admin')
}

// --- Rate limiter ---

const loginAttempts = new Map()

function rateLimit(ip) {
  const now = Date.now()
  const window = 15 * 60 * 1000
  if (!loginAttempts.has(ip)) loginAttempts.set(ip, [])
  const times = loginAttempts.get(ip).filter(t => t > now - window)
  if (times.length >= 10) return false
  times.push(now)
  loginAttempts.set(ip, times)
  return true
}

setInterval(() => {
  const cutoff = Date.now() - 15 * 60 * 1000
  for (const [ip, times] of loginAttempts) {
    const filtered = times.filter(t => t > cutoff)
    filtered.length ? loginAttempts.set(ip, filtered) : loginAttempts.delete(ip)
  }
}, 60 * 1000)

// --- Middleware ---

app.use(cors({
  origin: isProd ? false : ['http://localhost:5173', 'http://localhost:4173'],
  credentials: true,
}))

app.use((req, res, next) => {
  res.setHeader('X-Content-Type-Options', 'nosniff')
  res.setHeader('X-Frame-Options', 'DENY')
  res.setHeader('Referrer-Policy', 'same-origin')
  if (isProd) res.setHeader('Strict-Transport-Security', 'max-age=31536000; includeSubDomains')
  next()
})

app.use(express.json())
app.use(session({
  secret: process.env.SESSION_SECRET || 'fallback-secret-change-me',
  resave: false,
  saveUninitialized: false,
  cookie: {
    secure: isProd,
    httpOnly: true,
    sameSite: 'lax',
    maxAge: 24 * 60 * 60 * 1000,
  },
}))

app.use('/uploads', express.static(uploadsDir))

const distDir = path.join(__dirname, '..', 'dist')
if (fs.existsSync(distDir)) {
  app.use(express.static(distDir))
}

// --- Auth middleware ---

function requireAuth(req, res, next) {
  if (!req.session || !req.session.adminId) {
    return res.status(401).json({ message: 'Unauthorized' })
  }
  const origin = req.headers['origin']
  if (origin && isProd) {
    const allowed = process.env.ALLOWED_ORIGIN
    if (allowed && origin !== allowed) {
      return res.status(403).json({ message: 'Forbidden' })
    }
  }
  next()
}

// --- Auth routes ---

app.post('/api/auth/login', (req, res) => {
  const ip = req.ip || req.connection.remoteAddress
  if (!rateLimit(ip)) {
    return res.status(429).json({ message: 'Too many login attempts. Try again later.' })
  }

  const { username, password } = req.body
  if (!username || !password) {
    return res.status(400).json({ message: 'Username and password are required' })
  }

  const admin = db.prepare('SELECT * FROM admins WHERE username = ?').get(username)
  if (!admin || !bcrypt.compareSync(password, admin.password)) {
    return res.status(401).json({ message: 'Invalid credentials' })
  }

  req.session.adminId = admin.id
  res.json({
    message: 'Login successful',
    admin: { id: admin.id, username: admin.username },
  })
})

app.post('/api/auth/logout', (req, res) => {
  req.session.destroy()
  res.json({ message: 'Logged out' })
})

app.get('/api/auth/me', (req, res) => {
  if (!req.session || !req.session.adminId) {
    return res.status(401).json({ message: 'Not authenticated' })
  }
  const admin = db.prepare('SELECT id, username FROM admins WHERE id = ?').get(req.session.adminId)
  if (!admin) return res.status(401).json({ message: 'Not authenticated' })
  res.json({ admin })
})

// --- Product routes ---

app.get('/api/products', (req, res) => {
  const products = db.prepare('SELECT * FROM products ORDER BY created_at DESC').all()
  res.json(products)
})

app.get('/api/products/:id', (req, res) => {
  const product = db.prepare('SELECT * FROM products WHERE id = ?').get(req.params.id)
  if (!product) return res.status(404).json({ message: 'Product not found' })
  res.json(product)
})

app.post('/api/products', requireAuth, (req, res) => {
  upload.single('image')(req, res, (err) => {
    if (err) return res.status(400).json({ message: err.message })
    const { name, description, price, ingredients, category } = req.body
    if (!name || !price) {
      return res.status(400).json({ message: 'Name and price are required' })
    }
    const image = req.file ? `/uploads/${req.file.filename}` : null
    const result = db.prepare(
      'INSERT INTO products (name, description, price, ingredients, image, category) VALUES (?, ?, ?, ?, ?, ?)'
    ).run(name, description || '', parseFloat(price), ingredients || '', image, category || 'general')
    const product = db.prepare('SELECT * FROM products WHERE id = ?').get(result.lastInsertRowid)
    res.status(201).json(product)
  })
})

app.put('/api/products/:id', requireAuth, (req, res) => {
  upload.single('image')(req, res, (err) => {
    if (err) return res.status(400).json({ message: err.message })
    const existing = db.prepare('SELECT * FROM products WHERE id = ?').get(req.params.id)
    if (!existing) return res.status(404).json({ message: 'Product not found' })
    const { name, description, price, ingredients, category } = req.body
    const image = req.file ? `/uploads/${req.file.filename}` : existing.image
    db.prepare(
      'UPDATE products SET name=?, description=?, price=?, ingredients=?, image=?, category=?, updated_at=CURRENT_TIMESTAMP WHERE id=?'
    ).run(
      name || existing.name,
      description !== undefined ? description : existing.description,
      price ? parseFloat(price) : existing.price,
      ingredients !== undefined ? ingredients : existing.ingredients,
      image,
      category || existing.category,
      req.params.id
    )
    const product = db.prepare('SELECT * FROM products WHERE id = ?').get(req.params.id)
    res.json(product)
  })
})

app.delete('/api/products/:id', requireAuth, (req, res) => {
  const existing = db.prepare('SELECT * FROM products WHERE id = ?').get(req.params.id)
  if (!existing) return res.status(404).json({ message: 'Product not found' })
  if (existing.image) {
    const p = path.join(__dirname, '..', existing.image)
    if (fs.existsSync(p)) fs.unlinkSync(p)
  }
  db.prepare('DELETE FROM products WHERE id = ?').run(req.params.id)
  res.json({ message: 'Product deleted' })
})

// --- Error handler ---

app.use((err, req, res, next) => {
  console.error('Error:', err.message)
  if (res.headersSent) return next(err)
  res.status(err.status || 500).json({ message: err.message || 'Internal server error' })
})

// --- SPA fallback ---

const distIndex = path.join(distDir, 'index.html')
if (fs.existsSync(distIndex)) {
  app.use((req, res) => {
    if (!req.path.startsWith('/api/') && !req.path.startsWith('/uploads/')) {
      res.sendFile(distIndex)
    }
  })
}

app.listen(PORT, () => {
  console.log(`Server running on http://localhost:${PORT}`)
})
