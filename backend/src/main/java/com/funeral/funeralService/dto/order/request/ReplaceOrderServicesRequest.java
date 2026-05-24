package com.funeral.funeralService.dto.order.request;
import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Size;
import lombok.Data;

import java.util.ArrayList;
import java.util.List;

@Data
@Schema(description = "Запрос полной замены услуг заказа по id из каталога")
public class ReplaceOrderServicesRequest {

    @NotNull
    @Size(min = 1)
    @Schema(description = "Id услуг из каталога. Список полностью заменяет текущие услуги заказа.", example = "[1, 2]")
    private List<Long> serviceIds = new ArrayList<>();

}
