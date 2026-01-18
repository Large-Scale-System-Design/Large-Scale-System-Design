package com.djcha.prj.controller;

import com.djcha.prj.dto.NotificationRequest;
import com.djcha.prj.service.NotificationProducerService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api/v1/notifications")
@RequiredArgsConstructor
public class NotificationController {
    private final NotificationProducerService notificationProducerService;

    @PostMapping
    public ResponseEntity<String> send(@RequestBody NotificationRequest request) {
        Long logId = notificationProducerService.registerNotification(request);
        return ResponseEntity.ok("Notification Registered. (ID: " + logId + ")");
    }
}
