package com.funeral.funeralService.dto.order.request;

import jakarta.validation.constraints.NotNull;
import lombok.Data;

import java.util.ArrayList;
import java.util.List;

@Data
public class ReplaceOrderProductsRequest {

    @NotNull
    private List<Long> productIds = new ArrayList<>();
}
