package com.funeral.funeralService.dto.order.request;

import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.NotBlank;
import lombok.Data;

@Data
@Schema(description = "Запрос изменения статуса жизненного цикла заказа")
public class UpdateOrderStatusRequest {

    @NotBlank
    @Schema(description = "Новый статус заказа", example = "confirmed", allowableValues = {"processing", "confirmed", "completed", "cancelled"})
    private String status;
}
