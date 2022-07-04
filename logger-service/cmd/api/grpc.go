package main

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"logging/data"
	"logging/logs"
	"net"
)

type LogServer struct {
	logs.UnimplementedLogServiceServer
	Models data.Models
}

func (l *LogServer) WriteLog(ctx context.Context, r *logs.LogRequest) (*logs.LogResponse, error) {
	i := r.GetLogEntry()
	if err := l.Models.LogEntry.Insert(data.LogEntry{
		Name: i.Name,
		Data: i.Data,
	}); err != nil {
		return &logs.LogResponse{
			Result: "failed",
		}, err
	}
	return &logs.LogResponse{
		Result: "Logged!",
	}, nil
}

func (app *Config) gRPCListen() {
	l, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen for GRPC: %v", err)
	}
	s := grpc.NewServer()

	logs.RegisterLogServiceServer(s, &LogServer{Models: app.Models})

	log.Printf("gRPC Server started on port %s", grpcPort)

	if err := s.Serve(l); err != nil {
		log.Fatalf("failed to listen for gRPC: %v", err)
	}

}
