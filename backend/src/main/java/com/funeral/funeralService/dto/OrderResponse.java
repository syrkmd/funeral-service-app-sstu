package com.funeral.funeralService.dto;

import lombok.Data;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.util.ArrayList;
import java.util.List;

@Data
public class OrderResponse {

    private String id;

    private LocalDate date;

    private String createdAt;

    private String updatedAt;

    private String statusUpdatedAt;

    private String paymentConfirmedAt;

    private String status;

    private BigDecimal total;

    private String phone;

    private Boolean isPaid;

    private ClientDto client;

    private DeceasedDto deceased;

    private List<OrderItemDto> services = new ArrayList<>();

    private List<OrderItemDto> products = new ArrayList<>();

    private List<OrderDocumentDto> documents = new ArrayList<>();
}
