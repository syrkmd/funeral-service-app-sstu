package com.funeral.funeralService.dto.order.request;

import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.NotNull;
import lombok.Data;

@Data
@Schema(description = "Запрос изменения состояния оплаты")
public class UpdateOrderPaymentRequest {

    @NotNull
    @Schema(description = "Оплачен ли заказ", example = "true")
    private Boolean isPaid;
}
