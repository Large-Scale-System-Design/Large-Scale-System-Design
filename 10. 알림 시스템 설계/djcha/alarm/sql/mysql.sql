-- 1. 템플릿 정의 테이블
CREATE TABLE notification_template (
    code VARCHAR(50) PRIMARY KEY, -- 예: ORDER_COMPLETE
    title VARCHAR(100) NOT NULL,
    message_format TEXT NOT NULL, -- 예: "주문번호 ${orderId} 처리가 완료되었습니다."
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 2. 알림 발송 이력 테이블
CREATE TABLE notification_log (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    template_code VARCHAR(50) NOT NULL,
    recipient VARCHAR(100) NOT NULL,     -- 슬랙 채널명 or Webhook URL Key
    payload_args JSON,                   -- 치환할 변수들 (JSON)
    status VARCHAR(20) DEFAULT 'PENDING',-- PENDING, SENT, FAILED
    retry_count INT DEFAULT 0,
    error_message TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);


INSERT INTO notification_template (code, title, message_format) 
VALUES ('ORDER_COMPLETE', '주문 완료 알림', '✅ 주문이 완료되었습니다.\n주문번호: *${orderId}*\n결제금액: *${amount}원*');
