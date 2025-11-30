안정 해시 구현
안정 해시를 직접 만들어본다.

```` python
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
````
위 코드를 아래와 같이 스크립트를 작성하여 실행한다.

```` python
ring = ConsistentHashRing()

servers = ["A", "B", "C"]
for s in servers:
    ring.add_node(s)

users = [f"user{i}" for i in range(1, 21)]

# 초기 분배
print("=== Initial distribution ===")
initial_node = list()
for u in users:
    node = ring.get_node(u)
    initial_node.append((u, node))
    print(u, "->", ring.get_node(node))

# 서버 증설
ring.add_node("D")
print("\n=== After adding server D ===")
after_node = list()
for u in users:
    node = ring.get_node(u)
    after_node.append((u, node))
    print(u, "->", ring.get_node(u))

diff_node = dict()
for i in range(len(initial_node)):
    if initial_node[i] != after_node[i]:
        diff_node[initial_node[i][0]] = [initial_node[i][1], after_node[i][1]]

print("\ndiff-node-user: " )
for k, n in diff_node.items():
    print(k, '-->', n)
````

위 스크립트를 실행시키면 아래와 같이 출력된다.

````
=== Initial distribution ===
user1 -> B
user2 -> C
user3 -> B
user4 -> B
user5 -> C
user6 -> B
user7 -> B
user8 -> C
user9 -> C
user10 -> B
user11 -> C
user12 -> C
user13 -> C
user14 -> C
user15 -> C
user16 -> C
user17 -> B
user18 -> C
user19 -> B
user20 -> C

=== After adding server D ===
user1 -> B
user2 -> C
user3 -> B
user4 -> D
user5 -> C
user6 -> B
user7 -> D
user8 -> A
user9 -> C
user10 -> D
user11 -> C
user12 -> A
user13 -> A
user14 -> C
user15 -> A
user16 -> C
user17 -> B
user18 -> D
user19 -> B
user20 -> A

diff-node-user: 
user4 --> ['B', 'D']
user7 --> ['B', 'D']
user10 --> ['B', 'D']
user18 --> ['C', 'D']

Process finished with exit code 0
````

replicas 를 늘리면 Server에 User가 적절히 분배되지만, Server가 바뀌는 User가 늘어나게 된다. 반대로 replicas 를 줄이면 Server에 User가 적절히 분배되지는 않지만, Server가 바뀌는 User가 줄어든다.