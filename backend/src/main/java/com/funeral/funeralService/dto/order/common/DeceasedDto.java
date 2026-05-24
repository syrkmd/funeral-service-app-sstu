package com.funeral.funeralService.dto.order.common;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import lombok.Data;

import java.time.LocalDate;

@Data
public class DeceasedDto {

    @NotBlank
    private String name;

    private LocalDate dateOfBirth;

    @NotNull
    private LocalDate dateOfDeath;
}
