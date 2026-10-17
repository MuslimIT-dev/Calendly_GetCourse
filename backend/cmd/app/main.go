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
	"github.com/joho/godotenv"
	goredis "github.com/redis/go-redis/v9"
	kafkago "github.com/segmentio/kafka-go"

	migrationfiles "github.com/MuslimIT-dev/zaman/backend/db/migrations"
	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
	"github.com/MuslimIT-dev/zaman/backend/internal/infrastructure/broker"
	"github.com/MuslimIT-dev/zaman/backend/internal/infrastructure/cache"
	postgresdb "github.com/MuslimIT-dev/zaman/backend/internal/repository/db"
	"github.com/MuslimIT-dev/zaman/backend/internal/repository/postgres"
	"github.com/MuslimIT-dev/zaman/backend/internal/transport/connect/middleware"

	"github.com/MuslimIT-dev/zaman/backend/pkg/hasher"
	"github.com/MuslimIT-dev/zaman/backend/pkg/jwt"
	"github.com/MuslimIT-dev/zaman/backend/pkg/password"

	authv1connect "github.com/MuslimIT-dev/zaman/backend/gen/go/auth/v1/authv1connect"
	bookingv1connect "github.com/MuslimIT-dev/zaman/backend/gen/go/booking/v1/bookingv1connect"
	catalogv1connect "github.com/MuslimIT-dev/zaman/backend/gen/go/catalog/v1/catalogv1connect"
	locationv1connect "github.com/MuslimIT-dev/zaman/backend/gen/go/location/v1/locationv1connect"
	masterv1connect "github.com/MuslimIT-dev/zaman/backend/gen/go/master/v1/masterv1connect"
	notificationv1connect "github.com/MuslimIT-dev/zaman/backend/gen/go/notification/v1/notificationv1connect"
	paymentv1connect "github.com/MuslimIT-dev/zaman/backend/gen/go/payment/v1/paymentv1connect"
	reviewv1connect "github.com/MuslimIT-dev/zaman/backend/gen/go/review/v1/reviewv1connect"
	schedulev1connect "github.com/MuslimIT-dev/zaman/backend/gen/go/schedule/v1/schedulev1connect"
	servicev1connect "github.com/MuslimIT-dev/zaman/backend/gen/go/service/v1/servicev1connect"
	userv1connect "github.com/MuslimIT-dev/zaman/backend/gen/go/user/v1/userv1connect"

	"github.com/MuslimIT-dev/zaman/backend/internal/infrastructure/notifier"
	gateway "github.com/MuslimIT-dev/zaman/backend/internal/infrastructure/payment"
	connecttransport "github.com/MuslimIT-dev/zaman/backend/internal/transport/connect"

	authuc "github.com/MuslimIT-dev/zaman/backend/internal/usecase/auth"
	bookinguc "github.com/MuslimIT-dev/zaman/backend/internal/usecase/booking"
	cataloguc "github.com/MuslimIT-dev/zaman/backend/internal/usecase/catalog"
	locationuc "github.com/MuslimIT-dev/zaman/backend/internal/usecase/location"
	masteruc "github.com/MuslimIT-dev/zaman/backend/internal/usecase/master"
	notificationuc "github.com/MuslimIT-dev/zaman/backend/internal/usecase/notification"
	paymentuc "github.com/MuslimIT-dev/zaman/backend/internal/usecase/payment"
	reviewuc "github.com/MuslimIT-dev/zaman/backend/internal/usecase/review"
	scheduleuc "github.com/MuslimIT-dev/zaman/backend/internal/usecase/schedule"
	serviceuc "github.com/MuslimIT-dev/zaman/backend/internal/usecase/service"
	useruc "github.com/MuslimIT-dev/zaman/backend/internal/usecase/user"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on environment variables")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	dbURL := mustEnv("DATABASE_URL")
	redisURL := mustEnv("REDIS_URL")
	kafkaBroker := mustEnv("KAFKA_BROKER")
	jwtSecret := mustEnv("JWT_SECRET")

	log.Println("[1] env loaded")

	if err := runMigrations(dbURL); err != nil {
		log.Fatalf("migrations: %v", err)
	}
	log.Println("[2] migrations done")

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pool.Close()
	log.Println("[3] postgres pool created")

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

	// INITIALIZE REPOSITORIES

	queries := postgresdb.New(pool)

	userRepo := postgres.NewUserRepo(queries)
	masterRepo := postgres.NewMasterRepo(queries)
	locationRepo := postgres.NewLocationRepo(queries)
	serviceRepo := postgres.NewServiceRepo(queries)
	scheduleRepo := postgres.NewScheduleRepo(queries)
	bookingRepo := postgres.NewBookingRepo(queries)
	paymentRepo := postgres.NewPaymentRepo(queries)
	catalogRepo := postgres.NewCatalogRepo(queries)
	reviewRepo := postgres.NewReviewRepo(queries)
	notificationRepo := postgres.NewNotificationRepo(queries)

	paymentGateway := gateway.NewManualGateway()

	// INITIALIZE CACHES
	sessionCache := cache.NewRedisCache[authuc.SessionValue](rdb)
	verifyTokenCache := cache.NewRedisCache[authuc.VerifyEmailValue](rdb)
	passwordResetCache := cache.NewRedisCache[authuc.PasswordResetValue](rdb)
	userCache := cache.NewRedisCache[domain.CachedUser](rdb)
	masterCache := cache.NewRedisCache[domain.CachedMaster](rdb)
	locationCache := cache.NewRedisCache[domain.CachedLocations](rdb)
	serviceCache := cache.NewRedisCache[domain.CachedServices](rdb)
	scheduleCache := cache.NewRedisCache[domain.CachedScheduleRules](rdb)
	categoriesCache := cache.NewRedisCache[domain.CachedCategories](rdb)
	defaultServicesCache := cache.NewRedisCache[domain.CachedDefaultServices](rdb)
	reviewCache := cache.NewRedisCache[domain.CachedReviewSummary](rdb)

	// INITIALIZE KAFKA PUBLISHER
	publisher := broker.NewKafkaPublisher[authuc.UserRegisteredEvent](kafkaWriter)
	resetEvents := broker.NewKafkaPublisher[authuc.PasswordResetRequestedEvent](kafkaWriter)
	passwordEvents := broker.NewKafkaPublisher[authuc.PasswordChangedEvent](kafkaWriter)
	bookingEvents := broker.NewKafkaPublisher[any](kafkaWriter)
	paymentEvents := broker.NewKafkaPublisher[any](kafkaWriter)
	reviewEvents := broker.NewKafkaPublisher[any](kafkaWriter)

	// INITIALIZE SERVICES
	passwordHasher := hasher.NewBcryptHasher()
	tokenService := jwt.NewJWTService(jwtSecret, "calendly-clone")
	breachChecker := password.NewBreachChecker()

	// INITIALIZE USE CASES

	// auth

	registerUC := authuc.NewRegisterUseCase(authuc.Deps{
		Users:          userRepo,
		Masters:		masterRepo,
		Sessions:       sessionCache,
		VerifyTokens:   verifyTokenCache,
		Hasher:         passwordHasher,
		Tokens:         tokenService,
		Breach:         breachChecker,
		Events:         publisher,
		SessionTTL:     30 * 24 * time.Hour,
		VerifyTokenTTL: 24 * time.Hour,
	})

	loginUC := authuc.NewLoginUseCase(authuc.Deps{
		Users:      userRepo,
		Sessions:   sessionCache,
		Hasher:     passwordHasher,
		Tokens:     tokenService,
		SessionTTL: 30 * 24 * time.Hour,
	})

	refreshUC := authuc.NewRefreshUseCase(authuc.Deps{
		Users:      userRepo,
		Sessions:   sessionCache,
		Tokens:     tokenService,
		SessionTTL: 30 * 24 * time.Hour,
	})

	logoutUC := authuc.NewLogoutUseCase(authuc.Deps{
		Sessions: sessionCache,
		Tokens:   tokenService,
	})

	logoutAllUC := authuc.NewLogoutAllUseCase(authuc.Deps{
		Sessions: sessionCache,
		Tokens:   tokenService,
	})

	verifyEmailUC := authuc.NewVerifyEmailUseCase(authuc.Deps{
		Users:        userRepo,
		VerifyTokens: verifyTokenCache,
	})

	forgotPasswordUC := authuc.NewForgotPasswordUseCase(authuc.Deps{
		Users:            userRepo,
		PasswordResets:   passwordResetCache,
		ResetEvents:      resetEvents,
		PasswordResetTTL: time.Hour,
	})

	resetPasswordUC := authuc.NewResetPasswordUseCase(authuc.Deps{
		Users:          userRepo,
		Sessions:       sessionCache,
		PasswordResets: passwordResetCache,
		Hasher:         passwordHasher,
		ResetEvents:    resetEvents,
		PasswordEvents: passwordEvents,
	})

	changePasswordUC := authuc.NewChangePasswordUseCase(authuc.Deps{
		Users:    userRepo,
		Sessions: sessionCache,
		Hasher:   passwordHasher,
		Breach:   breachChecker,
	})

	listSessionsUC := authuc.NewListSessionsUseCase(authuc.Deps{
		Sessions: sessionCache,
	})

	// user

	getUserUC := useruc.NewGetUserUseCase(useruc.Deps{
		Users:    userRepo,
		Cache:    userCache,
		CacheTTL: 10 * time.Minute,
	})

	getMeUC := useruc.NewGetMeUseCase(useruc.Deps{
		Users:    userRepo,
		Cache:    userCache,
		CacheTTL: 10 * time.Minute,
	})

	updateUserUC := useruc.NewUpdateUserUseCase(useruc.Deps{
		Users:    userRepo,
		Cache:    userCache,
		CacheTTL: 10 * time.Minute,
	})

	// master
	masterDeps := masteruc.Deps{
		Masters:  masterRepo,
		Cache:    masterCache,
		CacheTTL: 10 * time.Minute,
	}
	listMastersUC := masteruc.NewListMastersUseCase(masterDeps)
	getMasterUC := masteruc.NewGetMasterUseCase(masterDeps)
	getMyProfileUC := masteruc.NewGetMyProfileUseCase(masterDeps)
	updateMasterUC := masteruc.NewUpdateMasterUseCase(masterDeps)
	updateSlugUC := masteruc.NewUpdateSlugUseCase(masterDeps)

	// location
	locationDeps := locationuc.Deps{
		Locations: locationRepo,
		Masters:   masterRepo,
		Cache:     locationCache,
		CacheTTL:  5 * time.Minute,
	}

	listLocationsUC := locationuc.NewListLocationsUseCase(locationDeps)
	getLocationUC := locationuc.NewGetLocationUseCase(locationDeps)
	createLocationUC := locationuc.NewCreateLocationUseCase(locationDeps)
	updateLocationUC := locationuc.NewUpdateLocationUseCase(locationDeps)
	deleteLocationUC := locationuc.NewDeleteLocationUseCase(locationDeps)

	// service
	serviceDeps := serviceuc.Deps{
		Services: serviceRepo,
		Masters:  masterRepo,
		Cache:    serviceCache,
		CacheTTL: 5 * time.Minute,
	}

	getMyServicesUC := serviceuc.NewGetMyServicesUseCase(serviceDeps)
	addServiceUC := serviceuc.NewAddServiceUseCase(serviceDeps)
	updateServiceUC := serviceuc.NewUpdateServiceUseCase(serviceDeps)
	deleteServiceUC := serviceuc.NewDeleteServiceUseCase(serviceDeps)

	// schedule
	scheduleDeps := scheduleuc.Deps{
		Schedule:  scheduleRepo,
		Locations: locationRepo,
		Masters:   masterRepo,
		Services:  serviceRepo,
		Cache:     scheduleCache,
		CacheTTL:  5 * time.Minute,
	}

	getMonthUC := scheduleuc.NewGetMonthUseCase(scheduleDeps)
	getDayUC := scheduleuc.NewGetDayUseCase(scheduleDeps)
	getNextUC := scheduleuc.NewGetNextAvailableUseCase(scheduleDeps)
	getRulesUC := scheduleuc.NewGetRulesUseCase(scheduleDeps)
	updateRulesUC := scheduleuc.NewUpdateRulesUseCase(scheduleDeps)
	listExceptionsUC := scheduleuc.NewListExceptionsUseCase(scheduleDeps)
	addExceptionUC := scheduleuc.NewAddExceptionUseCase(scheduleDeps)
	deleteExceptionUC := scheduleuc.NewDeleteExceptionUseCase(scheduleDeps)

	// booking
	bookingDeps := bookinguc.Deps{
		Bookings:  bookingRepo,
		Services:  serviceRepo,
		Masters:   masterRepo,
		Users:     userRepo,
		Locations: locationRepo,
		Cache:     userCache,
		CacheTTL:  5 * time.Minute,
	}

	createApptUC := bookinguc.NewCreateUseCase(bookingDeps, bookingEvents)
	getApptUC := bookinguc.NewGetUseCase(bookingDeps)
	listMyApptUC := bookinguc.NewListMyUseCase(bookingDeps)
	listMasterApptUC := bookinguc.NewListMasterUseCase(bookingDeps)
	cancelApptUC := bookinguc.NewCancelUseCase(bookingDeps, bookingEvents)
	rescheduleApptUC := bookinguc.NewRescheduleUseCase(bookingDeps, bookingEvents)
	confirmApptUC := bookinguc.NewConfirmUseCase(bookingDeps, bookingEvents)
	completeApptUC := bookinguc.NewCompleteUseCase(bookingDeps, bookingEvents)
	noShowUC := bookinguc.NewNoShowUseCase(bookingDeps)

	// payment

	paymentDeps := paymentuc.Deps{
		Payments: paymentRepo,
		Users:    userRepo,
		Bookings: bookingRepo,
	}

	createPaymentUC := paymentuc.NewCreateUseCase(paymentDeps, paymentGateway)
	getPaymentUC := paymentuc.NewGetUseCase(paymentDeps)
	listMyPaymentsUC := paymentuc.NewListMyUseCase(paymentDeps)
	listAllPaymentsUC := paymentuc.NewListAllUseCase(paymentDeps)
	refundPaymentUC := paymentuc.NewRefundUseCase(paymentDeps, paymentGateway, paymentEvents)
	cancelPaymentUC := paymentuc.NewCancelUseCase(paymentDeps)
	webhookUC := paymentuc.NewWebhookUseCase(paymentDeps, paymentGateway, paymentEvents)

	catalogDeps := cataloguc.Deps{
		Catalog:         catalogRepo,
		CategoriesCache: categoriesCache,
		ServicesCache:   defaultServicesCache,
		CacheTTL:        10 * time.Minute,
	}
	listCategoriesUC := cataloguc.NewListCategoriesUseCase(catalogDeps)
	getCategoryUC := cataloguc.NewGetCategoryUseCase(catalogDeps)
	listDefaultServicesUC := cataloguc.NewListDefaultServicesUseCase(catalogDeps)
	getDefaultServiceUC := cataloguc.NewGetDefaultServiceUseCase(catalogDeps)
	createCategoryUC := cataloguc.NewCreateCategoryUseCase(catalogDeps)
	updateCategoryUC := cataloguc.NewUpdateCategoryUseCase(catalogDeps)
	deleteCategoryUC := cataloguc.NewDeleteCategoryUseCase(catalogDeps)
	createDefaultServiceUC := cataloguc.NewCreateDefaultServiceUseCase(catalogDeps)
	updateDefaultServiceUC := cataloguc.NewUpdateDefaultServiceUseCase(catalogDeps)
	deleteDefaultServiceUC := cataloguc.NewDeleteDefaultServiceUseCase(catalogDeps)

	reviewDeps := reviewuc.Deps{
		Reviews:  reviewRepo,
		Bookings: bookingRepo,
		Masters:  masterRepo,
		Cache:    reviewCache,
		CacheTTL: 5 * time.Minute,
	}
	createReviewUC := reviewuc.NewCreateUseCase(reviewDeps, reviewEvents)
	getReviewUC := reviewuc.NewGetUseCase(reviewDeps)
	listTargetReviewsUC := reviewuc.NewListTargetUseCase(reviewDeps)
	listMyReviewsUC := reviewuc.NewListMyUseCase(reviewDeps)
	replyReviewUC := reviewuc.NewReplyUseCase(reviewDeps, reviewEvents)
	updateReviewUC := reviewuc.NewUpdateUseCase(reviewDeps)
	deleteReviewUC := reviewuc.NewDeleteUseCase(reviewDeps)
	hideReviewUC := reviewuc.NewHideUseCase(reviewDeps)
	summaryReviewUC := reviewuc.NewGetSummaryUseCase(reviewDeps)

	notificationDeps := notificationuc.Deps{
		Notifications: notificationRepo,
		Users:         userRepo,
	}
	notifierRegistry := notificationuc.NewNotifierRegistry(
		notifier.NewInAppNotifier(),
		notifier.NewEmailNotifier(""),
		notifier.NewTelegramNotifier(""),
	)
	sendNotificationUC := notificationuc.NewSendUseCase(notificationDeps, notifierRegistry)
	listNotificationsUC := notificationuc.NewListMyUseCase(notificationDeps)
	markNotificationReadUC := notificationuc.NewMarkReadUseCase(notificationDeps)
	markAllNotificationsReadUC := notificationuc.NewMarkAllReadUseCase(notificationDeps)
	deleteNotificationUC := notificationuc.NewDeleteUseCase(notificationDeps)
	unreadNotificationCountUC := notificationuc.NewUnreadCountUseCase(notificationDeps)
	getNotificationSettingsUC := notificationuc.NewGetSettingsUseCase(notificationDeps)
	updateNotificationSettingsUC := notificationuc.NewUpdateSettingsUseCase(notificationDeps)

	// INITIALIZE HANDLERS

	authHandler := connecttransport.NewAuthHandler(
		registerUC,
		loginUC,
		refreshUC,
		verifyEmailUC,
		forgotPasswordUC,
		resetPasswordUC,
		logoutUC,
		logoutAllUC,
		changePasswordUC,
		listSessionsUC,
	)

	userHandler := connecttransport.NewUserHandler(getUserUC, getMeUC, updateUserUC)

	masterHandler := connecttransport.NewMasterHandler(
		listMastersUC,
		getMasterUC,
		getMyProfileUC,
		updateMasterUC,
		updateSlugUC,
	)

	locationHandler := connecttransport.NewLocationHandler(
		listLocationsUC,
		getLocationUC,
		createLocationUC,
		updateLocationUC,
		deleteLocationUC,
	)

	serviceHandler := connecttransport.NewServiceHandler(
		getMyServicesUC,
		addServiceUC,
		updateServiceUC,
		deleteServiceUC,
	)

	scheduleHandler := connecttransport.NewScheduleHandler(
		getMonthUC, getDayUC, getNextUC,
		getRulesUC, updateRulesUC,
		listExceptionsUC, addExceptionUC, deleteExceptionUC,
	)

	bookingHandler := connecttransport.NewBookingHandler(
		createApptUC, getApptUC, listMyApptUC, listMasterApptUC,
		cancelApptUC, rescheduleApptUC, confirmApptUC, completeApptUC, noShowUC,
	)

	paymentHandler := connecttransport.NewPaymentHandler(
		createPaymentUC, getPaymentUC,
		listMyPaymentsUC, listAllPaymentsUC,
		refundPaymentUC, cancelPaymentUC, webhookUC,
	)

	catalogHandler := connecttransport.NewCatalogHandler(
		listCategoriesUC, getCategoryUC, listDefaultServicesUC, getDefaultServiceUC,
	)
	adminCatalogHandler := connecttransport.NewAdminCatalogHandler(
		createCategoryUC, updateCategoryUC, deleteCategoryUC,
		createDefaultServiceUC, updateDefaultServiceUC, deleteDefaultServiceUC,
	)
	reviewHandler := connecttransport.NewReviewHandler(
		createReviewUC, getReviewUC, listTargetReviewsUC, listMyReviewsUC,
		replyReviewUC, updateReviewUC, deleteReviewUC, hideReviewUC, summaryReviewUC,
	)
	notificationHandler := connecttransport.NewNotificationHandler(
		listNotificationsUC, markNotificationReadUC, markAllNotificationsReadUC,
		deleteNotificationUC, unreadNotificationCountUC, getNotificationSettingsUC,
		updateNotificationSettingsUC, sendNotificationUC,
	)

	// INITIALIZE HTTP SERVER

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

	userPath, userH := userv1connect.NewUserServiceHandler(
		userHandler,
		connect.WithInterceptors(
			middleware.NewAuthInterceptor(tokenService),
			middleware.NewRBACInterceptor(),
		),
	)
	mux.Handle(userPath, userH)

	masterPath, masterH := masterv1connect.NewMasterServiceHandler(
		masterHandler,
		connect.WithInterceptors(
			middleware.NewAuthInterceptor(tokenService),
			middleware.NewRBACInterceptor(),
		),
	)
	mux.Handle(masterPath, masterH)

	locationPath, locationH := locationv1connect.NewLocationServiceHandler(
		locationHandler,
		connect.WithInterceptors(
			middleware.NewAuthInterceptor(tokenService),
			middleware.NewRBACInterceptor(),
		),
	)
	mux.Handle(locationPath, locationH)

	servicePath, serviceH := servicev1connect.NewServiceServiceHandler(
		serviceHandler,
		connect.WithInterceptors(
			middleware.NewAuthInterceptor(tokenService),
			middleware.NewRBACInterceptor(),
		),
	)
	mux.Handle(servicePath, serviceH)

	schedulePath, scheduleH := schedulev1connect.NewScheduleServiceHandler(
		scheduleHandler,
		connect.WithInterceptors(
			middleware.NewAuthInterceptor(tokenService),
			middleware.NewRBACInterceptor(),
		),
	)
	mux.Handle(schedulePath, scheduleH)

	bookingPath, bookingH := bookingv1connect.NewBookingServiceHandler(
		bookingHandler,
		connect.WithInterceptors(
			middleware.NewAuthInterceptor(tokenService),
			middleware.NewRBACInterceptor(),
		),
	)
	mux.Handle(bookingPath, bookingH)

	paymentPath, paymentH := paymentv1connect.NewPaymentServiceHandler(
		paymentHandler,
		connect.WithInterceptors(
			middleware.NewAuthInterceptor(tokenService),
			middleware.NewRBACInterceptor(),
		),
	)
	mux.Handle(paymentPath, paymentH)

	catalogPath, catalogH := catalogv1connect.NewCatalogServiceHandler(
		catalogHandler,
		connect.WithInterceptors(
			middleware.NewAuthInterceptor(tokenService),
			middleware.NewRBACInterceptor(),
		),
	)
	mux.Handle(catalogPath, catalogH)

	adminCatalogPath, adminCatalogH := catalogv1connect.NewAdminCatalogServiceHandler(
		adminCatalogHandler,
		connect.WithInterceptors(
			middleware.NewAuthInterceptor(tokenService),
			middleware.NewRBACInterceptor(),
		),
	)
	mux.Handle(adminCatalogPath, adminCatalogH)

	reviewPath, reviewH := reviewv1connect.NewReviewServiceHandler(
		reviewHandler,
		connect.WithInterceptors(
			middleware.NewAuthInterceptor(tokenService),
			middleware.NewRBACInterceptor(),
		),
	)
	mux.Handle(reviewPath, reviewH)

	notificationPath, notificationH := notificationv1connect.NewNotificationServiceHandler(
		notificationHandler,
		connect.WithInterceptors(
			middleware.NewAuthInterceptor(tokenService),
			middleware.NewRBACInterceptor(),
		),
	)
	mux.Handle(notificationPath, notificationH)

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
