package com.funeral.funeralService.dto.order.request;

import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.NotNull;
import lombok.Data;

import java.time.LocalDate;

@Data
@Schema(description = "Запрос изменения основной даты заказа")
public class UpdateOrderDateRequest {

    @NotNull
    @Schema(description = "Новая основная дата заказа", example = "2026-05-25")
    private LocalDate date;
}
