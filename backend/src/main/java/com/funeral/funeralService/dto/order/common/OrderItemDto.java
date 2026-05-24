package com.funeral.funeralService.dto.order.common;

import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.PositiveOrZero;
import lombok.Data;

import java.math.BigDecimal;

@Data
@Schema(description = "Выбранная позиция заказа, скопированная из каталога на момент оформления")
public class OrderItemDto {

    @NotBlank
    @Schema(description = "Название услуги или товара", example = "Традиционные похороны")
    private String name;

    @NotNull
    @PositiveOrZero
    @Schema(description = "Цена позиции в рублях", example = "450000")
    private BigDecimal price;
}
