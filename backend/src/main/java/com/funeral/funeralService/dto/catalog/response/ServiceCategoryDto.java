package com.funeral.funeralService.dto.catalog.response;

import io.swagger.v3.oas.annotations.media.Schema;
import lombok.AllArgsConstructor;
import lombok.Data;

@Data
@AllArgsConstructor
@Schema(description = "Категория ритуальной услуги")
public class ServiceCategoryDto {

    @Schema(description = "Id категории", example = "1")
    private Long id;

    @Schema(description = "Название категории", example = "Похоронные услуги")
    private String name;
}
