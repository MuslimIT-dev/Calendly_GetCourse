package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
	kafkago "github.com/segmentio/kafka-go"

	migrationfiles "github.com/MuslimIT-dev/Calendly_GetCourse/backend/db/migrations"
	authv1connect "github.com/MuslimIT-dev/Calendly_GetCourse/backend/gen/go/auth/v1/authv1connect"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/infrastructure/broker"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/infrastructure/cache"
	postgresdb "github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/repository/db"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/repository/postgres"
	connecttransport "github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/transport/connect"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/transport/connect/middleware"
	authuc "github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/usecase/auth"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/pkg/hasher"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/pkg/jwt"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	dbURL := mustEnv("DATABASE_URL")
	redisURL := mustEnv("REDIS_URL")
	kafkaBroker := mustEnv("KAFKA_BROKER")
	jwtSecret := mustEnv("JWT_SECRET")

	if err := runMigrations(dbURL); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pool.Close()

	redisOpt, err := goredis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("redis parse: %v", err)
	}
	rdb := goredis.NewClient(redisOpt)
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis ping: %v", err)
	}

	kafkaWriter := &kafkago.Writer{
		Addr:     kafkago.TCP(kafkaBroker),
		Balancer: &kafkago.LeastBytes{},
	}
	defer kafkaWriter.Close()

	queries := postgresdb.New(pool)

	userRepo := postgres.NewUserRepo(queries)
	sessionCache := cache.NewRedisCache[authuc.SessionValue](rdb)
	verifyTokenCache := cache.NewRedisCache[authuc.VerifyEmailValue](rdb)
	publisher := broker.NewKafkaPublisher[authuc.UserRegisteredEvent](kafkaWriter)

	passwordHasher := hasher.NewBcryptHasher()
	tokenService := jwt.NewJWTService(jwtSecret, "calendly-clone")

	registerUC := authuc.NewRegisterUseCase(authuc.Deps{
		Users:          userRepo,
		Sessions:       sessionCache,
		VerifyTokens:   verifyTokenCache,
		Hasher:         passwordHasher,
		Tokens:         tokenService,
		Events:         publisher,
		SessionTTL:     30 * 24 * time.Hour,
		VerifyTokenTTL: 24 * time.Hour,
	})

	authHandler := connecttransport.NewAuthHandler(registerUC)

	mux := http.NewServeMux()
	mux.Handle("/healthz", healthHandler())
	mux.Handle("/readyz", readyHandler(pool, rdb))

	authPath, authH := authv1connect.NewAuthServiceHandler(
		authHandler,
		connect.WithInterceptors(
			middleware.NewAuthInterceptor(tokenService),
			middleware.NewRBACInterceptor(),
		),
	)
	mux.Handle(authPath, authH)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		log.Println("listening on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("http shutdown: %v", err)
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("env %s is required", key)
	}
	return v
}

func runMigrations(dbURL string) error {
	src, err := iofs.New(migrationfiles.Files, ".")
	if err != nil {
		return err
	}
	m, err := migrate.NewWithSourceInstance("iofs", src, dbURL)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func healthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}
}

func readyHandler(pool *pgxpool.Pool, rdb *goredis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			http.Error(w, "postgres unavailable", http.StatusServiceUnavailable)
			return
		}
		if err := rdb.Ping(r.Context()).Err(); err != nil {
			http.Error(w, "redis unavailable", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	}
}
