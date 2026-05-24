package com.funeral.funeralService.dto.catalog.request;

import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.PositiveOrZero;
import lombok.Data;

import java.math.BigDecimal;

@Data
@Schema(description = "Админский запрос для создания товара каталога")
public class CreateCatalogProductRequest {

    @NotBlank
    @Schema(description = "Название товара", example = "Книга памяти")
    private String title;

    @Schema(description = "Описание товара", example = "Памятная книга для записей гостей")
    private String description;

    @NotNull
    @PositiveOrZero
    @Schema(description = "Цена товара в рублях", example = "7500")
    private BigDecimal price;

    @Schema(description = "Необязательная ссылка на изображение товара", example = "https://example.com/products/memory-book.jpg")
    private String imageUrl;

    @NotNull
    @Schema(description = "Id существующей категории товара", example = "2")
    private Long categoryId;

    @Schema(description = "Показывается ли товар в публичном каталоге", example = "true")
    private Boolean active = true;

    @Schema(description = "Порядок отображения. Чем меньше число, тем выше элемент в списке.", example = "10")
    private Integer sortOrder;
}
