package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ==========================================
// 1. Snowflake 설정 상수 (총 64비트 설계)
// ==========================================
const (
	// Epoch: 기준 시간 (사용자 정의 시작일)
	// 이 값이 너무 옛날이면 타임스탬프 공간(41비트) 낭비가 심하므로,
	// 보통 서비스 오픈 시점이나 현재 시점(예: 2024-01-01)을 기준으로 잡습니다.
	epoch = int64(1704067200000)

	// 각 구역이 차지할 비트 수 (총합 1 + 41 + 10 + 12 = 64비트)
	// - Sign Bit(1): 양수 표현을 위해 0으로 고정 (코드엔 상수 선언 불필요)
	// - Timestamp(41): 밀리초 단위 시간 기록
	nodeBits = uint(10) // Node ID: 서버/인스턴스 식별 (2^10 = 1024개 노드 가능)
	stepBits = uint(12) // Sequence: 같은 밀리초 내 순서 (2^12 = 4096개 ID/ms 가능)

	// 각 구역의 최대값 (비트 마스킹용)
	// 예: nodeMax는 10비트가 모두 1인 값(1023)이 됨.
	nodeMax = int64(-1 ^ (-1 << nodeBits))
	stepMax = int64(-1 ^ (-1 << stepBits))

	// 비트 이동(Shift) 량 계산
	// ID 구조: [Sign:1][Timestamp:41][Node:10][Sequence:12]
	// 따라서 Timestamp는 왼쪽으로 (10+12) = 22칸 이동해야 함.
	timeShift = nodeBits + stepBits // 22
	// Node ID는 왼쪽으로 12칸 이동해야 함.
	nodeShift = stepBits // 12
)

// ==========================================
// 2. Node 구조체 (ID 생성기)
// ==========================================
type Node struct {
	mu        sync.Mutex // 동시성 관리를 위한 락 (여러 고루틴 접근 방어)
	timestamp int64      // 마지막으로 ID를 생성한 시간(밀리초)
	nodeID    int64      // 이 서버의 고유 번호 (0 ~ 1023)
	step      int64      // 같은 밀리초 내에서의 순번 (0 ~ 4095)
}

// NewNode: 생성기 초기화 함수
func NewNode(nodeID int64) (*Node, error) {
	// 노드 ID가 허용 범위를 넘는지 체크 (0 ~ 1023)
	if nodeID < 0 || nodeID > nodeMax {
		return nil, fmt.Errorf("node ID must be between 0 and %d", nodeMax)
	}

	// 초기화된 Node 반환
	return &Node{
		timestamp: 0,
		nodeID:    nodeID,
		step:      0,
	}, nil
}

// Generate: 고유 ID 생성 (핵심 로직)
func (n *Node) Generate() int64 {
	// 1. 락을 걸어 다른 고루틴이 동시에 상태를 변경하지 못하게 함
	n.mu.Lock()
	defer n.mu.Unlock() // 함수 종료 시 락 해제

	// 2. 현재 시간 가져오기 (밀리초)
	now := time.Now().UnixMilli()

	// 3. 시간 역전 체크 (시스템 시계 오류 등)
	if now < n.timestamp {
		panic(errors.New("clock moved backwards"))
	}

	// 4. 같은 밀리초 내에 요청이 들어온 경우 (충돌 방지 로직)
	if now == n.timestamp {
		// 시퀀스(step)를 1 증가시킴
		n.step = (n.step + 1) & stepMax

		// 시퀀스가 꽉 찼다면 (4096번째 요청), 다음 밀리초가 될 때까지 대기
		if n.step == 0 {
			for now <= n.timestamp {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		// 새로운 밀리초(시간이 흐름)라면 시퀀스를 0으로 초기화
		n.step = 0
	}

	// 5. 마지막 생성 시간 업데이트
	n.timestamp = now

	// 6. 비트 연산으로 최종 ID 조립
	// (현재시간 - 기준시간)을 왼쪽으로 22비트 밀고
	// NodeID를 왼쪽으로 12비트 밀고
	// Step을 마지막에 붙임 (OR 연산)
	return ((now - epoch) << timeShift) | (n.nodeID << nodeShift) | n.step
}

// ==========================================
// 3. 시각화 함수 (비트 구조 분석용)
// ==========================================
func VisualizeID(caseName string, id int64) {
	// --- [비트 역추적 로직] ---

	// 1. Sign Bit (1 bit): 맨 앞 1비트 추출
	// (항상 0이어야 함, 양수)
	sign := (id >> 63) & 1

	// 2. Timestamp (41 bits): 상위 41비트 추출
	// 실제 의미: (생성시간 - Epoch)
	timeRaw := (id >> timeShift) & ((1 << 41) - 1)

	// 3. Node ID (10 bits): 중간 10비트 추출
	// 실제 의미: ID를 생성한 서버 번호
	nodeVal := (id >> nodeShift) & nodeMax

	// 4. Sequence (12 bits): 하위 12비트 추출
	// 실제 의미: 동시간대 생성 순서
	stepVal := id & stepMax

	// --- [출력] ---
	fmt.Printf("\n[%s]\n", caseName)
	fmt.Printf("최종 ID (10진수): %d\n", id)

	// 헤더 출력
	fmt.Println(strings.Repeat("-", 110))
	// 각 필드 설명 헤더
	fmt.Printf("| %-12s | %-48s | %-16s | %-16s |\n",
		"Sign(1)",       // 부호 비트
		"Timestamp(41)", // 시간값 (Epoch 이후 경과 시간)
		"Node ID(10)",   // 서버 번호
		"Sequence(12)")  // 일련 번호
	fmt.Println(strings.Repeat("-", 110))

	// 데이터 출력: 2진수(10진수) 형태로 출력
	// %01b: 2진수 1자리로 표현 (빈자리는 0)
	// %041b: 2진수 41자리로 표현
	fmt.Printf("| %01b(%d)         | %041b(%d)      | %010b(%d)        | %012b(%d)         |\n",
		sign, sign,
		timeRaw, timeRaw,
		nodeVal, nodeVal,
		stepVal, stepVal)

	fmt.Println(strings.Repeat("-", 110))
}

func main() {
	// [시나리오 1] 1번 서버(Node 1)에서 ID를 '연속' 생성
	// 목표: Timestamp는 같고, Sequence(맨 뒤)만 1씩 증가하는지 확인
	node1, _ := NewNode(1)
	fmt.Println("=== TEST 1: 연속 생성 (Sequence 증가 확인) ===")
	for i := 0; i < 3; i++ {
		id := node1.Generate()
		VisualizeID(fmt.Sprintf("Node 1 - 반복회차 %d", i+1), id)
	}

	// [시나리오 2] 1023번 서버(Node 1023)에서 생성
	// 목표: Node ID 영역(중간)이 꽉 찬 비트(1111111111)로 나오는지 확인
	nodeMaxVal, _ := NewNode(1023)
	fmt.Println("\n=== TEST 2: 다른 서버 (Node ID 변경 확인) ===")
	id2 := nodeMaxVal.Generate()
	VisualizeID("Node 1023 - 단건 생성", id2)

	// [시나리오 3] 시간을 약간(0.1초) 두고 생성
	// 목표: Timestamp 영역(앞쪽)의 비트값이 변하는지 확인
	time.Sleep(100 * time.Millisecond)
	fmt.Println("\n=== TEST 3: 시간 경과 후 (Timestamp 변경 확인) ===")
	id3 := node1.Generate()
	VisualizeID("Node 1 - 0.1초 후 생성", id3)
}
