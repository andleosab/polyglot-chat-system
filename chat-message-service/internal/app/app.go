// Wires up the application components and starts the HTTP server
package app

import (
	"chat-message-service/internal/config"
	sqlcdb "chat-message-service/internal/db/sqlc"
	"chat-message-service/internal/kafka"
	"chat-message-service/internal/repository"
	"chat-message-service/internal/rest"
	"chat-message-service/internal/rest/handler"
	"chat-message-service/internal/service"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Server   *http.Server
	pool     *pgxpool.Pool
	consumer *kafka.Consumer
	producer *kafka.Producer
}

func NewApp() *App {
	return &App{}
}

func (a *App) Start() error {

	// ---- Create a context that cancels on SIGINT/SIGTERM ----
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ---- Load config ----
	config, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		return err
	}

	// ---- Startup context for DB/Kafka initialization ----
	startupCtx, startupCancel := context.WithTimeout(ctx, 5*time.Second)
	defer startupCancel()

	pool, err := NewPool(startupCtx, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	a.pool = pool
	log.Println("DB pool initialized")

	// ---- SQLC queries ----
	q := sqlcdb.New(pool)

	// ---- Repositories ----
	convRepo := repository.NewConversationRepoSQLC(q, pool)
	partRepo := repository.NewParticipantRepoSQLC(q)
	msgRepo := repository.NewMessageRepoSQLC(q)

	// ---- Start consumer ----
	consumer, err := kafka.NewConsumer(startupCtx, config, msgRepo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "kafka consumer error: %v\n", err)
		return err
	}
	a.consumer = consumer
	log.Println("Kafka consumer initialized")

	// wire up kafka producer if needed for testing or other events
	producer, err := kafka.NewProducer(startupCtx, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "kafka producer error: %v\n", err)
		return err
	}
	a.producer = producer
	log.Println("Kafka producer initialized")

	// ---- Services ----
	convSvc := service.NewConversationService(convRepo)
	partSvc := service.NewParticipantService(partRepo, producer)
	msgSvc := service.NewMessageService(msgRepo)

	// ---- Handlers ----
	convHandler := handler.NewConversationHandler(convSvc)
	partHandler := handler.NewParticipantHandler(partSvc)
	msgHandler := handler.NewMessageHandler(msgSvc)

	// ---- Router ----
	rootRouter := rest.NewRootRouter("/messaging", &config.JWTConfig)
	rootRouter.BuildRouter(convHandler, partHandler, msgHandler)

	PrintRoutes(rootRouter.GetRouter())

	// ---- HTTP Server ----
	a.Server = &http.Server{
		Addr:         config.HttpPort,
		Handler:      rootRouter.GetRouter(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	// ---- Start server in goroutine ----
	go func() {
		log.Printf("Server running on %s\n", config.HttpPort)
		if err := a.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// --- Start Consumer ---
	// ---- Start Kafka consumer in goroutine (blocking inside Start) ----
	go func() {
		log.Printf("Kafka consumer starting on %s\n", config.SeedBrokers)
		a.consumer.Start(ctx) // now Start() is blocking
		log.Println("Kafka consumer stopped")
	}()

	// ---- Wait until context is done (signal received) ----
	<-ctx.Done()
	log.Println("Shutdown signal received")

	// ensure DB and other resources close when Start exits
	defer func() {
		a.pool.Close()
		a.consumer.Stop()
		a.producer.Close()
		log.Println("Application shutdown complete")
	}()

	// ---- Graceful shutdown context with timeout ----
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// ---- Stop HTTP server gracefully ----
	if err := a.Server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}
	log.Println("Server gracefully stopped")

	return nil
}

func PrintRoutes(r chi.Routes) {
	log.Println("==== ROUTES ====")
	_ = chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		log.Printf("%-6s %s\n", method, route)
		return nil
	})
	log.Println("================")
}
