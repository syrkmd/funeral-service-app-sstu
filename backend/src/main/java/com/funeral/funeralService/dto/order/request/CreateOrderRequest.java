package com.funeral.funeralService.dto.order.request;

import com.funeral.funeralService.dto.order.common.ClientDto;
import com.funeral.funeralService.dto.order.common.DeceasedDto;
import com.funeral.funeralService.dto.order.common.OrderItemDto;
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

    private LocalDate serviceDate;

    private String serviceTime;

    private String serviceAddress;

    private String cemetery;

    private String cemeteryNotes;

    private Long cemeteryPlotId;

    private String cemeteryPlotLabel;

    @Valid
    @Size(min = 1)
    private List<OrderItemDto> services = new ArrayList<>();

    @Valid
    private List<OrderItemDto> products = new ArrayList<>();

    @Valid
    private List<OrderDocumentRequest> documents = new ArrayList<>();
}
