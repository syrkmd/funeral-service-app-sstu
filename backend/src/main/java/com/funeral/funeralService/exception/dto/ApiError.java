package com.funeral.funeralService.exception.dto;

import lombok.Getter;

import java.time.LocalDateTime;

@Getter
public class ApiError {

    private final String code;
    private final String message;
    private final LocalDateTime timestamp;


    public ApiError(String code, String message) {
        this.code = code;
        this.message = message;
        this.timestamp = LocalDateTime.now();
    }
}
