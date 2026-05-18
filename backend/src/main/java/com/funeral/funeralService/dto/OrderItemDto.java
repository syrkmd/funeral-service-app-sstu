package com.funeral.funeralService.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.PositiveOrZero;
import lombok.Data;

import java.math.BigDecimal;

@Data
public class OrderItemDto {

    @NotBlank
    private String name;

    @NotNull
    @PositiveOrZero
    private BigDecimal price;
}
