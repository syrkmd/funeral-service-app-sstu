package com.funeral.funeralService.dto.order.common;

import com.fasterxml.jackson.annotation.JsonIgnore;
import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.AssertTrue;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Pattern;
import lombok.Data;

import java.time.LocalDate;

@Data
@Schema(description = "Данные умершего человека")
public class DeceasedDto {

    @NotBlank
    @Pattern(
            regexp = "^[\\p{L}]+(?:[ '\\-][\\p{L}]+)*$",
            message = "Имя может содержать только буквы, пробелы, дефисы и апострофы"
    )
    @Schema(description = "Полное имя умершего", example = "Пётр Иванов")
    private String name;

    @Schema(description = "Дата рождения", example = "1945-03-10")
    private LocalDate dateOfBirth;

    @NotNull
    @Schema(description = "Дата смерти", example = "2026-05-20")
    private LocalDate dateOfDeath;

    @AssertTrue(message = "Дата смерти должна быть позже даты рождения")
    @JsonIgnore
    @Schema(hidden = true)
    public boolean isDeathDateAfterBirth() {
        return dateOfBirth == null
                || dateOfDeath == null
                || dateOfDeath.isAfter(dateOfBirth);
    }
}
