package api

import "net/http"

func (s *Server) handlerDaemonStatus(w http.ResponseWriter,
	r *http.Request) (int, error) {

	count := s.manager.CountWorkers(r.Context())

	type workerCount struct {
		WorkerCount int `json:"worker_count"`
	}

	code := http.StatusOK
	handlerReplyWithJSON(w, code, workerCount{WorkerCount: count})
	return code, nil
}
