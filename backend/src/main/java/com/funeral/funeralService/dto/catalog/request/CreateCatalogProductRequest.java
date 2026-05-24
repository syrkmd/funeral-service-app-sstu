package com.funeral.funeralService.dto.catalog.request;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.PositiveOrZero;
import lombok.Data;

import java.math.BigDecimal;

@Data
public class CreateCatalogProductRequest {

    @NotBlank
    private String title;

    private String description;

    @NotNull
    @PositiveOrZero
    private BigDecimal price;

    private String imageUrl;

    @NotNull
    private Long categoryId;

    private Boolean active = true;

    private Integer sortOrder;
}