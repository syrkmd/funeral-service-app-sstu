package com.funeral.funeralService.dto.catalog.response;

import lombok.AllArgsConstructor;
import lombok.Data;

import java.math.BigDecimal;

@Data
@AllArgsConstructor
public class CatalogProductDto {

    private Long id;

    private String title;

    private String description;

    private BigDecimal price;

    private String imageUrl;

    private ProductCategoryDto category;

    private Boolean active;

    private Integer sortOrder;
}
