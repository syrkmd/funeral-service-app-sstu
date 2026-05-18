package com.funeral.funeralService.dto;

import jakarta.validation.Valid;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.PositiveOrZero;
import jakarta.validation.constraints.Size;
import lombok.Data;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.util.ArrayList;
import java.util.List;

@Data
public class CreateOrderRequest {

    @NotNull
    private LocalDate date;

    private String createdAt;

    private String status;

    @NotNull
    @PositiveOrZero
    private BigDecimal total;

    @NotBlank
    private String phone;

    @NotNull
    private Boolean isPaid;

    @Valid
    @NotNull
    private ClientDto client;

    @Valid
    @NotNull
    private DeceasedDto deceased;

    @Valid
    @Size(min = 1)
    private List<OrderItemDto> services = new ArrayList<>();

    @Valid
    private List<OrderItemDto> products = new ArrayList<>();

    @Valid
    private List<OrderDocumentRequest> documents = new ArrayList<>();
}
