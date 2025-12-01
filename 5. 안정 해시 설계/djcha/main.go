package main

import (
	"crypto/sha256"   // SHA-256 해시 알고리즘을 사용하기 위한 패키지
	"encoding/binary" // "encoding/binary" : 바이트(byte) 배열을 숫자(uint32 등)로 변환하거나 그 반대의 작업을 수행하기 위한 패키지
	"fmt"
	"slices"
	"sort"
	"strconv" // 문자열을 숫자로 바뀌기 위한 패키지
)

type ConsistentHash struct {
	hashFunc  func(data []byte) uint32 // 해시 알고리즘 함수 타입 정의
	replicas  int                      // 가상 노드(복제본)의 개수를 저장할 정수(int) 타입 필드
	rings     []uint32                 // 해시 링을 나타낼 동적 배열 (unsigned int 4byte) (0 ~ 42.9억 정도까지)
	serverMap map[uint32]string        // 해시 값에 따른 서버 이름 Map
}

// 생성자
func NewConsistentHash(replicas int) *ConsistentHash { // 구조체의 포인터를 반환 *ConsistentHash (구초제를 복사하지 않고 메모리 주소를 전달)
	// '&ConsistentHash{ ... }' 는 'ConsistentHash' 구조체의 인스턴스(실체)를 메모리에 생성하고, 그 메모리 주소(&)를 'ch' 변수에 할당
	ch := &ConsistentHash{
		replicas:  replicas,
		serverMap: make(map[uint32]string),
		rings:     make([]uint32, 0), // 길이가 0인 비어있는 'uint32' 슬라이스를 생성합니다.
		hashFunc: func(data []byte) uint32 {
			hash := sha256.Sum256(data) // 입력값에 대한 32 Byte (ex. 2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824)
			// 앞의 4byte 까지만 자르고 -> [byte1, byte2, byte3, byte4] -> byte1 을 제일 큰자리수부터 해서 uint32 로 채움 (이것이 바로 빅 엔디안 방식)
			// 이때, hash 값에서 문자 또는 숫자는 16진수라고 보면 됨
			return binary.BigEndian.Uint32(hash[:4])
		},
	}

	return ch
}

// 서버에 대한 가상 노드 추가 함수
// ConsistentHash 구조체 내부의 함수라고 생각하면 됨. ConsistenHash.Add
func (ch *ConsistentHash) Add(server string) {
	// 서버 1대를 기준으로 생성자에서 설정한 replicas 수만큼, 가상노드를 생성
	for i := 0; i < ch.replicas; i++ {
		// 가상 노드 키 생성 (ex. "Server-A-1", "Server-A-2")
		virtualKey := server + "-" + strconv.Itoa(i)

		// 가상 노드의 키에 대한 Hash 값 취득 (링 위의 값)
		hash := ch.hashFunc([]byte(virtualKey))

		// 가상 노드 키를 키 목록에 추가
		ch.rings = append(ch.rings, hash)

		// 가상 노드 키 키값(hash)에 대한 서버 이름 저장
		ch.serverMap[hash] = server
	}

	// 키 목록 오름차순 정렬 함수
	slices.Sort(ch.rings)
}

// 주어진 세션 ID가 어떤 서버에 매핑되어있는지 확인하는 함수
func (ch *ConsistentHash) Get(sessionId string) (uint32, string) {
	if len(ch.rings) == 0 {
		return 0, ""
	}

	// 입력된 서버 키에 대한 hash key 값
	hash := ch.hashFunc([]byte(sessionId))

	// 세션 ID 에 대한 hash key 값 보다 크거나 같은 서버의 key 값에 대한 index 찾기 (sort.Search 를 이용해서 이진 탐색)
	idx := sort.Search(
		len(ch.rings),
		func(i int) bool {
			return ch.rings[i] >= hash
		},
	)

	// hash key 값보다 큰 서버의 key 가 없을 경우(index 가 rings의 크기와 동일할 경우), 가장 첫번째 서버(index=0)에 매핑
	if idx == len(ch.rings) {
		idx = 0
	}

	// 찾은 index 를 이용해서 서버의 key 를 찾아서 -> 서버 Map 에서 서버 이름 찾기
	return ch.rings[idx], ch.serverMap[ch.rings[idx]]
}

// 테스트 코드
func main() {
	// 1. 서버 1대당 가상 노드를 1개로 설정한 안정 해시 링 생성
	ch := NewConsistentHash(10000)

	// 2. 초기 서버 3대를 피터지는 해시 링 전장에 참여
	initialServers := []string{"Server-A", "Server-B", "Server-C"}
	for _, server := range initialServers { // _: index 인데 _를 입력함으로써 for 문에서 index를 사용하지 않겠다는 암묵적인 의미
		ch.Add(server) // 서버에 대한 가상 노드 추가 함수 수행
	}

	// 3. 테스트할 세션 ID 목록 정의
	sessionIDs := []string{
		"session-id-1-yjkang",
		"session-id-2-djcha",
		"session-id-3-khyou",
		"session-id-4-swma",
	}

	fmt.Println("--- 1. 서버 3대 (A, B, C) ---")

	// 세션ID => 서버 이름에 대한 Map 생성
	initialMap := make(map[string]string)
	for _, sessionId := range sessionIDs {
		sessionIdKey := ch.hashFunc([]byte(sessionId))
		hash, server := ch.Get(sessionId)                                           // 세션 ID에 대한 서버 이름 취득
		initialMap[sessionId] = server                                              // Map에 저장
		fmt.Printf("[%s][%d] -> [%s][%d]\n", sessionId, sessionIdKey, server, hash) // Map 에 출력
	}

	// 서버 1대 추가'Server-D' 추가
	ch.Add("Server-D")
	fmt.Println("\n----------------------------------------")
	fmt.Println("--- 2. 'Server-D' 1대 추가 (A, B, C, D) ---")
	fmt.Println("----------------------------------------")

	cacheMissCount := 0 // 캐시 미스 카운트
	for _, sessionId := range sessionIDs {
		sessionIdKey := ch.hashFunc([]byte(sessionId))
		hash, newServer := ch.Get(sessionId) // 서버 1대 추가 후 세션 ID에 대한 서버 이름 취득
		oldServer := initialMap[sessionId]   // 이전에 캐싱된 세션 ID에 대한 서버 이름 취득

		// 두 서버 이름이 같지 않으면 CACHE MISS
		status := "OOOOO CACHE HIT OOOOO" // 기본 상태는 'HIT'
		if newServer != oldServer {
			status = "XXXXX  CACHE MISS  XXXXX"
			cacheMissCount++ // 캐시 미스 카운트 증가
		}

		fmt.Printf("[%s][%d] -> [%s][%d] (이전: %s) [%s]\n", sessionId, sessionIdKey, newServer, hash, oldServer, status)
	}

	fmt.Printf("\n총 %d개 세션 중 %d개 캐시 미스 발생.\n", len(sessionIDs), cacheMissCount)
}
