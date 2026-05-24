package com.funeral.funeralService.dto.catalog.response;

import io.swagger.v3.oas.annotations.media.Schema;
import lombok.AllArgsConstructor;
import lombok.Data;

import java.math.BigDecimal;

@Data
@AllArgsConstructor
@Schema(description = "Ответ с товаром каталога")
public class CatalogProductDto {

    @Schema(description = "Id товара", example = "8")
    private Long id;

    @Schema(description = "Название товара", example = "Книга памяти")
    private String title;

    private String description;

    @Schema(description = "Цена товара в рублях", example = "7500")
    private BigDecimal price;

    private String imageUrl;

    private ProductCategoryDto category;

    @Schema(description = "Активен товар или находится в архиве", example = "true")
    private Boolean active;

    @Schema(description = "Порядок отображения", example = "10")
    private Integer sortOrder;
}
