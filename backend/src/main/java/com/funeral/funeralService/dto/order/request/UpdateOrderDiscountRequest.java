package com.funeral.funeralService.dto.order.request;

import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.PositiveOrZero;
import lombok.Data;

import java.math.BigDecimal;

@Data
@Schema(description = "Запрос применения ручной скидки к заказу")
public class UpdateOrderDiscountRequest {

    @NotNull
    @PositiveOrZero
    @Schema(description = "Размер скидки в рублях", example = "10000")
    private BigDecimal discountAmount;

    @Schema(description = "Необязательная бизнес-причина скидки", example = "Скидка по согласованию с клиентом")
    private String reason;
}
