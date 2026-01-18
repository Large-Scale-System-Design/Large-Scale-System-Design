package com.djcha.prj.dto;

import java.util.Map;

public class NotificationRequest {

    private String templateCode;
    private String recipient;
    private Map<String, Object> args;

    // ✅ [핵심] 이 기본 생성자가 없으면 Jackson이 에러를 냅니다!
    public NotificationRequest() {
    }

    public NotificationRequest(String templateCode, String recipient, Map<String, Object> args) {
        this.templateCode = templateCode;
        this.recipient = recipient;
        this.args = args;
    }

    // Getter & Setter
    public String getTemplateCode() { return templateCode; }
    public void setTemplateCode(String templateCode) { this.templateCode = templateCode; }

    public String getRecipient() { return recipient; }
    public void setRecipient(String recipient) { this.recipient = recipient; }

    public Map<String, Object> getArgs() { return args; }
    public void setArgs(Map<String, Object> args) { this.args = args; }
}