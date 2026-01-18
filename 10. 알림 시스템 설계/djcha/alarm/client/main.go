package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// 요청 데이터 구조체
type NotificationRequest struct {
	TemplateCode string                 `json:"templateCode"`
	Recipient    string                 `json:"recipient"`
	Args         map[string]interface{} `json:"args"`
}

func main() {
	// 랜덤 시드 설정
	rand.Seed(time.Now().UnixNano())

	targetCount := 10 // 1만개 전송
	var wg sync.WaitGroup

	fmt.Printf("🚀 %d개의 알림 폭격 시작! (정상 + 불량 데이터 혼합)\n", targetCount)
	startTime := time.Now()

	for i := 1; i <= targetCount; i++ {
		wg.Add(1)

		go func(idx int) {
			defer wg.Done()

			// 1. 기본값 설정 (정상 데이터)
			targetChannel := "#스터디" // ⚠️ 본인의 실제 채널명으로 확인 필수
			isFailureCase := false

			// 2. 고의적인 에러 주입 (10번째 요청마다 이상한 채널로 설정)
			if idx%9 == 0 {
				targetChannel = "#ghost-channel-999" // 존재하지 않는 채널
				isFailureCase = true
			}

			// 3. 금액 랜덤 생성 (1,000 ~ 100,000원)
			randomAmount := (rand.Intn(100) + 1) * 1000

			reqBody := NotificationRequest{
				TemplateCode: "ORDER_COMPLETE",
				Recipient:    targetChannel,
				Args: map[string]interface{}{
					"orderId": fmt.Sprintf("ORD-%d", idx),
					"amount":  randomAmount,
				},
			}
			jsonData, _ := json.Marshal(reqBody)

			// 4. 전송 (지연 없음)
			resp, err := http.Post("http://localhost:8080/api/v1/notifications", "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				fmt.Printf("❌ 요청 실패 [%d]: %v\n", idx, err)
				return
			}
			defer resp.Body.Close()

			// 5. 로그 출력 (불량 데이터는 눈에 띄게 표시)
			if isFailureCase {
				fmt.Printf("💀 [불량 데이터 발송] ID: %d | 채널: %s (실패 유도)\n", idx, targetChannel)
			} else if idx%1000 == 0 {
				fmt.Printf("🌊 전송 중... %d / %d\n", idx, targetCount)
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(startTime)
	fmt.Printf("✅ %d개 전송 완료! 걸린 시간: %s\n", targetCount, elapsed)
}
