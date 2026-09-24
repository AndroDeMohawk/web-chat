package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AndroDeMohawk/web-chat/internal/handler"
	"github.com/AndroDeMohawk/web-chat/internal/kafka"
	"github.com/AndroDeMohawk/web-chat/internal/repository"
	"github.com/AndroDeMohawk/web-chat/internal/service"
	"github.com/AndroDeMohawk/web-chat/internal/ws"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func main() {
	// 1. Инициализация Logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("error initializing zap logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 2. Инициализация PostgreSQL Pool
	cfg := repository.NewPostgresConfig(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	pool, err := repository.NewPostgresPool(ctx, *cfg)
	if err != nil {
		logger.Fatal("error initializing postgres pool", zap.Error(err))
	}
	defer pool.Close()

	repo := repository.NewRepository(pool)

	// 3. Сервисы и менеджеры
	authService := service.NewAuthService(repo)
	sessionManager := service.NewSessionManager()
	authMiddleware := handler.NewAuthMiddleware(sessionManager)

	// 4. WebSocket Hub
	hub := ws.NewHub()
	go hub.Run()

	// 5. Kafka Producer и Consumer
	brokers := []string{"localhost:9092"}
	kafkaTopic := "chat-messages"

	producer := kafka.NewProducer(brokers, kafkaTopic)
	defer producer.Close()

	consumer := kafka.NewConsumer(brokers, kafkaTopic, "chat-group")
	defer consumer.Close()

	// 6. Хэндлеры
	authHandler := handler.NewAuthHandler(authService, sessionManager)
	wsHandler := handler.NewWSHandler(hub, producer, repo)

	chatHandler, err := handler.NewChatHandler(repo, hub)
	if err != nil {
		logger.Fatal("failed to load templates", zap.Error(err))
	}

	// 7. Запуск Kafka Consumer -> Hub Broadcast
	consumer.Start(ctx, func(payload kafka.MessagePayload) {
		htmlBytes, err := chatHandler.RenderMessageHTML(payload)
		if err != nil {
			logger.Error("failed to render message html", zap.Error(err))
			return
		}
		hub.Broadcast(htmlBytes)
	}, logger)

	// 8. Настройка Chi Router
	r := chi.NewRouter()

	// Встроенные мидлвари Chi
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	// Вспомогательный опциональный middleware для чтения сессии
	// (чтобы главная страница понимала, авторизован юзер или нет)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(handler.CookieName)
			if err == nil {
				if sess, ok := sessionManager.GetSession(cookie.Value); ok {
					ctx := context.WithValue(r.Context(), handler.UserIDCtxKey, sess.UserID)
					ctx = context.WithValue(ctx, handler.UsernameCtxKey, sess.Username)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	})

	// Эндпоинты аутентификации и главной
	r.Get("/", chatHandler.RenderIndex)
	r.Post("/register", authHandler.Register)
	r.Post("/login", authHandler.Login)
	r.Get("/logout", authHandler.Logout)
	r.Post("/chat/clear", chatHandler.ClearChat)
	// Защищенный WebSocket эндпоинт
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.Middleware)
		r.Get("/ws", wsHandler.ServeWS)
	})

	// 9. Старт сервера
	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		logger.Info("starting server", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("error starting server", zap.Error(err))
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down server gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("error during server shutdown", zap.Error(err))
	}

	logger.Info("server stopped successfully")
}
