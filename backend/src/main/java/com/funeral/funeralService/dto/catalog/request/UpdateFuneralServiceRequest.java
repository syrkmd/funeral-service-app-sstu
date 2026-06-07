package com.funeral.funeralService.dto.catalog.request;

import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.Positive;
import jakarta.validation.constraints.PositiveOrZero;
import lombok.Data;

import java.math.BigDecimal;

@Data
@Schema(description = "Админский PATCH-запрос для ритуальной услуги. Меняются только переданные поля.")
public class UpdateFuneralServiceRequest {

    @Schema(description = "Название услуги", example = "Традиционные похороны")
    private String title;

    @Schema(description = "Описание услуги", example = "Организация церемонии и сопровождение")
    private String description;

    @PositiveOrZero
    @Schema(description = "Цена услуги в рублях", example = "450000")
    private BigDecimal price;

    @Positive
    @Schema(description = "Id существующей категории услуги", example = "1")
    private Long categoryId;

    @Schema(description = "false отправляет в архив, true восстанавливает из архива", example = "true")
    private Boolean active;

    @PositiveOrZero
    @Schema(description = "Порядок отображения. Чем меньше число, тем выше элемент в списке.", example = "10")
    private Integer sortOrder;
}
