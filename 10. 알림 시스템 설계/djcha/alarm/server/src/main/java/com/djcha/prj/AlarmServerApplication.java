package com.djcha.prj;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.retry.annotation.EnableRetry;

@EnableRetry
@SpringBootApplication
public class AlarmServerApplication {
    public static void main(String[] args) {
        SpringApplication.run(AlarmServerApplication.class);
    }
}
