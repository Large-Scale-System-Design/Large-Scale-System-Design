package com.djcha.prj.service;

import com.djcha.prj.entity.NotificationLog;
import com.djcha.prj.repository.NotificationLogRepository;
import com.djcha.prj.dto.NotificationRequest;
import com.djcha.prj.repository.NotificationTemplateRepository;
import com.google.gson.Gson;
import lombok.RequiredArgsConstructor;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

@Service
@RequiredArgsConstructor
public class NotificationProducerService {

    private final NotificationLogRepository logRepository;
    private final NotificationTemplateRepository templateRepository;
    private final KafkaTemplate<String, String> kafkaTemplate;
    private final Gson gson = new Gson();

    @Transactional
    public Long registerNotification(NotificationRequest request) {
        // 1. 템플릿 존재 여부 확인
        if (!templateRepository.existsById(request.getTemplateCode())) {
            throw new IllegalArgumentException("Unknown Template Code");
        }

        // 2. DB에 PENDING 상태로 저장
        NotificationLog log = new NotificationLog();
        log.setTemplateCode(request.getTemplateCode());
        log.setRecipient(request.getRecipient());
        log.setPayloadArgs(gson.toJson(request.getArgs())); // Map -> JSON String
        log.setStatus("PENDING");

        NotificationLog savedLog = logRepository.save(log);

        // 3. Kafka로 'Log ID'만 전송 (가볍게)
        kafkaTemplate.send("alarm.req", String.valueOf(savedLog.getId()));

        return savedLog.getId();
    }
}
