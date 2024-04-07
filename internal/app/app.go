package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ZakirAvrora/exchange-rate/config"
	v1 "github.com/ZakirAvrora/exchange-rate/internal/controller/http/v1"
	"github.com/ZakirAvrora/exchange-rate/internal/exchangerates"
	"github.com/ZakirAvrora/exchange-rate/internal/exchangerates/repo"
	"github.com/ZakirAvrora/exchange-rate/pkg/external/exchangeratesapi"
	"github.com/ZakirAvrora/exchange-rate/pkg/httpserver"
	"github.com/ZakirAvrora/exchange-rate/pkg/postgres"
	"github.com/gin-gonic/gin"
)

func Run(cfg *config.Config) {

	// Repository
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.PostgresConfig.User,
		cfg.PostgresConfig.Password,
		cfg.PostgresConfig.Host,
		cfg.PostgresConfig.Port,
		cfg.PostgresConfig.DbName)

	connPool, err := postgres.New(dbUrl, postgres.MaxPoolSize(cfg.PostgresConfig.PoolMax))
	if err != nil {
		log.Fatalln(err)
	}
	defer connPool.Close()

	rep, err := repo.NewRecordsRepository(connPool)
	if err != nil {
		log.Fatalln(err)
	}

	// Migrate
	if err := initMigrate(dbUrl); err != nil {
		log.Fatalln(err)
	}

	// Service
	recorder := exchangerates.NewService(rep)

	// External client
	client, err := exchangeratesapi.NewProvider()
	if err != nil {
		log.Fatalln(err)
	}

	ch := recorder.Queue()
	go func() {
		for {
			ctx := context.Background()
			t := <-ch
			func(r exchangerates.Record) {
				rate, err := client.GetLatestRate(ctx, r.Base, r.Secondary)
				if err != nil {
					log.Fatalln(err)
				}
				log.Println("Rate:", rate)
				if err = recorder.Update(ctx, r.Identifier, rate.Value); err != nil {
					log.Fatalln(err)
				}
			}(t)
		}
	}()

	// HTTP Server
	handler := gin.New()
	v1.NewRouter(handler, recorder)
	httpServer := httpserver.New(handler, httpserver.Port(cfg.HTTP.Port))

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		log.Println("app - Run - signal: " + s.String())
	case err = <-httpServer.Notify():
		log.Println(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	}

	// Gracefull Shutdown
	err = httpServer.Shutdown()
	if err != nil {
		log.Println(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}

}
