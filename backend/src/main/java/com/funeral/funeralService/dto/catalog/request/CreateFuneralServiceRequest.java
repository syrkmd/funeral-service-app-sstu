package com.funeral.funeralService.dto.catalog.request;

import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.PositiveOrZero;
import lombok.Data;

import java.math.BigDecimal;

@Data
@Schema(description = "Админский запрос для создания ритуальной услуги")
public class CreateFuneralServiceRequest {

    @NotBlank
    @Schema(description = "Название услуги", example = "Традиционные похороны")
    private String title;

    @Schema(description = "Описание услуги", example = "Организация церемонии и сопровождение")
    private String description;

    @NotNull
    @PositiveOrZero
    @Schema(description = "Цена услуги в рублях", example = "450000")
    private BigDecimal price;

    @NotNull
    @Schema(description = "Id существующей категории услуги", example = "1")
    private Long categoryId;

    @Schema(description = "Показывается ли услуга при оформлении заказа", example = "true")
    private Boolean active = true;

    @Schema(description = "Порядок отображения. Чем меньше число, тем выше элемент в списке.", example = "10")
    private Integer sortOrder;
}
