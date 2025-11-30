import hashlib
import bisect

class ConsistentHashRing:
    def __init__(self, replicas=10):
        # replicas: 각 노드를 가상 노드로 몇 개 복제할지 (부하 균형용)
        self.replicas = replicas

        # ring: 해시 값(key) -> 실제 노드 이름(node) 매핑
        self.ring = {}

        # sorted_keys: 해시 링의 모든 key를 정렬된 상태로 저장 (이진 탐색용)
        self.sorted_keys = []

    def _hash(self, key):
        # 주어진 문자열(key)을 MD5 해시로 변환 → 16진수 → 정수형으로 반환
        return int(hashlib.md5(key.encode()).hexdigest(), 16)

    def add_node(self, node):
        # 실제 노드를 추가할 때, replicas(예: 100개) 만큼 가상 노드 생성
        for i in range(self.replicas):
            # 노드 이름 + 인덱스 조합으로 가상 노드 구분
            key = f"{node}:{i}"

            # 해당 key의 해시값 계산
            h = self._hash(key)

            # 해시 링에 등록 (해시값 -> 노드)
            self.ring[h] = node

            # 해시 값을 정렬 리스트에 삽입 (bisect: 이진 탐색 기반 정렬 삽입)
            bisect.insort(self.sorted_keys, h)

    def remove_node(self, node):
        # 노드를 제거할 때는 가상 노드들도 모두 제거
        for i in range(self.replicas):
            key = f"{node}:{i}"
            h = self._hash(key)
            del self.ring[h]
            self.sorted_keys.remove(h)

    def get_node(self, key):
        # 링이 비어 있으면 None 반환
        if not self.ring:
            return None

        # 요청 키의 해시값 계산
        h = self._hash(key)

        # sorted_keys에서 h보다 큰 첫 번째 위치를 찾음 (이진 탐색)
        # % len(...) 을 해서 해시 링의 끝을 넘어가면 처음으로 순환되게 함
        idx = bisect.bisect(self.sorted_keys, h) % len(self.sorted_keys)

        # 해당 위치의 노드를 반환
        return self.ring[self.sorted_keys[idx]]