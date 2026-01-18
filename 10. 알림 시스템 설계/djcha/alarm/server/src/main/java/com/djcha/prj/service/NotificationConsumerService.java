package com.djcha.prj.service;

import com.djcha.prj.entity.NotificationLog;
import com.djcha.prj.entity.NotificationTemplate;
import com.djcha.prj.repository.NotificationLogRepository;
import com.djcha.prj.repository.NotificationTemplateRepository;
import com.djcha.prj.slack.SlackSenderClient;
import com.google.gson.Gson;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.text.StringSubstitutor;
import org.apache.kafka.clients.consumer.Consumer;
import org.apache.kafka.common.TopicPartition;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.kafka.support.KafkaHeaders;
import org.springframework.messaging.handler.annotation.Header;
import org.springframework.retry.annotation.Backoff;
import org.springframework.retry.annotation.Recover;
import org.springframework.retry.annotation.Retryable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.Collections;
import java.util.Map;

@Slf4j
@Service
@RequiredArgsConstructor
public class NotificationConsumerService {

    private final NotificationLogRepository logRepository;
    private final NotificationTemplateRepository templateRepository;
    private final SlackSenderClient slackSenderClient; // 실제 슬랙 전송 객체 (구현 생략)
    private final Gson gson = new Gson();

    @KafkaListener(topics = "alarm.req", groupId = "notification-group")
    public void consume(
            String logIdStr,
            @Header(KafkaHeaders.RECEIVED_PARTITION) int partition,
            @Header(KafkaHeaders.OFFSET) long offset,
            Consumer<?, ?> consumer // 3. 카프카 상태 확인용 컨슈머 객체 주입
    ) {
        // ---------------------------------------------------------
        // 3. 카프카 메시지 적재량(Lag) 확인 로직
        // ---------------------------------------------------------
        TopicPartition topicPartition = new TopicPartition("alarm.req", partition);
        Map<TopicPartition, Long> endOffsets = consumer.endOffsets(Collections.singletonList(topicPartition));
        long endOffset = endOffsets.get(topicPartition);
        long lag = endOffset - offset - 1; // (끝 번호) - (현재 처리중인 번호) - 1

        log.info("🔥 [처리중] ID: {} | Partition: {} | Offset: {} | 📦 남은 메시지(Lag): {} 개",
                logIdStr, partition, offset, lag);

        Long logId = Long.parseLong(logIdStr);
        processNotificationWithRetry(logId);
    }

    @Retryable(retryFor = { RuntimeException.class }, maxAttempts = 3, backoff = @Backoff(delay = 2000))
    @Transactional
    public void processNotificationWithRetry(Long logId) {
        NotificationLog notiLog = logRepository.findById(logId).orElseThrow();
        NotificationTemplate template = templateRepository.findById(notiLog.getTemplateCode()).orElseThrow();

        try {
            // 2. 강제 1초 지연 (Throttling)
            Thread.sleep(1000);

            Map<String, Object> argsMap = gson.fromJson(notiLog.getPayloadArgs(), Map.class);
            StringSubstitutor substitutor = new StringSubstitutor(argsMap);
            String message = substitutor.replace(template.getMessageFormat());

            // Slack 발송
            slackSenderClient.sendToSlack(notiLog.getRecipient(), message);

            // 성공 처리
            notiLog.setStatus("SENT");
            notiLog.setErrorMessage(null);
            logRepository.save(notiLog);

        } catch (Exception e) {
            notiLog.setRetryCount(notiLog.getRetryCount() + 1);
            logRepository.save(notiLog);
            log.error("발송 실패... 재시도 합니다.", e);
            throw new RuntimeException("Slack send failed");
        }
    }

    // -------------------------------------------------------------
    // 재시도 횟수 초과 시 실행되는 메서드 (Fallback)
    // -------------------------------------------------------------
    @Recover
    public void recover(RuntimeException e, Long logId) {
        log.error("최종 실패 ID: {}", logId);
        NotificationLog notiLog = logRepository.findById(logId).orElse(null);
        if (notiLog != null) {
            notiLog.setStatus("FAILED");
            notiLog.setErrorMessage(e.getMessage());
            logRepository.save(notiLog);
        }
    }
}
