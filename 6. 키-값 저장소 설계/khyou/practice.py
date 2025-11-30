from flask import Flask, request, jsonify
import redis
import threading
import time

app = Flask(__name__)

# Redis 노드 목록
REDIS_NODES = [
    redis.Redis(host='127.0.0.1', port=6379),
    redis.Redis(host='127.0.0.1', port=6380),
    redis.Redis(host='127.0.0.1', port=6381),
]

MAJORITY = 2  # 3대 중 2대 성공하면 commit 성공


# -------------------------------------------------------------------
# Write: 모든 노드에 SET 요청 → Majority 성공 시 OK
# -------------------------------------------------------------------
@app.route("/set", methods=["POST"])
def set_value():
    data = request.json
    key = data["key"]
    value = data["value"]

    success_count = 0

    for node in REDIS_NODES:
        try:
            node.set(key, value)
            success_count += 1
        except Exception as e:
            print(f"[WARN] SET 실패: {e}")

    if success_count >= MAJORITY:
        return jsonify({"status": "OK", "success_nodes": success_count})
    else:
        return jsonify({"status": "FAIL", "success_nodes": success_count}), 500


# -------------------------------------------------------------------
# Read: 모든 노드에 GET 요청 후 가장 빠르게 응답한 값 반환
# -------------------------------------------------------------------
@app.route("/get", methods=["GET"])
def get_value():
    key = request.args.get("key")

    responses = []
    threads = []

    def fetch(node):
        try:
            val = node.get(key)
            if val is not None:
                responses.append(val.decode())
        except:
            pass

    # 병렬로 GET 요청
    for node in REDIS_NODES:
        t = threading.Thread(target=fetch, args=(node,))
        t.start()
        threads.append(t)

    # 응답을 300ms만 기다림 → 가장 빠른 응답만 사용
    start = time.time()
    while time.time() - start < 0.3:
        if responses:
            break
        time.sleep(0.01)

    # fallback: 모든 스레드 기다리기
    for t in threads:
        t.join(timeout=0.1)

    if responses:
        return jsonify({"status": "OK", "value": responses[0]})
    else:
        return jsonify({"status": "NOT_FOUND"}), 404


if __name__ == "__main__":
    app.run(port=5000)