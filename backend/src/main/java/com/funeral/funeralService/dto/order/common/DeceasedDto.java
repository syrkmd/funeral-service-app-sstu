package com.funeral.funeralService.dto.order.common;

import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import lombok.Data;

import java.time.LocalDate;

@Data
@Schema(description = "Данные умершего человека")
public class DeceasedDto {

    @NotBlank
    @Schema(description = "Полное имя умершего", example = "Пётр Иванов")
    private String name;

    @Schema(description = "Дата рождения", example = "1945-03-10")
    private LocalDate dateOfBirth;

    @NotNull
    @Schema(description = "Дата смерти", example = "2026-05-20")
    private LocalDate dateOfDeath;
}
