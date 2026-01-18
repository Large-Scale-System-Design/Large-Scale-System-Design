package protocol

// 상수 정의
const (
	ChannelEmail = "email"
	// 향후 추가: ChannelSMS = "sms"
	
	QueuePrefix = "queue:"
)

// 범용 알림 페이로드
type NotificationPayload struct {
	ID         string            `json:"id"`
	Channel    string            `json:"channel"`    // 라우팅 기준 (email, sms...)
	Recipient  string            `json:"recipient"`  // 수신처 (이메일 주소, 전화번호...)
	Properties map[string]string `json:"properties"` // 채널별 상세 데이터 (Subject, Body 등)
	RetryCount int               `json:"retry_count"`
}
