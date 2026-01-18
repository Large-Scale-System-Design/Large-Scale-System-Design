package sender

// Sender : 모든 알림 채널 구현체가 따라야 할 인터페이스
type Sender interface {
	// properties 맵을 통해 채널별로 유연한 데이터 전달
	Send(recipient string, props map[string]string) error
}
