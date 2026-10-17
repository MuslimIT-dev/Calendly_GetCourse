package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/MuslimIT-dev/zaman/backend/internal/infrastructure/broker"
	"github.com/MuslimIT-dev/zaman/backend/internal/infrastructure/notifier"
	"github.com/MuslimIT-dev/zaman/backend/internal/repository/db"
	"github.com/MuslimIT-dev/zaman/backend/internal/repository/postgres"
	notificationuc "github.com/MuslimIT-dev/zaman/backend/internal/usecase/notification"
	"github.com/MuslimIT-dev/zaman/backend/internal/worker"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on environment variables")
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	pool, err := pgxpool.New(ctx, requiredEnv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pool.Close()

	brokers := splitBrokers(requiredEnv("KAFKA_BROKER"))
	queries := db.New(pool)
	notificationRepo := postgres.NewNotificationRepo(queries)
	userRepo := postgres.NewUserRepo(queries)
	masterRepo := postgres.NewMasterRepo(queries)

	registry := notificationuc.NewNotifierRegistry(
		notifier.NewInAppNotifier(),
		notifier.NewEmailNotifier(os.Getenv("EMAIL_FROM")),
		notifier.NewTelegramNotifier(os.Getenv("TELEGRAM_BOT_TOKEN")),
	)
	
	sendUC := notificationuc.NewSendUseCase(notificationuc.Deps{
		Notifications: notificationRepo,
		Users:         userRepo,
	}, registry)

	consumers := []consumerRunner{
		{consumer: broker.NewKafkaConsumer[worker.AppointmentEvent](
			brokers, "appointments", "notification-worker-appointments", worker.NewAppointmentsConsumer(sendUC, masterRepo),
		)},
		{consumer: broker.NewKafkaConsumer[worker.PaymentEvent](
			brokers, "payments", "notification-worker-payments", worker.NewPaymentsConsumer(sendUC),
		)},
		{consumer: broker.NewKafkaConsumer[worker.ReviewEvent](
			brokers, "reviews", "notification-worker-reviews", worker.NewReviewsConsumer(sendUC, masterRepo),
		)},
		{consumer: broker.NewKafkaConsumer[worker.UserRegisteredEvent](
			brokers, "user.registered", "notification-worker-users", worker.NewUsersConsumer(sendUC),
		)},
	}

	errs := make(chan error, len(consumers))
	for _, runner := range consumers {
		go func(runner consumerRunner) {
			errs <- runner.consumer.Start(ctx)
		}(runner)
	}
	log.Println("notification worker started")
	select {
	case <-ctx.Done():
		log.Println("notification worker shutting down")
	case err := <-errs:
		if err != nil {
			log.Fatalf("notification consumer stopped: %v", err)
		}
	}
}

type consumerRunner struct {
	consumer interface {
		Start(context.Context) error
	}
}

func requiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("env %s is required", key)
	}
	return value
}

func splitBrokers(value string) []string {
	var brokers []string
	for _, brokerAddress := range strings.Split(value, ",") {
		if address := strings.TrimSpace(brokerAddress); address != "" {
			brokers = append(brokers, address)
		}
	}
	if len(brokers) == 0 {
		log.Fatal("KAFKA_BROKER must contain at least one broker address")
	}
	return brokers
}
