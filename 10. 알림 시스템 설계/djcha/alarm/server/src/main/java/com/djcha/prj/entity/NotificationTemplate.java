package com.djcha.prj.entity;

import jakarta.persistence.*;
import lombok.*;
import org.springframework.data.annotation.CreatedDate;
import org.springframework.data.jpa.domain.support.AuditingEntityListener;
import java.time.LocalDateTime;

@Entity
@Table(name = "notification_template")
@Getter @Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
@EntityListeners(AuditingEntityListener.class) // 생성 시간 자동 관리
public class NotificationTemplate {

    @Id
    @Column(length = 50)
    private String code; // 템플릿 코드 (PK, 예: ORDER_COMPLETE)

    @Column(nullable = false, length = 100)
    private String title; // 템플릿 설명

    @Lob // 대용량 텍스트 처리
    @Column(name = "message_format", nullable = false)
    private String messageFormat; // 치환자가 포함된 메시지 본문

    @CreatedDate
    @Column(name = "created_at", updatable = false)
    private LocalDateTime createdAt;
}
