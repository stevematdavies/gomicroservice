package main

import (
	"logging/data"
	"logging/logs"
)

type LogServer struct {
	logs.UnimplementedLogServiceServer
	Models data.Models
}

func (l *LogServer) WriteLog(_, r *logs.LogRequest) (*logs.LogResponse, error) {
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
