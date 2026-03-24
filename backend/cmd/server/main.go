package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"simplestock/internal/config"
	"simplestock/internal/domain"
	"simplestock/internal/handler"
	"simplestock/internal/middleware"
	"simplestock/internal/repository"
)

func main() {
	cfg := config.Load()

	// Database
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Не удалось подключиться к БД: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("БД недоступна: %v", err)
	}
	log.Println("Подключение к БД установлено")

	// Migrations
	runMigrations(cfg.DatabaseURL)

	// Seed admin user
	seedAdmin(pool)

	// Session store
	middleware.InitSessionStore(cfg.SessionKey)

	// Repos
	userRepo := repository.NewUserRepo(pool)
	categoryRepo := repository.NewCategoryRepo(pool)
	productRepo := repository.NewProductRepo(pool)
	docRepo := repository.NewDocumentRepo(pool)
	invRepo := repository.NewInventoryRepo(pool)

	// Handlers
	authH := handler.NewAuthHandler(userRepo)
	categoryH := handler.NewCategoryHandler(categoryRepo)
	productH := handler.NewProductHandler(productRepo)
	docH := handler.NewDocumentHandler(docRepo)
	invH := handler.NewInventoryHandler(invRepo)
	dashH := handler.NewDashboardHandler(pool)
	exportH := handler.NewExportHandler(pool)

	// Router
	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	}))

	// Public routes
	r.Post("/api/auth/login", authH.Login)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth)

		r.Post("/api/auth/logout", authH.Logout)
		r.Get("/api/auth/me", authH.Me)

		// Categories
		r.Get("/api/categories", categoryH.List)
		r.Post("/api/categories", categoryH.Create)

		// Products
		r.Get("/api/products", productH.List)
		r.Get("/api/products/low-stock", productH.LowStock)
		r.Get("/api/products/{id}", productH.GetByID)
		r.Post("/api/products", productH.Create)
		r.Put("/api/products/{id}", productH.Update)
		r.Delete("/api/products/{id}", productH.Delete)

		// Documents
		r.Get("/api/documents", docH.List)
		r.Get("/api/documents/{id}", docH.GetByID)
		r.Post("/api/documents", docH.Create)
		r.Post("/api/documents/{id}/post", docH.Post)
		r.Delete("/api/documents/{id}", docH.Delete)

		// Movements
		r.Get("/api/movements", docH.Movements)

		// Inventory
		r.Get("/api/inventories", invH.List)
		r.Get("/api/inventories/{id}", invH.GetByID)
		r.Post("/api/inventories", invH.Create)
		r.Post("/api/inventories/{id}/items", invH.AddItem)
		r.Post("/api/inventories/{id}/complete", invH.Complete)

		// Dashboard & Reports
		r.Get("/api/dashboard/summary", dashH.Summary)
		r.Get("/api/reports/stock", dashH.StockReport)
		r.Get("/api/reports/turnover", dashH.TurnoverReport)
		r.Get("/api/reports/export/excel", exportH.ExportExcel)
	})

	// Server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Printf("Сервер запущен на порту %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка сервера: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Остановка сервера...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	log.Println("Сервер остановлен")
}

func runMigrations(dbURL string) {
	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		log.Printf("Миграции: не удалось инициализировать: %v", err)
		return
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Printf("Миграции: ошибка: %v", err)
		return
	}
	log.Println("Миграции применены")
}

func seedAdmin(pool *pgxpool.Pool) {
	ctx := context.Background()
	var exists bool
	pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE username = 'admin')").Scan(&exists)
	if exists {
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	admin := &domain.User{
		Username: "admin",
		Password: string(hash),
		FullName: "Администратор",
		Role:     "admin",
	}

	repo := repository.NewUserRepo(pool)
	if err := repo.Create(ctx, admin); err != nil {
		log.Printf("Seed admin: %v", err)
		return
	}
	log.Println("Создан пользователь admin/admin")
}
