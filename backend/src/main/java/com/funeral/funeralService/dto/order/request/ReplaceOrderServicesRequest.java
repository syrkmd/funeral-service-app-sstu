package com.funeral.funeralService.dto.order.request;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Size;
import lombok.Data;

import java.util.ArrayList;
import java.util.List;

@Data
public class ReplaceOrderServicesRequest {

    @NotNull
    @Size(min = 1)
    private List<Long> serviceIds = new ArrayList<>();

}
