package com.funeral.funeralService.dto.order.request;

import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.PositiveOrZero;
import lombok.Data;

import java.math.BigDecimal;

@Data
public class UpdateOrderDiscountRequest {

    @NotNull
    @PositiveOrZero
    private BigDecimal discountAmount;

    private String reason;
}
