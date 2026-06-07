package com.funeral.funeralService.dto.order.request;

import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Positive;
import lombok.Data;

import java.time.LocalDate;

@Data
@Schema(description = "Метаданные документа, прикреплённого к заказу")
public class OrderDocumentRequest {

    @Positive
    @Schema(description = "Id существующего документа, обычно не передаётся при создании", example = "1")
    private Long id;

    @NotBlank
    @Schema(description = "Отображаемое название документа", example = "Свидетельство о смерти")
    private String name;

    @NotBlank
    @Schema(description = "Тип документа", example = "certificate")
    private String type;

    @NotNull
    @Schema(description = "Дата документа", example = "2026-05-24")
    private LocalDate date;

    @Schema(description = "Размер файла в читаемом виде", example = "2 MB")
    private String size;
}
