package com.djcha.prj.entity;

import jakarta.persistence.*;
import lombok.*;
import org.springframework.data.annotation.CreatedDate;
import org.springframework.data.annotation.LastModifiedDate;
import org.springframework.data.jpa.domain.support.AuditingEntityListener;
import java.time.LocalDateTime;

@Entity
@Table(name = "notification_log")
@Getter @Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
@EntityListeners(AuditingEntityListener.class)
public class NotificationLog {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id; // PK

    @Column(name = "template_code", nullable = false)
    private String templateCode; // 연관관계를 맺지 않고 코드로만 관리 (로그의 독립성 보장)

    @Column(nullable = false, length = 100)
    private String recipient; // 수신자 (Slack 채널 등)

    @Lob
    @Column(name = "payload_args")
    private String payloadArgs; // 템플릿 변수 데이터 (JSON String으로 저장)

    @Column(length = 20)
    private String status; // PENDING, SENT, FAILED

    @Column(name = "retry_count")
    private int retryCount = 0; // 재시도 횟수

    @Lob
    @Column(name = "error_message")
    private String errorMessage; // 실패 원인

    @CreatedDate
    @Column(name = "created_at", updatable = false)
    private LocalDateTime createdAt;

    @LastModifiedDate
    @Column(name = "updated_at")
    private LocalDateTime updatedAt;
}
