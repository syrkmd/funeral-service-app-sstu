package com.funeral.funeralService.dto.catalog.request;

import jakarta.validation.constraints.PositiveOrZero;
import lombok.Data;

import java.math.BigDecimal;

@Data
public class UpdateCatalogProductRequest {

    private String title;

    private String description;

    @PositiveOrZero
    private BigDecimal price;

    private String imageUrl;

    private Long categoryId;

    private Boolean active;

    private Integer sortOrder;
}
