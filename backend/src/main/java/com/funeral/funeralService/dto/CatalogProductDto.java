package com.funeral.funeralService.dto;

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
}
