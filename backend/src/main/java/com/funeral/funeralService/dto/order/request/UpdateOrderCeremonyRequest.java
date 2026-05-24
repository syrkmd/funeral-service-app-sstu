package com.funeral.funeralService.dto.order.request;

import jakarta.validation.constraints.FutureOrPresent;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Size;
import lombok.Data;

import java.time.LocalDate;

@Data
public class UpdateOrderCeremonyRequest {

    @NotNull
    @FutureOrPresent
    private LocalDate serviceDate;

    @NotBlank
    private String serviceTime;

    @NotBlank
    private String serviceAddress;

    @NotBlank
    private String cemetery;

    @Size(max = 1000)
    private String cemeteryNotes;
}
