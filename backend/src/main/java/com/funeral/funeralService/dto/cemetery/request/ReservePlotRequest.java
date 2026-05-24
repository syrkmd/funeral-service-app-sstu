package com.funeral.funeralService.dto.cemetery.request;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import lombok.Data;

@Data
public class ReservePlotRequest {

    @NotNull
    private Long plotId;

    @NotBlank
    private String orderId;
}
