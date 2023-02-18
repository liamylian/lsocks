package dashboard

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/liamylian/lsocks/pkg/log"
)

type Handler struct {
	storage Storage
}

func NewHandler(storage Storage) *Handler {
	return &Handler{storage}
}

func (h *Handler) Serve(addr string) {
	// TODO
	statics := http.FileServer(http.Dir("cmd/dashboard/statics"))
	http.Handle("/", http.StripPrefix("/", statics))

	http.HandleFunc("/api/traffics", h.listTraffics)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.WithError(err).Fatalf("listen http failed: %s", addr)
	}
	return
}

func (h *Handler) listTraffics(w http.ResponseWriter, r *http.Request) {
	// TODO
	records, err := h.storage.List("liam", time.Minute, time.Now().Add(-7*24*time.Hour), time.Now())
	if err != nil {
		w.WriteHeader(500)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	bytes, _ := json.Marshal(records)
	w.Write(bytes)
}
