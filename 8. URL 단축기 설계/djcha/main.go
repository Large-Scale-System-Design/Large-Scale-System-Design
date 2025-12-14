package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/bwmarrin/snowflake"
)

// ---------------------------
// 1. Base62 Logic
// ---------------------------
const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Base62Encode: Snowflake ID(int64)를 Base62 문자열로 변환
func Base62Encode(num int64) string {
	if num == 0 {
		return string(base62Chars[0])
	}

	chars := []byte{}
	base := int64(len(base62Chars))

	for num > 0 {
		rem := num % base
		chars = append(chars, base62Chars[rem])
		num = num / base
	}

	// 역순으로 뒤집어야 함
	for i, j := 0, len(chars)-1; i < j; i, j = i+1, j-1 {
		chars[i], chars[j] = chars[j], chars[i]
	}

	return string(chars)
}

// ---------------------------
// 2. Storage (In-Memory)
// ---------------------------
type URLMap struct {
	sync.RWMutex
	shortToLong map[string]string
}

var store = URLMap{
	shortToLong: make(map[string]string),
}

// ---------------------------
// 3. Server Logic
// ---------------------------
type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
}

func main() {
	// Snowflake Node 초기화 (Node 번호: 1)
	node, err := snowflake.NewNode(1)
	if err != nil {
		log.Fatalf("Snowflake node 생성 실패: %v", err)
		return
	}

	// 핸들러: URL 단축 요청 (POST)
	http.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST 메소드만 허용됩니다.", http.StatusMethodNotAllowed)
			return
		}

		var req ShortenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "잘못된 JSON 형식입니다.", http.StatusBadRequest)
			return
		}

		// URL 유효성 간단 체크 및 http prefix 추가
		targetURL := req.URL
		if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
			targetURL = "http://" + targetURL
		}

		// 1. Snowflake ID 생성
		id := node.Generate().Int64()

		// 2. Base62 인코딩
		shortCode := Base62Encode(id)

		// 3. 저장 (메모리)
		store.Lock()
		store.shortToLong[shortCode] = targetURL
		store.Unlock()

		// 4. 결과 반환
		resp := ShortenResponse{
			OriginalURL: targetURL,
			ShortURL:    fmt.Sprintf("http://localhost:8080/%s", shortCode),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// 핸들러: 리다이렉트 (GET /{shortCode})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 루트 경로("/") 요청은 무시 혹은 안내
		if r.URL.Path == "/" {
			fmt.Fprint(w, "URL Shortener가 실행 중입니다. /shorten 으로 POST 요청을 보내세요.")
			return
		}

		// Path에서 shortCode 추출 (예: /AbCd12 -> AbCd12)
		shortCode := strings.TrimPrefix(r.URL.Path, "/")

		store.RLock()
		originalURL, exists := store.shortToLong[shortCode]
		store.RUnlock()

		if !exists {
			http.NotFound(w, r)
			return
		}

		// 302 Found 리다이렉트
		http.Redirect(w, r, originalURL, http.StatusFound)
	})

	fmt.Println("서버 시작: http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
