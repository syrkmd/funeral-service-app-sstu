package com.funeral.funeralService.dto.order.request;

import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.NotNull;
import lombok.Data;

import java.util.ArrayList;
import java.util.List;

@Data
@Schema(description = "Запрос полной замены товаров заказа по id из каталога")
public class ReplaceOrderProductsRequest {

    @NotNull
    @Schema(description = "Id товаров из каталога. Пустой список означает, что товары не выбраны.", example = "[3, 7]")
    private List<Long> productIds = new ArrayList<>();
}
