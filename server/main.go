package main

import (
	"bufio"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"html"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	_ "modernc.org/sqlite"
	"golang.org/x/crypto/bcrypt"
)

func loadEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, "\"'")
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}

type rateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	limit    int
	window   time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		attempts: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	times := rl.attempts[key]
	var filtered []time.Time
	for _, t := range times {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) >= rl.limit {
		rl.attempts[key] = filtered
		return false
	}

	filtered = append(filtered, now)
	rl.attempts[key] = filtered
	return true
}

func (rl *rateLimiter) cleanup() {
	for {
		time.Sleep(rl.window)
		rl.mu.Lock()
		cutoff := time.Now().Add(-rl.window)
		for key, times := range rl.attempts {
			var filtered []time.Time
			for _, t := range times {
				if t.After(cutoff) {
					filtered = append(filtered, t)
				}
			}
			if len(filtered) == 0 {
				delete(rl.attempts, key)
			} else {
				rl.attempts[key] = filtered
			}
		}
		rl.mu.Unlock()
	}
}

var loginLimiter = newRateLimiter(10, 15*time.Minute)

var (
	db      *sql.DB
	store   *sessions.CookieStore
	uploads string
	dataDir string
)

type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Ingredients string  `json:"ingredients"`
	Image       *string `json:"image"`
	Category    string  `json:"category"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func main() {
	var err error

	execPath, err := os.Executable()
	rootDir := "/app"
	if err == nil {
		execDir := filepath.Dir(execPath)
		parent := filepath.Dir(execDir)
		if filepath.Base(execDir) == "server" {
			rootDir = parent
		} else {
			rootDir = execDir
		}
	}

	loadEnv(filepath.Join(rootDir, ".env"))

	dataDir = filepath.Join(rootDir, "data")
	uploads = filepath.Join(rootDir, "uploads")

	os.MkdirAll(dataDir, 0755)
	os.MkdirAll(uploads, 0755)

	dbPath := filepath.Join(dataDir, "bakery.db")
	db, err = sql.Open("sqlite", "file:"+dbPath+"?_pragma=journal_mode(WAL)")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	initDB()

	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		b := make([]byte, 32)
		rand.Read(b)
		secret = hex.EncodeToString(b)
	}
	isProd := os.Getenv("NODE_ENV") == "production"

	store = sessions.NewCookieStore([]byte(secret))
	store.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		Secure:   isProd,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	}

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "same-origin")
		if isProd {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	})

	if !isProd {
		r.Use(func(c *gin.Context) {
			c.Header("Access-Control-Allow-Origin", "http://localhost:5173")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Accept")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(204)
				return
			}
			c.Next()
		})
	}

	r.POST("/api/auth/login", handleLogin)
	r.POST("/api/auth/logout", handleLogout)
	r.GET("/api/auth/me", handleAuthMe)

	r.GET("/api/products", handleGetProducts)
	r.GET("/api/products/:id", handleGetProduct)
	r.POST("/api/products", requireAuth, handleCreateProduct)
	r.PUT("/api/products/:id", requireAuth, handleUpdateProduct)
	r.DELETE("/api/products/:id", requireAuth, handleDeleteProduct)

	r.Static("/uploads", uploads)

	distDir := filepath.Join(rootDir, "dist")
	if _, err := os.Stat(distDir); err == nil {
		r.Use(func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/api/") || strings.HasPrefix(c.Request.URL.Path, "/uploads/") {
				c.Next()
				return
			}
			filePath := filepath.Join(distDir, c.Request.URL.Path)
			if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
				c.File(filePath)
				c.Abort()
				return
			}
			c.File(filepath.Join(distDir, "index.html"))
			c.Abort()
		})
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}
	log.Printf("Server running on http://localhost:%s", port)
	r.Run(":" + port)
}

func initDB() {
	_, err := db.Exec(`
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
	if err != nil {
		log.Fatal("Failed to init database:", err)
	}

	pass := os.Getenv("ADMIN_PASSWORD")
	if pass == "" {
		pass = "admin123"
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM admins").Scan(&count)
	if count == 0 {
		hash, err := bcrypt.GenerateFromPassword([]byte(pass), 10)
		if err != nil {
			log.Fatal("Failed to hash password:", err)
		}
		_, err = db.Exec("INSERT INTO admins (username, password) VALUES (?, ?)", "admin", string(hash))
		if err != nil {
			log.Fatal("Failed to create admin:", err)
		}
	} else if os.Getenv("ADMIN_PASSWORD") != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(pass), 10)
		if err != nil {
			log.Fatal("Failed to hash password:", err)
		}
		db.Exec("UPDATE admins SET password = ? WHERE username = ?", string(hash), "admin")
	}
}

func requireAuth(c *gin.Context) {
	session, _ := store.Get(c.Request, "bakery-session")
	adminID, ok := session.Values["admin_id"]
	if !ok || adminID == nil {
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
		return
	}

	origin := c.GetHeader("Origin")
	if origin != "" {
		allowed := os.Getenv("ALLOWED_ORIGIN")
		if allowed == "" {
			allowed = "http://localhost:5173"
			if os.Getenv("NODE_ENV") == "production" {
				allowed = ""
			}
		}
		if allowed != "" && origin != allowed {
			c.AbortWithStatusJSON(403, gin.H{"message": "Forbidden"})
			return
		}
	}

	c.Set("admin_id", adminID)
	c.Next()
}

func handleLogin(c *gin.Context) {
	clientIP, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
	if !loginLimiter.allow(clientIP) {
		c.JSON(429, gin.H{"message": "Too many login attempts. Try again later."})
		return
	}

	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"message": "Invalid request"})
		return
	}
	if body.Username == "" || body.Password == "" {
		c.JSON(400, gin.H{"message": "Username and password are required"})
		return
	}

	var admin struct {
		ID       int
		Username string
		Password string
	}
	err := db.QueryRow("SELECT id, username, password FROM admins WHERE username = ?", body.Username).Scan(&admin.ID, &admin.Username, &admin.Password)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(body.Password)) != nil {
		c.JSON(401, gin.H{"message": "Invalid credentials"})
		return
	}

	session, _ := store.Get(c.Request, "bakery-session")
	session.Values["admin_id"] = admin.ID
	session.Options.MaxAge = 86400
	if err := session.Save(c.Request, c.Writer); err != nil {
		c.JSON(500, gin.H{"message": "Failed to create session"})
		return
	}

	c.JSON(200, gin.H{
		"message": "Login successful",
		"admin": gin.H{
			"id":       admin.ID,
			"username": admin.Username,
		},
	})
}

func handleLogout(c *gin.Context) {
	session, _ := store.Get(c.Request, "bakery-session")
	session.Values["admin_id"] = nil
	session.Save(c.Request, c.Writer)
	c.JSON(200, gin.H{"message": "Logged out"})
}

func handleAuthMe(c *gin.Context) {
	session, _ := store.Get(c.Request, "bakery-session")
	adminID, ok := session.Values["admin_id"]
	if !ok || adminID == nil {
		c.JSON(401, gin.H{"message": "Not authenticated"})
		return
	}

	var admin struct {
		ID       int
		Username string
	}
	err := db.QueryRow("SELECT id, username FROM admins WHERE id = ?", adminID).Scan(&admin.ID, &admin.Username)
	if err != nil {
		c.JSON(401, gin.H{"message": "Not authenticated"})
		return
	}

	c.JSON(200, gin.H{
		"admin": gin.H{
			"id":       admin.ID,
			"username": admin.Username,
		},
	})
}

func handleGetProducts(c *gin.Context) {
	rows, err := db.Query("SELECT id, name, description, price, ingredients, image, category, created_at, updated_at FROM products ORDER BY created_at DESC")
	if err != nil {
		c.JSON(500, gin.H{"message": "Database error"})
		return
	}
	defer rows.Close()

	products := []Product{}
	for rows.Next() {
		var p Product
		var desc, ingredients, image, category, createdAt, updatedAt sql.NullString
		if err := rows.Scan(&p.ID, &p.Name, &desc, &p.Price, &ingredients, &image, &category, &createdAt, &updatedAt); err != nil {
			continue
		}
		p.Description = desc.String
		p.Ingredients = ingredients.String
		if image.Valid {
			p.Image = &image.String
		}
		p.Category = category.String
		p.CreatedAt = createdAt.String
		p.UpdatedAt = updatedAt.String
		products = append(products, p)
	}

	c.JSON(200, products)
}

func handleGetProduct(c *gin.Context) {
	id := c.Param("id")
	var p Product
	var desc, ingredients, image, category, createdAt, updatedAt sql.NullString
	err := db.QueryRow("SELECT id, name, description, price, ingredients, image, category, created_at, updated_at FROM products WHERE id = ?", id).Scan(
		&p.ID, &p.Name, &desc, &p.Price, &ingredients, &image, &category, &createdAt, &updatedAt,
	)
	if err != nil {
		c.JSON(404, gin.H{"message": "Product not found"})
		return
	}
	p.Description = desc.String
	p.Ingredients = ingredients.String
	if image.Valid {
		p.Image = &image.String
	}
	p.Category = category.String
	p.CreatedAt = createdAt.String
	p.UpdatedAt = updatedAt.String
	c.JSON(200, p)
}

func handleCreateProduct(c *gin.Context) {
	name := strings.TrimSpace(c.PostForm("name"))
	priceStr := c.PostForm("price")
	if name == "" || priceStr == "" {
		c.JSON(400, gin.H{"message": "Name and price are required"})
		return
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid price"})
		return
	}

	var imagePath *string
	file, err := c.FormFile("image")
	if err == nil {
		ext := filepath.Ext(file.Filename)
		b := make([]byte, 16)
		rand.Read(b)
		filename := hex.EncodeToString(b) + ext
		dest := filepath.Join(uploads, filename)
		if err := c.SaveUploadedFile(file, dest); err == nil {
			path := "/uploads/" + filename
			imagePath = &path
		}
	}

	description := html.EscapeString(c.PostForm("description"))
	ingredients := html.EscapeString(c.PostForm("ingredients"))
	category := c.PostForm("category")
	if category == "" {
		category = "general"
	}

	var imageSQL *string
	if imagePath != nil {
		imageSQL = imagePath
	}

	result, err := db.Exec(
		"INSERT INTO products (name, description, price, ingredients, image, category) VALUES (?, ?, ?, ?, ?, ?)",
		name, description, price, ingredients, imageSQL, category,
	)
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to create product"})
		return
	}

	id, _ := result.LastInsertId()
	var p Product
	var desc, ing, img, cat, createdAt, updatedAt sql.NullString
	db.QueryRow("SELECT id, name, description, price, ingredients, image, category, created_at, updated_at FROM products WHERE id = ?", id).Scan(
		&p.ID, &p.Name, &desc, &p.Price, &ing, &img, &cat, &createdAt, &updatedAt,
	)
	p.Description = desc.String
	p.Ingredients = ing.String
	if img.Valid {
		p.Image = &img.String
	}
	p.Category = cat.String
	p.CreatedAt = createdAt.String
	p.UpdatedAt = updatedAt.String

	c.JSON(201, p)
}

func handleUpdateProduct(c *gin.Context) {
	id := c.Param("id")

	var existing Product
	var desc, ingredients, image, category, createdAt, updatedAt sql.NullString
	err := db.QueryRow("SELECT id, name, description, price, ingredients, image, category, created_at, updated_at FROM products WHERE id = ?", id).Scan(
		&existing.ID, &existing.Name, &desc, &existing.Price, &ingredients, &image, &category, &createdAt, &updatedAt,
	)
	if err != nil {
		c.JSON(404, gin.H{"message": "Product not found"})
		return
	}
	existing.Description = desc.String
	existing.Ingredients = ingredients.String
	if image.Valid {
		existing.Image = &image.String
	}
	existing.Category = category.String

	name := c.PostForm("name")
	if name == "" {
		name = existing.Name
	}
	priceStr := c.PostForm("price")
	price := existing.Price
	if priceStr != "" {
		price, err = strconv.ParseFloat(priceStr, 64)
		if err != nil {
			c.JSON(400, gin.H{"message": "Invalid price"})
			return
		}
	}
	description := c.PostForm("description")
	if description == "" {
		description = existing.Description
	}
	ingred := c.PostForm("ingredients")
	if ingred == "" {
		ingred = existing.Ingredients
	}
	cat := c.PostForm("category")
	if cat == "" {
		cat = existing.Category
	}

	var imagePath *string
	if existing.Image != nil {
		imagePath = existing.Image
	}
	file, err := c.FormFile("image")
	if err == nil {
		if existing.Image != nil {
			oldPath := filepath.Join(uploads, filepath.Base(*existing.Image))
			os.Remove(oldPath)
		}
		ext := filepath.Ext(file.Filename)
		b := make([]byte, 16)
		rand.Read(b)
		filename := hex.EncodeToString(b) + ext
		dest := filepath.Join(uploads, filename)
		if err := c.SaveUploadedFile(file, dest); err == nil {
			path := "/uploads/" + filename
			imagePath = &path
		}
	}

	_, err = db.Exec(
		"UPDATE products SET name=?, description=?, price=?, ingredients=?, image=?, category=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
		name, description, price, ingred, imagePath, cat, id,
	)
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to update product"})
		return
	}

	var p Product
	var d, ing, img, ca, cr, up sql.NullString
	db.QueryRow("SELECT id, name, description, price, ingredients, image, category, created_at, updated_at FROM products WHERE id = ?", id).Scan(
		&p.ID, &p.Name, &d, &p.Price, &ing, &img, &ca, &cr, &up,
	)
	p.Description = d.String
	p.Ingredients = ing.String
	if img.Valid {
		p.Image = &img.String
	}
	p.Category = ca.String
	p.CreatedAt = cr.String
	p.UpdatedAt = up.String

	c.JSON(200, p)
}

func handleDeleteProduct(c *gin.Context) {
	id := c.Param("id")

	var image sql.NullString
	err := db.QueryRow("SELECT image FROM products WHERE id = ?", id).Scan(&image)
	if err != nil {
		c.JSON(404, gin.H{"message": "Product not found"})
		return
	}

	if image.Valid && image.String != "" {
		oldPath := filepath.Join(uploads, filepath.Base(image.String))
		os.Remove(oldPath)
	}

	_, err = db.Exec("DELETE FROM products WHERE id = ?", id)
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to delete product"})
		return
	}

	c.JSON(200, gin.H{"message": "Product deleted"})
}

func init() {
	gin.SetMode(gin.ReleaseMode)
}
