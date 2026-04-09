package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"

	"github.com/anand34577/proxmox-mcp/internal/config"
	"github.com/anand34577/proxmox-mcp/internal/logger"
	pxclient "github.com/anand34577/proxmox-mcp/internal/proxmox"
	"github.com/anand34577/proxmox-mcp/internal/tools"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
    log, err := logger.New(cfg.Log.Level, cfg.Log.Format)
	if err != nil {
		return fmt.Errorf("build logger: %w", err)
	}
	defer log.Sync() //nolint:errcheck

	log.Info("starting Proxmox MCP Server",
		zap.String("name", cfg.Server.Name),
		zap.String("version", cfg.Server.Version),
		zap.String("transport", cfg.Server.Transport),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pxClient, err := pxclient.New(ctx, &cfg.Proxmox, log)
	if err != nil {
		return fmt.Errorf("proxmox client: %w", err)
	}

	mcpServer := server.NewMCPServer(
		cfg.Server.Name,
		cfg.Server.Version,
		server.WithToolCapabilities(true),
		server.WithRecovery(), // auto-recover from panics in handlers
	)
	handler := tools.NewHandler(pxClient, cfg, log)
	tools.RegisterAll(mcpServer, handler)

	log.Info("tools registered",
		zap.Bool("allow_destructive", cfg.Security.AllowDestructive),
	)

	errCh := make(chan error, 1)

	switch cfg.Server.Transport {
	case "stdio":
		log.Info("transport: stdio")
		go func() {
			errCh <- server.ServeStdio(mcpServer)
		}()

	case "sse":
		log.Info("transport: SSE (Server-Sent Events)",
			zap.String("addr", cfg.Server.ListenAddr))
		sseServer := server.NewSSEServer(mcpServer,
			server.WithBaseURL(fmt.Sprintf("http://%s", cfg.Server.ListenAddr)),
		)
		go func() {
			errCh <- sseServer.Start(cfg.Server.ListenAddr)
		}()

	case "http":
		log.Info("transport: Streamable HTTP",
			zap.String("addr", cfg.Server.ListenAddr))
		httpServer := server.NewStreamableHTTPServer(mcpServer)
		go func() {
			errCh <- httpServer.Start(cfg.Server.ListenAddr)
		}()

	default:
		return fmt.Errorf("unknown transport %q", cfg.Server.Transport)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Info("shutdown signal received", zap.String("signal", sig.String()))
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}
	}

	log.Info("server stopped gracefully")
	return nil
}