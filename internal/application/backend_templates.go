package application

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — Go Backend Templates
//  Стандартные модули для быстрой сборки бэкенда
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// GoTemplateAuth — модуль аутентификации (JWT + bcrypt)
const GoTemplateAuth = `package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("CHANGE_ME_IN_PRODUCTION")

type AuthHandler struct {
	db *DB
}

type LoginRequest struct {
	Email    string ` + "`json:\"email\"`" + `
	Password string ` + "`json:\"password\"`" + `
}

type RegisterRequest struct {
	Email    string ` + "`json:\"email\"`" + `
	Password string ` + "`json:\"password\"`" + `
	Name     string ` + "`json:\"name\"`" + `
}

type TokenResponse struct {
	AccessToken  string ` + "`json:\"access_token\"`" + `
	RefreshToken string ` + "`json:\"refresh_token\"`" + `
	ExpiresIn    int64  ` + "`json:\"expires_in\"`" + `
}

func NewAuthHandler(db *DB) *AuthHandler {
	return &AuthHandler{db: db}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, ` + "`" + `{"error":"invalid request"}` + "`" + `, http.StatusBadRequest)
		return
	}

	user, err := h.db.GetUserByEmail(req.Email)
	if err != nil {
		http.Error(w, ` + "`" + `{"error":"invalid credentials"}` + "`" + `, http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, ` + "`" + `{"error":"invalid credentials"}` + "`" + `, http.StatusUnauthorized)
		return
	}

	token, err := generateJWT(user.ID, user.Email)
	if err != nil {
		http.Error(w, ` + "`" + `{"error":"token generation failed"}` + "`" + `, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(token)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, ` + "`" + `{"error":"invalid request"}` + "`" + `, http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, ` + "`" + `{"error":"password hashing failed"}` + "`" + `, http.StatusInternalServerError)
		return
	}

	user := &User{
		Email:        req.Email,
		PasswordHash: string(hash),
		Name:         req.Name,
		CreatedAt:    time.Now(),
	}

	if err := h.db.CreateUser(user); err != nil {
		http.Error(w, ` + "`" + `{"error":"user creation failed"}` + "`" + `, http.StatusConflict)
		return
	}

	token, err := generateJWT(user.ID, user.Email)
	if err != nil {
		http.Error(w, ` + "`" + `{"error":"token generation failed"}` + "`" + `, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(token)
}

func generateJWT(userID, email string) (*TokenResponse, error) {
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return nil, err
	}

	refreshBytes := make([]byte, 32)
	rand.Read(refreshBytes)

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: hex.EncodeToString(refreshBytes),
		ExpiresIn:    86400,
	}, nil
}
`

// GoTemplateRouter — модуль маршрутизации API
const GoTemplateRouter = `package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Router struct {
	mux        *http.ServeMux
	middleware  []Middleware
}

type Middleware func(http.Handler) http.Handler

func NewRouter() *Router {
	return &Router{
		mux: http.NewServeMux(),
	}
}

func (r *Router) Use(mw Middleware) {
	r.middleware = append(r.middleware, mw)
}

func (r *Router) Handle(pattern string, handler http.HandlerFunc) {
	r.mux.HandleFunc(pattern, handler)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var handler http.Handler = r.mux
	for i := len(r.middleware) - 1; i >= 0; i-- {
		handler = r.middleware[i](handler)
	}
	handler.ServeHTTP(w, req)
}

// CORSMiddleware разрешает cross-origin запросы
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware логирует все запросы
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[%s] %s %s %v", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))
	})
}

// JWTAuthMiddleware проверяет JWT токен
func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Публичные эндпоинты
		publicPaths := []string{"/api/auth/login", "/api/auth/register", "/api/health"}
		for _, path := range publicPaths {
			if strings.HasPrefix(r.URL.Path, path) {
				next.ServeHTTP(w, r)
				return
			}
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, ` + "`" + `{"error":"unauthorized"}` + "`" + `, http.StatusUnauthorized)
			return
		}

		// Token validation logic here
		next.ServeHTTP(w, r)
	})
}

// RateLimitMiddleware ограничивает кол-во запросов
func RateLimitMiddleware(next http.Handler) http.Handler {
	// Простой rate limiter на основе IP
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// В production используйте Redis-based rate limiter
		next.ServeHTTP(w, r)
	})
}

func StartServer(router *Router) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Server starting on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
`

// GoTemplateDBConnect — модуль подключения к БД (PostgreSQL)
const GoTemplateDBConnect = `package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type DB struct {
	conn *sql.DB
}

type User struct {
	ID           string    ` + "`json:\"id\" db:\"id\"`" + `
	Email        string    ` + "`json:\"email\" db:\"email\"`" + `
	PasswordHash string    ` + "`json:\"-\" db:\"password_hash\"`" + `
	Name         string    ` + "`json:\"name\" db:\"name\"`" + `
	CreatedAt    time.Time ` + "`json:\"created_at\" db:\"created_at\"`" + `
	UpdatedAt    time.Time ` + "`json:\"updated_at\" db:\"updated_at\"`" + `
}

func NewDB() (*DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://localhost:5432/istok?sslmode=disable"
	}

	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✅ Database connected successfully")
	return &DB{conn: conn}, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) Migrate() error {
	migrations := []string{
		` + "`" + `CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL DEFAULT '',
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)` + "`" + `,
		` + "`" + `CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)` + "`" + `,
	}

	for _, m := range migrations {
		if _, err := db.conn.Exec(m); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	log.Println("✅ Database migrations completed")
	return nil
}

func (db *DB) GetUserByEmail(email string) (*User, error) {
	var user User
	err := db.conn.QueryRow(
		"SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE email = $1",
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *DB) CreateUser(user *User) error {
	return db.conn.QueryRow(
		"INSERT INTO users (email, password_hash, name) VALUES ($1, $2, $3) RETURNING id, created_at",
		user.Email, user.PasswordHash, user.Name,
	).Scan(&user.ID, &user.CreatedAt)
}
`

// GoTemplateMain — точка входа бэкенда
const GoTemplateMain = `package main

import (
	"log"
	"os"

	"backend/db"
	"backend/handlers"
)

func main() {
	// Database
	database, err := db.NewDB()
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}
	defer database.Close()

	if err := database.Migrate(); err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	// Handlers
	auth := handlers.NewAuthHandler(database)

	// Router
	router := NewRouter()
	router.Use(CORSMiddleware)
	router.Use(LoggingMiddleware)
	router.Use(JWTAuthMiddleware)
	router.Use(RateLimitMiddleware)

	// Routes
	router.Handle("POST /api/auth/login", auth.Login)
	router.Handle("POST /api/auth/register", auth.Register)
	router.Handle("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(` + "`" + `{"status":"ok"}` + "`" + `))
	})

	// Start
	StartServer(router)
}
`

// backendTemplateContext формирует контекст Go-шаблонов для Coder-агента
func backendTemplateContext(manifest *SystemManifest) string {
	if manifest == nil || manifest.Backend.Language != "go" {
		return ""
	}

	return `
GO BACKEND REFERENCE TEMPLATES (use these as starting point, adapt to the project):

=== AUTH MODULE (handlers/auth.go) ===
` + GoTemplateAuth + `

=== API ROUTER (main.go) ===
` + GoTemplateRouter + `

=== DB CONNECTION (db/connect.go) ===
` + GoTemplateDBConnect + `

=== MAIN ENTRY POINT (main.go) ===
` + GoTemplateMain + `

IMPORTANT: Adapt these templates to match the project specification. Add/remove endpoints as needed.
Do NOT copy blindly — customize for the specific use case.`
}
