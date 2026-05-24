package com.funeral.funeralService.dto.cemetery.request;

import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import lombok.Data;

@Data
@Schema(description = "Запрос резервирования места захоронения через интеграционный сервис")
public class ReservePlotRequest {

    @NotNull
    @Schema(description = "Id места захоронения", example = "2")
    private Long plotId;

    @NotBlank
    @Schema(description = "Id заказа, для которого создаётся резервирование", example = "ORD-F12EEB09")
    private String orderId;
}
