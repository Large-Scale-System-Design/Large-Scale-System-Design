package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
	"protocol"
)

var ctx = context.Background()
var rdb *redis.Client

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	rdb = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	http.HandleFunc("/send", handleSend)
	log.Println("API Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.NotificationPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// 필수 값 검증
	if req.Channel == "" || req.Recipient == "" {
		http.Error(w, "Missing channel or recipient", http.StatusBadRequest)
		return
	}

	req.RetryCount = 0
	payload, err := json.Marshal(req)
	if err != nil {
		http.Error(w, "JSON marshal error", http.StatusInternalServerError)
		return
	}

	// [핵심] 동적 큐 라우팅 (예: queue:email, queue:sms)
	queueName := protocol.QueuePrefix + req.Channel

	err = rdb.RPush(ctx, queueName, payload).Err()
	if err != nil {
		log.Printf("Redis error: %v", err)
		http.Error(w, "Failed to enqueue", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Notification queued"))
}
