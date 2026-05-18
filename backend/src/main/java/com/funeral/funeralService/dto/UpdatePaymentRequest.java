package com.funeral.funeralService.dto;

import jakarta.validation.constraints.NotNull;
import lombok.Data;

@Data
public class UpdatePaymentRequest {

    @NotNull
    private Boolean isPaid;
}
