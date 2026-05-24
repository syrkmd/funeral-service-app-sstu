package com.funeral.funeralService.dto.order.request;

import jakarta.validation.constraints.NotNull;
import lombok.Data;

import java.time.LocalDate;

@Data
public class UpdateOrderDateRequest {

    @NotNull
    private LocalDate date;
}
