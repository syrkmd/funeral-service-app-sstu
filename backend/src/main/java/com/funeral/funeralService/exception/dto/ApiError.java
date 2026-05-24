package com.funeral.funeralService.exception.dto;

import io.swagger.v3.oas.annotations.media.Schema;
import lombok.Getter;

import java.time.LocalDateTime;

@Getter
@Schema(description = "Стандартное тело ответа с ошибкой")
public class ApiError {

    @Schema(description = "Стабильный технический код ошибки", example = "ORDER_NOT_FOUND")
    private final String code;

    @Schema(description = "Сообщение об ошибке для чтения человеком", example = "Order not found: ORD-F12EEB09")
    private final String message;

    @Schema(description = "Время возникновения ошибки, сформированное backend", example = "2026-05-24T20:14:23.838")
    private final LocalDateTime timestamp;


    public ApiError(String code, String message) {
        this.code = code;
        this.message = message;
        this.timestamp = LocalDateTime.now();
    }
}
