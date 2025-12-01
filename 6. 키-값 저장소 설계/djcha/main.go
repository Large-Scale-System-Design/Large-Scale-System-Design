package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// 설정
const (
	NumFields     = 10000000 // 시뮬레이션할 필드 개수 (1억 개)
	FieldDataSize = 100      // 필드당 데이터 크기 (Byte)
)

// 머클 트리 노드 정의: 해시값, 자식 노드, 리프 여부 및 원본 데이터 인덱스 포함
type MerkleNode struct {
	Hash     string
	Left     *MerkleNode
	Right    *MerkleNode
	IsLeaf   bool
	FieldIdx int
}

// 비교할 원본 레코드 구조체
type Record struct {
	Fields []string
}

// SHA-256 알고리즘을 사용하여 데이터의 해시값을 생성하는 헬퍼 함수
func CalculateHash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// 1. 단순 비교 (Naive)
// 모든 데이터를 네트워크로 전송받아 하나씩 비교한다고 가정하는 방식
func NaiveSync(recA, recB Record) (int, int, []int) {
	comparisons := 0
	transferBytes := 0
	var diffIndices []int

	// 전체 필드 개수 * 크기만큼 네트워크 전송이 발생한다고 시뮬레이션
	transferBytes = len(recA.Fields) * FieldDataSize

	// 모든 필드를 처음부터 끝까지 순차적으로 비교 (O(N))
	for i := 0; i < len(recA.Fields); i++ {
		comparisons++
		if recA.Fields[i] != recB.Fields[i] {
			diffIndices = append(diffIndices, i)
		}
	}
	return comparisons, transferBytes, diffIndices
}

// 2. 머클 트리 (Merkle Tree) 빌드 함수
// 데이터 리스트를 받아 리프 노드부터 루트까지 트리를 구성
func BuildTree(fields []string) *MerkleNode {
	var nodes []*MerkleNode
	// 입력받은 모든 데이터를 리프 노드로 변환
	for i, data := range fields {
		nodes = append(nodes, &MerkleNode{
			Hash:     CalculateHash(data),
			IsLeaf:   true,
			FieldIdx: i,
		})
	}
	// 루트 노드 하나가 남을 때까지 레벨을 올리며 트리 구성
	for len(nodes) > 1 {
		var nextLevel []*MerkleNode
		for i := 0; i < len(nodes); i += 2 {
			left := nodes[i]
			var right *MerkleNode
			// 오른쪽 자식이 없으면 왼쪽 자식을 복제하여 균형 맞춤
			if i+1 < len(nodes) {
				right = nodes[i+1]
			} else {
				right = left
			}
			// 좌우 자식의 해시를 합쳐 부모 노드의 해시 생성
			parentHash := CalculateHash(left.Hash + right.Hash)
			nextLevel = append(nextLevel, &MerkleNode{
				Hash:     parentHash,
				Left:     left,
				Right:    right,
				IsLeaf:   false,
				FieldIdx: -1,
			})
		}
		nodes = nextLevel
	}
	return nodes[0] // 최종 루트 노드 반환
}

// 머클 트리 동기화 함수
// 루트부터 시작해 해시가 다른 경로만 찾아 내려가는 효율적 비교
func MerkleSync(nodeA, nodeB *MerkleNode, comparisons *int, transferBytes *int) []int {
	var diffIndices []int
	*comparisons++
	*transferBytes += 64 // 해시값(32byte hex string * 2)만 전송한다고 가정

	// 해시가 같다면 하위 데이터는 완벽히 동일하므로 탐색 중단 (핵심 최적화)
	if nodeA.Hash == nodeB.Hash {
		return diffIndices
	}

	// 해시가 다른데 리프 노드라면, 이곳이 변경된 데이터임
	if nodeA.IsLeaf && nodeB.IsLeaf {
		diffIndices = append(diffIndices, nodeA.FieldIdx)
		return diffIndices
	}

	// 내부 노드의 해시가 다르다면 자식 노드로 내려가 재귀 탐색
	if nodeA.Left != nil && nodeB.Left != nil {
		diffs := MerkleSync(nodeA.Left, nodeB.Left, comparisons, transferBytes)
		diffIndices = append(diffIndices, diffs...)
	}
	if nodeA.Right != nil && nodeB.Right != nil {
		diffs := MerkleSync(nodeA.Right, nodeB.Right, comparisons, transferBytes)
		diffIndices = append(diffIndices, diffs...)
	}

	return diffIndices
}

func main() {
	fmt.Printf("시뮬레이션 시작. 필드 개수: %d\n", NumFields)
	fmt.Println("--------------------------------------------------")

	// 1억 개의 더미 데이터 생성
	fieldsA := make([]string, NumFields)
	for i := 0; i < NumFields; i++ {
		fieldsA[i] = "CommonData"
	}
	recA := Record{Fields: fieldsA}

	// 데이터 복제 후 중간값 하나를 임의로 변경 (불일치 발생 시나리오)
	fieldsB := make([]string, NumFields)
	copy(fieldsB, fieldsA)
	fieldsB[NumFields/2] = "CHANGED"
	recB := Record{Fields: fieldsB}

	fmt.Println("데이터 메모리 로드 완료")
	fmt.Println("")

	// [1] 단순 필드 비교 방식 실행 및 측정
	startNaive := time.Now()
	naiveOps, naiveBytes, naiveIndices := NaiveSync(recA, recB)
	durNaive := time.Since(startNaive)

	fmt.Println("[1] 단순 필드 비교")
	fmt.Printf("   - 비교 연산 횟수 : %d\n", naiveOps)
	fmt.Printf("   - 네트워크 전송량: %d Bytes\n", naiveBytes)
	fmt.Printf("   - 소요 시간      : %v\n", durNaive)
	fmt.Printf("   - 찾은 인덱스    : %v\n", naiveIndices)

	// [2] 머클 트리 동기화 방식 실행 및 측정
	fmt.Println("\n[2] 머클 트리 동기화")
	startBuild := time.Now()
	rootA := BuildTree(recA.Fields)
	rootB := BuildTree(recB.Fields)
	durBuild := time.Since(startBuild) // 트리 생성(빌드) 시간 별도 측정

	startMerkle := time.Now()
	merkleOps := 0
	merkleBytes := 0
	merkleIndices := MerkleSync(rootA, rootB, &merkleOps, &merkleBytes)
	durMerkle := time.Since(startMerkle) // 실제 비교(동기화) 시간 측정

	fmt.Printf("   - 비교 연산 횟수 : %d\n", merkleOps)
	fmt.Printf("   - 네트워크 전송량: %d Bytes\n", merkleBytes)
	fmt.Printf("   - 트리 빌드 시간 : %v\n", durBuild)
	fmt.Printf("   - 동기화 소요 시간: %v\n", durMerkle)
	fmt.Printf("   - 찾은 인덱스    : %v\n", merkleIndices)

	// [3] 최종 결과 효율성 비교 출력
	fmt.Println("\n[3] 최종 결과 비교")

	// 단순 비교 대비 머클 트리의 효율성(배수) 계산
	ratioOps := float64(naiveOps) / float64(merkleOps)
	ratioBytes := float64(naiveBytes) / float64(merkleBytes)
	ratioTime := float64(durNaive.Nanoseconds()) / float64(durMerkle.Nanoseconds())
	if durMerkle.Nanoseconds() == 0 {
		ratioTime = 0
	}

	// 결과 표 출력
	fmt.Println("-------------------------------------------------------------------------------------")
	fmt.Printf("| %-18s | %-20s | %-20s | %-12s |\n", "구분", "단순 비교 (Naive)", "머클 트리 (Merkle)", "효율 (배수)")
	fmt.Println("-------------------------------------------------------------------------------------")
	fmt.Printf("| %-18s | %18d 회 | %18d 회 | %10.1f 배 |\n",
		"비교 연산 횟수", naiveOps, merkleOps, ratioOps)
	fmt.Printf("| %-18s | %18d Byte | %18d Byte | %10.1f 배 |\n",
		"네트워크 전송량", naiveBytes, merkleBytes, ratioBytes)
	fmt.Printf("| %-18s | %18s | %18s | %10.1f 배 |\n",
		"동기화 소요 시간", durNaive, durMerkle, ratioTime)
	fmt.Println("-------------------------------------------------------------------------------------")

}
