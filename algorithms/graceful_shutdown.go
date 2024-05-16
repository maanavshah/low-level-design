package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/cors"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
)

func main() {
	portInfo := ":3000"
	mux := http.NewServeMux()
	muxHandler := cors.AllowAll().Handler(mux) // Allows all origins and request types currently
	// creating a server and providing the handler which acts as a reverse proxy
	server := &http.Server{
		Handler: muxHandler,
	}

	grpcServer := grpc.NewServer()

	l, err := net.Listen("tcp", portInfo)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}

	m := cmux.New(l)
	// a different listener for HTTP1
	httpL := m.Match(cmux.HTTP1Fast())
	// a different listener for HTTP2 since gRPC uses HTTP2
	grpcL := m.Match(cmux.HTTP2())

	allConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
		<-sigint

		fmt.Println("Service interrupt received")

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			fmt.Println("HTTP server shutdown error:", err)
		}
		fmt.Println("HTTP graceful shutdown successful")

		grpcServer.GracefulStop()
		fmt.Println("gRPC graceful shutdown successful")

		close(allConnsClosed)
	}()

	fmt.Println("Starting server on port", portInfo)
	go server.Serve(httpL)
	go grpcServer.Serve(grpcL)
	m.Serve()

	<-allConnsClosed
	fmt.Println("Stopped server")
}
