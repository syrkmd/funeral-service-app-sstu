package com.funeral.funeralService.dto.cemetery.response;

import io.swagger.v3.oas.annotations.media.Schema;
import lombok.AllArgsConstructor;
import lombok.Data;

@Data
@AllArgsConstructor
@Schema(description = "Место захоронения, возвращаемое fake-интеграцией кладбища")
public class CemeteryPlotResponse {

    @Schema(description = "Id места", example = "2")
    private Long id;

    @Schema(description = "Понятное название места", example = "A-13")
    private String label;

    @Schema(description = "Можно ли выбрать это место", example = "true")
    private Boolean available;
}
