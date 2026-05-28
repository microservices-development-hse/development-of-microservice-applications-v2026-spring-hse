package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/microservices-development-hse/auth/internal/models"
	"github.com/microservices-development-hse/auth/internal/repository/postgres"
	"github.com/microservices-development-hse/auth/internal/service"
	gorm_postgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	userRepo  *postgres.UserRepository
	jwtSecret string
)

type loginResponse struct {
	Token     string `json:"token"`
	Email     string `json:"email"`
	ExpiresAt int64  `json:"expiresAt"`
}

func main() {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	jwtSecret = os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}

	var db *gorm.DB
	var err error
	for i := 0; i < 10; i++ {
		db, err = gorm.Open(gorm_postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("Failed to connect to database, retrying in 5s... (%d/10)", i+1)
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		log.Fatal("Could not connect to database after retries")
	}

	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}

	userRepo = postgres.NewUserRepository(db)

	mux := http.NewServeMux()
	mux.HandleFunc("/register", handleRegister)
	mux.HandleFunc("/login", handleLogin)
	mux.HandleFunc("/health", handleHealth)

	log.Println("Auth service starting on :8083...")
	log.Fatal(http.ListenAndServe(":8083", withCORS(mux)))
}

func withCORS(next http.Handler) http.Handler {
	allowedOrigins := map[string]struct{}{
		"http://localhost:4200": {},
		"http://localhost:3000": {},
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if _, ok := allowedOrigins[origin]; ok {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}

		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	input.Email = strings.TrimSpace(input.Email)
	input.Password = strings.TrimSpace(input.Password)

	if input.Email == "" || input.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	hashedPassword, err := service.HashPassword(input.Password)
	if err != nil {
		http.Error(w, "Error processing password", http.StatusInternalServerError)
		return
	}

	user := models.User{
		Email:        input.Email,
		PasswordHash: hashedPassword,
	}

	if err := userRepo.CreateUser(&user); err != nil {
		http.Error(w, "User already exists or DB error", http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "User created successfully"})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	input.Email = strings.TrimSpace(input.Email)

	user, err := userRepo.GetByEmail(input.Email)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	if !service.CheckPasswordHash(input.Password, user.PasswordHash) {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	token, expiresAt, err := service.GenerateToken(user.ID, jwtSecret)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(loginResponse{
		Token:     token,
		Email:     user.Email,
		ExpiresAt: expiresAt.Unix(),
	})
}
