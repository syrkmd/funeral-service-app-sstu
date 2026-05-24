package com.funeral.funeralService.dto.catalog.response;

import io.swagger.v3.oas.annotations.media.Schema;
import lombok.AllArgsConstructor;
import lombok.Data;

@Data
@AllArgsConstructor
@Schema(description = "Категория товара")
public class ProductCategoryDto {

    @Schema(description = "Id категории", example = "2")
    private Long id;

    @Schema(description = "Название категории", example = "Ритуальные товары")
    private String name;
}
