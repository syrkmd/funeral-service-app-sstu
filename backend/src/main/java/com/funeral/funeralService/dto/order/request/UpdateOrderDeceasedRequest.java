package com.funeral.funeralService.dto.order.request;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import lombok.Data;

import java.time.LocalDate;

@Data
public class UpdateOrderDeceasedRequest {

    @NotBlank
    private String name;

    private LocalDate dateOfBirth;

    @NotNull
    private LocalDate dateOfDeath;
}
