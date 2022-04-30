package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mgjules/minion/build"
	"github.com/mgjules/minion/health"
	"github.com/mgjules/minion/logger"
	"github.com/mgjules/minion/minion"
	"github.com/mgjules/minion/tracer"
	"github.com/mgjules/minion/transport/http"
	"github.com/urfave/cli/v2"
)

var serve = &cli.Command{
	Name:    "serve",
	Aliases: []string{"s"},
	Usage:   "Starts the HTTP server",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "minion-name",
			Value:   "minion",
			Usage:   "name of our minion",
			EnvVars: []string{"MINION_NAME"},
		},
		&cli.BoolFlag{
			Name:    "prod",
			Value:   false,
			Usage:   "whether running in PROD or DEBUG mode",
			EnvVars: []string{"PROD"},
		},
		&cli.StringFlag{
			Name:    "http-server-host",
			Value:   "localhost",
			Usage:   "host/IP for HTTP server",
			EnvVars: []string{"HTTP_SERVER_HOST"},
		},
		&cli.IntFlag{
			Name:    "http-server-port",
			Value:   9001,
			Usage:   "port for HTTP server",
			EnvVars: []string{"HTTP_SERVER_PORT"},
		},
		&cli.Int64Flag{
			Name:    "cache-max-keys",
			Value:   64,
			Usage:   "maximum number of keys for cache",
			EnvVars: []string{"CACHE_MAX_KEYS"},
		},
		&cli.Int64Flag{
			Name:    "cache-max-cost",
			Value:   1000000,
			Usage:   "maximum size of cache (in bytes)",
			EnvVars: []string{"CACHE_MAX_COST"},
		},
		&cli.StringFlag{
			Name:    "jaeger-endpoint",
			Value:   "http://localhost:14268/api/traces",
			Usage:   "jaeger collector endpoint",
			EnvVars: []string{"JAEGER_ENDPOINT"},
		},
		&cli.StringFlag{
			Name:    "amqp-uri",
			Value:   "amqp://guest:guest@localhost:5672",
			Usage:   "amqp 0-9-1 Uniform Resource Identifier",
			EnvVars: []string{"AMQP_URI"},
		},
		&cli.StringFlag{
			Name:        "minion-key",
			Usage:       "sample minion key",
			EnvVars:     []string{"MINION_KEY"},
			DefaultText: "random",
		},
	},
	Action: func(c *cli.Context) error {
		name := c.String("minion-name")
		prod := c.Bool("prod")
		host := c.String("http-server-host")
		port := c.Int("http-server-port")
		jaegerEndpoint := c.String("jaeger-endpoint")
		key := c.String("minion-key")

		logger, err := logger.New(prod)
		if err != nil {
			return fmt.Errorf("logger: %w", err)
		}
		defer logger.Sync() //nolint:errcheck

		tracer, err := tracer.New(prod, name, jaegerEndpoint)
		if err != nil {
			return fmt.Errorf("tracer: %w", err)
		}

		checks := health.New()

		minion, err := minion.New(name, key)
		if err != nil {
			return fmt.Errorf("minion: %w", err)
		}

		info, err := build.New()
		if err != nil {
			return fmt.Errorf("build: %w", err)
		}

		server := http.NewServer(prod, host, port, logger, tracer, checks, minion, info)
		go func() {
			if err := server.Start(); err != nil {
				logger.Errorf("server: %v", err)
			}
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			return fmt.Errorf("server: %w", err)
		}

		return nil
	},
}
