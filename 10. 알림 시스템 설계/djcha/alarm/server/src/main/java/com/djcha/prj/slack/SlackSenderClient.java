package com.djcha.prj.slack;

import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.MediaType;
import org.springframework.stereotype.Component;
import org.springframework.web.reactive.function.client.WebClient;
import org.springframework.web.reactive.function.client.WebClientResponseException;

import java.util.Map;

@Slf4j
@Component
@RequiredArgsConstructor
public class SlackSenderClient {

    @Value("${slack.api-url}")
    private String slackApiUrl;

    @Value("${slack.bot-token}")
    private String botToken;

    private final WebClient webClient;

    /**
     * 슬랙 메시지 전송 (WebClient 동기 방식)
     */
    public void sendToSlack(String channel, String message) {
        try {
            // 1. 요청 바디 생성
            SlackRequest requestBody = new SlackRequest(channel, message);

            // 2. WebClient 호출 및 동기 대기 (.block())
            Map response = webClient.post()
                    .uri(slackApiUrl)
                    .header("Authorization", "Bearer " + botToken)
                    .contentType(MediaType.APPLICATION_JSON)
                    .bodyValue(requestBody)
                    .retrieve() // 응답 추출 시작
                    .bodyToMono(Map.class) // 결과를 Map으로 변환
                    .block(); // ⛔ 동기 실행: 응답이 올 때까지 여기서 대기합니다.

            // 3. Slack API 논리적 오류 확인 ("ok": false 인 경우)
            if (response != null && Boolean.FALSE.equals(response.get("ok"))) {
                String error = (String) response.get("error");
                log.error("Slack API Logic Error: {}", error);
                throw new RuntimeException("Slack API responded with error: " + error);
            }

            log.info("Slack message sent to {}", channel);

        } catch (WebClientResponseException e) {
            // HTTP 4xx, 5xx 에러 발생 시 여기로 잡힘
            log.error("Slack HTTP Error: {} - {}", e.getStatusCode(), e.getResponseBodyAsString());
            // Retryable이 동작하도록 RuntimeException 재발생
            throw new RuntimeException("Slack HTTP Error", e);

        } catch (Exception e) {
            // 타임아웃이나 기타 네트워크 오류
            log.error("Unknown Error while sending to Slack: {}", e.getMessage());
            throw new RuntimeException("Slack Client Error", e);
        }
    }

    // 슬랙 전송용 DTO
    @Getter
    @AllArgsConstructor
    static class SlackRequest {
        private String channel;
        private String text;
    }
}