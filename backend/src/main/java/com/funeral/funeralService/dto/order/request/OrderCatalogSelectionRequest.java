package com.funeral.funeralService.dto.order.request;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.NotBlank;
import lombok.Data;

@Data
@JsonIgnoreProperties(ignoreUnknown = true)
@Schema(description = "Выбранная позиция каталога. Цена всегда определяется backend по активному каталогу.")
public class OrderCatalogSelectionRequest {

    @NotBlank
    @Schema(description = "Точное название активной позиции каталога", example = "Традиционные похороны")
    private String name;
}
