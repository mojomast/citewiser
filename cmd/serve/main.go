package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/mojomast/citewiser/pkg/rag"
)

func main() {
	addr := os.Getenv("CITEWISERAG_ADDR")
	if len(os.Args) > 1 && os.Args[1] == "stdio" {
		if err := runStdio(os.Stdin, os.Stdout, rag.NewPipeline()); err != nil {
			slog.ErrorContext(context.Background(), "stdio request failed", "error", err)
			os.Exit(1)
		}
		return
	}
	if addr == "" {
		addr = ":8080"
	}
	server := &http.Server{Addr: addr, Handler: newServer(rag.NewPipeline()), ReadHeaderTimeout: 5 * time.Second}
	slog.InfoContext(context.Background(), "starting citewiserag server", "addr", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.ErrorContext(context.Background(), "server failed", "error", err)
		os.Exit(1)
	}
}
