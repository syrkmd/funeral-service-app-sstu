package com.funeral.funeralService.dto.cemetery.response;

import io.swagger.v3.oas.annotations.media.Schema;
import lombok.AllArgsConstructor;
import lombok.Data;

@Data
@AllArgsConstructor
@Schema(description = "Результат резервирования места захоронения")
public class ReservePlotResponse {

    @Schema(description = "Успешно ли прошло резервирование", example = "true")
    private boolean success;

    @Schema(description = "Внешний id резервирования из сервиса кладбища", example = "RES-2-ORD-F12EEB09")
    private String reservationId;
}
