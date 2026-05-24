package com.funeral.funeralService.dto.order.request;

import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.FutureOrPresent;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Size;
import lombok.Data;

import java.time.LocalDate;

@Data
@Schema(description = "Запрос изменения данных церемонии и кладбища")
public class UpdateOrderCeremonyRequest {

    @NotNull
    @FutureOrPresent
    @Schema(description = "Дата церемонии", example = "2026-05-27")
    private LocalDate serviceDate;

    @NotBlank
    @Schema(description = "Время церемонии в формате HH:mm", example = "12:30")
    private String serviceTime;

    @NotBlank
    @Schema(description = "Адрес церемонии", example = "ул. Центральная, 10")
    private String serviceAddress;

    @NotBlank
    @Schema(description = "Название кладбища", example = "Основное кладбище")
    private String cemetery;

    @Size(max = 1000)
    @Schema(description = "Необязательные примечания по кладбищу", example = "Вход со стороны центральных ворот")
    private String cemeteryNotes;
}
