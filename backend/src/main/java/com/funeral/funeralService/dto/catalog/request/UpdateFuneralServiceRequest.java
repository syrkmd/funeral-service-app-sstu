package com.funeral.funeralService.dto.catalog.request;

import jakarta.validation.constraints.PositiveOrZero;
import lombok.Data;

import java.math.BigDecimal;

@Data
public class UpdateFuneralServiceRequest {

    private String title;

    private String description;

    @PositiveOrZero
    private BigDecimal price;

    private Long categoryId;

    private Boolean active;

    private Integer sortOrder;
}
