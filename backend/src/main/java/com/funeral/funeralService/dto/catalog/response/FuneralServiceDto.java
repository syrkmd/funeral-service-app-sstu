package com.funeral.funeralService.dto.catalog.response;

import io.swagger.v3.oas.annotations.media.Schema;
import lombok.AllArgsConstructor;
import lombok.Data;

import java.math.BigDecimal;

@Data
@AllArgsConstructor
@Schema(description = "Ответ с ритуальной услугой из каталога")
public class FuneralServiceDto {

    @Schema(description = "Id услуги", example = "1")
    private Long id;

    @Schema(description = "Название услуги", example = "Традиционные похороны")
    private String title;

    private String description;

    @Schema(description = "Цена услуги в рублях", example = "450000")
    private BigDecimal price;

    private ServiceCategoryDto category;

    @Schema(description = "Активна услуга или находится в архиве", example = "true")
    private Boolean active;

    @Schema(description = "Порядок отображения", example = "10")
    private Integer sortOrder;
}
