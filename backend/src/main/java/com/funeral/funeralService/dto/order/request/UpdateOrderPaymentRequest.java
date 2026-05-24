package com.funeral.funeralService.dto.order.request;

import jakarta.validation.constraints.NotNull;
import lombok.Data;

@Data
public class UpdateOrderPaymentRequest {

    @NotNull
    private Boolean isPaid;
}
