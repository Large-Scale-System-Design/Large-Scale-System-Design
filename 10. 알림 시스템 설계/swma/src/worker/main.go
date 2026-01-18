package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"protocol"
	"worker/sender"
)

var ctx = context.Background()
const MaxRetries = 3

func main() {
	// 1. 환경변수 확인 및 의존성 주입
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	
	// 이 워커가 담당할 채널 타입 (예: email)
	channelType := os.Getenv("CHANNEL_TYPE")
	if channelType == "" {
		log.Fatal("CHANNEL_TYPE is required")
	}

	// 2. Sender 구현체 선택 (Factory Pattern)
	var currentSender sender.Sender
	
	switch channelType {
	case protocol.ChannelEmail:
		currentSender = sender.NewEmailSender()
		log.Println("Worker initialized for channel: EMAIL")
	default:
		log.Fatalf("Unsupported channel type: %s", channelType)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// 3. 큐 이름 결정
	queueName := protocol.QueuePrefix + channelType
	queueRetry := queueName + ":retry"
	queueDLQ := queueName + ":dlq"

	// [Goroutine] 재시도 모니터링
	go func() {
		for {
			time.Sleep(10 * time.Second)
			val, err := rdb.RPopLPush(ctx, queueRetry, queueName).Result()
			if err == nil {
				log.Printf("[RETRY MANAGER] Re-queued: %s", val)
			}
		}
	}()

	log.Printf("Listening on queue: %s", queueName)

	// [Main Loop]
	for {
		result, err := rdb.BLPop(ctx, 0, queueName).Result()
		if err != nil {
			log.Printf("Redis connection error: %v. Retrying...", err)
			time.Sleep(3 * time.Second)
			continue
		}

		payloadStr := result[1]
		processNotification(rdb, currentSender, payloadStr, queueRetry, queueDLQ)
	}
}

func processNotification(rdb *redis.Client, s sender.Sender, payloadStr string, qRetry, qDLQ string) {
	var req protocol.NotificationPayload
	if err := json.Unmarshal([]byte(payloadStr), &req); err != nil {
		log.Printf("[ERROR] Bad payload: %v", err)
		return
	}

	log.Printf("[ATTEMPT %d/%d] Sending to %s via %s", req.RetryCount+1, MaxRetries, req.Recipient, req.Channel)

	// 다형성 활용: 구체적인 Sender가 무엇인지 몰라도 Send 호출
	err := s.Send(req.Recipient, req.Properties)
	if err == nil {
		log.Println("[SUCCESS]")
		return
	}

	log.Printf("[FAIL] %v", err)

	// 재시도 처리
	req.RetryCount++
	newPayload, _ := json.Marshal(req)

	if req.RetryCount >= MaxRetries {
		log.Println("[GIVE UP] Moving to DLQ")
		rdb.RPush(ctx, qDLQ, newPayload)
	} else {
		log.Println("[RETRY] Moving to Retry Queue")
		rdb.RPush(ctx, qRetry, newPayload)
	}
}
