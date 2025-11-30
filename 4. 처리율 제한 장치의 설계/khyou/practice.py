import redis
import time
from flask import Flask, request, jsonify

app = Flask(__name__)

# 전역 변수로 프로세스 이름 저장
PROCESS_NAME = None

# Redis 연결 (모든 프로세스가 같은 Redis 사용)
redis_client = redis.Redis(host='localhost', port=6379, db=0)

# rate limit 설정
MAX_REQUESTS = 10       # 허용 요청 수
WINDOW_SIZE = 60       # 초 단위

lua_script = """
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])

-- 오래된 항목 제거
redis.call('ZREMRANGEBYSCORE', key, 0, now - window)

-- 현재 개수 확인
local count = redis.call('ZCARD', key)

if count < limit then
    redis.call('ZADD', key, now, tostring(now))
    redis.call('EXPIRE', key, window)
    return 1 -- 허용
else
    return 0 -- 거절
end
"""

rate_limiter = redis_client.register_script(lua_script)

def is_request_allowed(user_id):
    key = f"rate_limit:{user_id}"
    now = time.time()
    allowed = rate_limiter(keys=[key], args=[now, WINDOW_SIZE, MAX_REQUESTS])  # 윈도우 10초, 최대 5회
    return allowed == 1

@app.route("/")
def index():
    user_id = request.args.get("userId")
    if not user_id:
        return jsonify({"error": "userId query parameter is required"}), 400

    allowed = is_request_allowed(user_id)

    if allowed:
        return jsonify({
            "status": "ok",
            "userId": user_id,
            "message": "Request accepted",
            "processName": PROCESS_NAME
        }), 200
    else:
        return jsonify({
            "status": "rate_limited",
            "userId": user_id,
            "message": "Too many requests, try again later",
            "processName": PROCESS_NAME
        }), 429


if __name__ == "__main__":
    import sys
    if len(sys.argv) < 2:
        print("Usage: python rate_limiter.py <port>")
        exit(1)

    port = int(sys.argv[1])
    PROCESS_NAME = "myapp_" + str(port)
    app.run(host="127.0.0.1", port=port)