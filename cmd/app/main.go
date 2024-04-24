package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/tsel-ticketmaster/tm-order/config"
	customerapp_event "github.com/tsel-ticketmaster/tm-order/internal/module/customerapp/event"
	"github.com/tsel-ticketmaster/tm-order/internal/module/customerapp/midtrans"
	customerapp_order "github.com/tsel-ticketmaster/tm-order/internal/module/customerapp/order"
	customerapp_ticket "github.com/tsel-ticketmaster/tm-order/internal/module/customerapp/ticket"
	"github.com/tsel-ticketmaster/tm-order/internal/pkg/jwt"
	internalMiddleare "github.com/tsel-ticketmaster/tm-order/internal/pkg/middleware"
	"github.com/tsel-ticketmaster/tm-order/internal/pkg/session"
	"github.com/tsel-ticketmaster/tm-order/pkg/applogger"
	"github.com/tsel-ticketmaster/tm-order/pkg/gctasks"
	"github.com/tsel-ticketmaster/tm-order/pkg/kafka"
	"github.com/tsel-ticketmaster/tm-order/pkg/middleware"
	"github.com/tsel-ticketmaster/tm-order/pkg/monitoring"
	"github.com/tsel-ticketmaster/tm-order/pkg/postgresql"
	"github.com/tsel-ticketmaster/tm-order/pkg/pubsub"
	"github.com/tsel-ticketmaster/tm-order/pkg/redis"
	"github.com/tsel-ticketmaster/tm-order/pkg/server"
	"github.com/tsel-ticketmaster/tm-order/pkg/validator"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
)

var (
	c           *config.Config
	CustomerApp string
	AdminApp    string
)

func init() {
	c = config.Get()
	AdminApp = fmt.Sprintf("%s/%s", c.Application.Name, "adminapp")
	CustomerApp = fmt.Sprintf("%s/%s", c.Application.Name, "customerapp")
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := applogger.GetLogrus()

	mon := monitoring.NewOpenTelemetry(
		c.Application.Name,
		c.Application.Environment,
		c.GCP.ProjectID,
	)

	mon.Start(ctx)

	validate := validator.Get()

	hc := http.DefaultClient

	jsonWebToken := jwt.NewJSONWebToken(c.JWT.PrivateKey, c.JWT.PublicKey)

	psqldb := postgresql.GetDatabase()
	if err := psqldb.Ping(); err != nil {
		logger.WithContext(ctx).WithError(err).Error()
	}

	publisher := pubsub.PublisherFromConfluentKafkaProducer(logger, kafka.NewProducer())

	rc := redis.GetClient()
	if err := rc.Ping(context.Background()).Err(); err != nil {
		logger.WithContext(ctx).WithError(err).Error()
	}

	cloudTask := gctasks.NewGCTasks(logger, c.GCP.ProjectID, c.GCP.ServiceAccount)

	session := session.NewRedisSessionStore(logger, rc)

	customerSessionMiddleware := internalMiddleare.NewCustomerSessionMiddleware(jsonWebToken, session)

	router := mux.NewRouter()
	router.Use(
		otelmux.Middleware(c.Application.Name),
		middleware.HTTPResponseTraceInjection,
		middleware.NewHTTPRequestLogger(logger, c.Application.Debug, http.StatusInternalServerError).Middleware,
	)

	// admin's app

	// customer's app
	customerappEventRepo := customerapp_event.NewEventRepository(logger, psqldb)
	customerappShowRepo := customerapp_event.NewShowRepository(logger, psqldb)
	customerappOrderRepo := customerapp_order.NewOrderRepository(logger, psqldb)
	customerappOrderItemRepo := customerapp_order.NewItemRepository(logger, psqldb)
	customerappOrderRuleRangeDateRepo := customerapp_order.NewOrderRuleRangeDateRepository(logger, psqldb)
	customerappOrderRuleDayRepo := customerapp_order.NewOrderRuleDayRepository(logger, psqldb)
	customerappTicketRepo := customerapp_ticket.NewTicketStockRepository(logger, psqldb)
	midtransRepo := midtrans.NewMidtransRepository(c.Midtrans.BaseURL, c.Midtrans.BasicAuthKey, logger, hc)
	customerappOrderUseCase := customerapp_order.NewOrderUseCase(customerapp_order.OrderUseCaseProperty{
		Logger:                       logger,
		Timeout:                      c.Application.Timeout,
		BaseURL:                      c.Application.TMOrder.BaseURL,
		OrderExpireDuration:          c.Order.Expiration,
		ServiceChargePercentage:      c.Order.ServiceChargePercentage,
		TaxPercentage:                c.Order.TaxChargePercentage,
		EventRepository:              customerappEventRepo,
		ShowRepository:               customerappShowRepo,
		TicketStockRepository:        customerappTicketRepo,
		OrderRuleRangeDateRepository: customerappOrderRuleRangeDateRepo,
		OrderRuleDay:                 customerappOrderRuleDayRepo,
		OrderRepository:              customerappOrderRepo,
		ItemRepository:               customerappOrderItemRepo,
		Publisher:                    publisher,
		MidtransRepository:           midtransRepo,
		CloudTask:                    cloudTask,
	})
	customerapp_order.InitHTTPHandler(router, customerSessionMiddleware, validate, customerappOrderUseCase)

	handler := middleware.SetChain(
		router,
		cors.New(cors.Options{
			AllowedOrigins:   c.CORS.AllowedOrigins,
			AllowedMethods:   c.CORS.AllowedMethods,
			AllowedHeaders:   c.CORS.AllowedHeaders,
			ExposedHeaders:   c.CORS.ExposedHeaders,
			MaxAge:           c.CORS.MaxAge,
			AllowCredentials: c.CORS.AllowCredentials,
		}).Handler,
	)

	srv := &server.Server{
		Server: http.Server{
			Addr:    fmt.Sprintf(":%d", c.Application.Port),
			Handler: handler,
		},
		Logger: logger,
	}

	go func() {
		srv.ListenAndServe()
	}()

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm

	srv.Shutdown(ctx)
	publisher.Close()
	psqldb.Close()
	rc.Close()
	mon.Stop(ctx)
}
