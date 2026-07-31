package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/alexhokl/privateserver/server"
	"github.com/alexhokl/todo-cli/database"
	"github.com/alexhokl/todo-cli/internal"
	"github.com/alexhokl/todo-cli/proto"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/gorm"
)

// DefaultPort is the port the gRPC server listens on when none is configured.
const DefaultPort = 8080

const maxMessageSize = 10 * 1024 * 1024 // 10 MB

const shutdownTimeout = 30 * time.Second

type serveOptions struct {
	Port                    int
	DatabaseFilePath        string
	Hostname                string
	TailscaleAuthKey        string
	TailscaleStateDirectory string
}

var serveOpts serveOptions

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts a server of todo",
	RunE:  runServe,
	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
		bindEnvironmentVariablesToServeOptions(cmd, &serveOpts)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	flags := serveCmd.Flags()
	flags.IntVarP(&serveOpts.Port, "port", "p", DefaultPort, "Port to run the server on")
	flags.StringVarP(&serveOpts.DatabaseFilePath, "database", "d", defaultDatabaseFilePath(), "Path to the database file")
	flags.StringVar(&serveOpts.Hostname, "hostname", "", "Hostname for Tailscale (if empty, the server runs without Tailscale authentication)")
	flags.StringVar(&serveOpts.TailscaleAuthKey, "ts-auth-key", "", "Tailscale auth key (required when --hostname is set)")
	flags.StringVar(&serveOpts.TailscaleStateDirectory, "ts-state-dir", "./tailscale-state", "Directory to store Tailscale state (if empty, it would use a temporary directory)")

	_ = viper.BindPFlag("port", flags.Lookup("port"))
	_ = viper.BindPFlag("database", flags.Lookup("database"))
	_ = viper.BindPFlag("hostname", flags.Lookup("hostname"))
	_ = viper.BindPFlag("ts_auth_key", flags.Lookup("ts-auth-key"))
	_ = viper.BindPFlag("ts_state_dir", flags.Lookup("ts-state-dir"))
}

// defaultDatabaseFilePath returns $HOME/.todo.db, falling back to a relative
// path when the home directory cannot be determined.
func defaultDatabaseFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Sprintf(".%s.db", AppName)
	}
	return filepath.Join(home, fmt.Sprintf(".%s.db", AppName))
}

func bindEnvironmentVariablesToServeOptions(cmd *cobra.Command, opts *serveOptions) {
	if !cmd.Flags().Changed("port") {
		if v := viper.GetInt("port"); v != 0 {
			opts.Port = v
		}
	}
	if !cmd.Flags().Changed("database") {
		if v := viper.GetString("database"); v != "" {
			opts.DatabaseFilePath = v
		}
	}
	if opts.Hostname == "" {
		opts.Hostname = viper.GetString("hostname")
	}
	if opts.TailscaleAuthKey == "" {
		opts.TailscaleAuthKey = viper.GetString("ts_auth_key")
	}
	if !cmd.Flags().Changed("ts-state-dir") {
		if v := viper.GetString("ts_state_dir"); v != "" {
			opts.TailscaleStateDirectory = v
		}
	}
}

func validateFlags(opts serveOptions) error {
	if opts.Port <= 0 || opts.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", opts.Port)
	}
	if opts.DatabaseFilePath == "" {
		return fmt.Errorf("database file path cannot be empty")
	}
	if opts.Hostname != "" && opts.TailscaleAuthKey == "" {
		return fmt.Errorf("--ts-auth-key is required when --hostname is set")
	}
	return nil
}

func runServe(cmd *cobra.Command, _ []string) error {
	if err := validateFlags(serveOpts); err != nil {
		return err
	}

	// Initialise OpenTelemetry (TracerProvider + MeterProvider + LoggerProvider
	// + default slog). This must happen before any slog calls so all log output
	// is routed through the OTLP log exporter and carries trace correlation
	// fields.
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	otelShutdown, err := internal.SetupOTel(ctx)
	if err != nil {
		// Non-fatal: log to stderr and continue without OTel rather than
		// refusing to start the server.
		slog.Error("failed to set up OpenTelemetry", slog.String("error", err.Error()))
	}
	defer func() {
		if shutdownErr := otelShutdown(context.Background()); shutdownErr != nil {
			slog.Error("error shutting down OpenTelemetry", slog.String("error", shutdownErr.Error()))
		}
	}()

	dbConn, err := database.Open(serveOpts.DatabaseFilePath)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	var listener net.Listener
	var unaryAuth grpc.UnaryServerInterceptor
	var streamAuth grpc.StreamServerInterceptor

	if serveOpts.Hostname != "" {
		privateServerConfig := &server.ServerConfig{
			Hostname:                serveOpts.Hostname,
			TailscaleAuthKey:        serveOpts.TailscaleAuthKey,
			TailscaleStateDirectory: serveOpts.TailscaleStateDirectory,
		}

		privateServer, err := server.NewServer(privateServerConfig)
		if err != nil {
			return fmt.Errorf("failed to create private server: %w", err)
		}
		defer func() { _ = privateServer.Close() }()

		listeners, _, _, err := privateServer.Listen([]int{serveOpts.Port})
		if err != nil {
			return fmt.Errorf("failed to start private server: %w", err)
		}
		listener = listeners[0]

		interceptor := internal.NewTailscaleAuthenticationInterceptor(dbConn, privateServer)
		unaryAuth = interceptor.Intercept
		streamAuth = interceptor.InterceptStream

		slog.InfoContext(ctx, "Tailscale is enabled", slog.String("hostname", serveOpts.Hostname))
	} else {
		listener, err = net.Listen("tcp", fmt.Sprintf(":%d", serveOpts.Port))
		if err != nil {
			slog.ErrorContext(
				ctx,
				"failed to listen to port of gRPC server",
				slog.Int("port", serveOpts.Port),
				slog.String("error", err.Error()),
			)
			return err
		}

		unaryAuth = internal.DummyAuthenticationInterceptor
		streamAuth = internal.DummyStreamAuthenticationInterceptor
	}

	// Create a context that will be cancelled on SIGTERM/SIGINT
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	g, ctx := errgroup.WithContext(ctx)

	grpcServer := getGrpcServer(dbConn, unaryAuth, streamAuth)

	// Goroutine to handle shutdown signals
	g.Go(func() error {
		select {
		case sig := <-sigChan:
			slog.InfoContext(context.Background(), "received shutdown signal", slog.String("signal", sig.String()))
		case <-ctx.Done():
			return nil
		}

		slog.InfoContext(context.Background(), "initiating graceful shutdown")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer shutdownCancel()

		// Stops accepting new connections and waits for existing ones
		grpcServer.GracefulStop()
		slog.InfoContext(shutdownCtx, "gRPC server stopped gracefully")

		cancel()
		return nil
	})

	// gRPC server goroutine
	g.Go(func() error {
		slog.InfoContext(
			ctx,
			"gRPC server is serving",
			slog.Int("port", serveOpts.Port),
		)

		if err := grpcServer.Serve(listener); err != nil {
			// grpc.Server.Serve returns nil on GracefulStop, but check for
			// other errors
			select {
			case <-ctx.Done():
				return nil
			default:
				return err
			}
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

	slog.InfoContext(context.Background(), "server shutdown complete")
	return nil
}

// getGrpcServer builds the gRPC server with the standard interceptor chain and
// OpenTelemetry instrumentation. Service implementations are registered here.
func getGrpcServer(dbConn *gorm.DB, unaryAuth grpc.UnaryServerInterceptor, streamAuth grpc.StreamServerInterceptor) *grpc.Server {
	grpcServer := grpc.NewServer(
		// OTel stats handler: starts a span for every incoming RPC and
		// propagates W3C trace context so logs emitted inside handlers are
		// correlated with the active trace.
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.MaxRecvMsgSize(maxMessageSize),
		grpc.MaxSendMsgSize(maxMessageSize),
		grpc.ChainUnaryInterceptor(
			unaryAuth,
			internal.ErrorLoggingInterceptor,
		),
		grpc.ChainStreamInterceptor(
			streamAuth,
		),
	)

	proto.RegisterItemServiceServer(grpcServer, internal.NewItemServer(dbConn))

	// Server reflection allows tools such as grpcurl to discover the available
	// services without a compiled copy of the protobuf descriptors.
	reflection.Register(grpcServer)

	return grpcServer
}