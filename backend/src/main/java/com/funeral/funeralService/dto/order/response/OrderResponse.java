package com.funeral.funeralService.dto.order.response;

import com.funeral.funeralService.dto.order.common.ClientDto;
import com.funeral.funeralService.dto.order.common.DeceasedDto;
import com.funeral.funeralService.dto.order.common.OrderItemDto;
import io.swagger.v3.oas.annotations.media.Schema;
import lombok.Data;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.util.ArrayList;
import java.util.List;

@Data
@Schema(description = "Ответ с заказом для личного кабинета клиента и админ-панели")
public class OrderResponse {

    @Schema(description = "Бизнес-id заказа", example = "ORD-F12EEB09")
    private String id;

    @Schema(description = "Основная дата заказа", example = "2026-05-24")
    private LocalDate date;

    private String createdAt;

    private String updatedAt;

    private String statusUpdatedAt;

    private String paymentConfirmedAt;

    @Schema(description = "Текущий статус жизненного цикла заказа", example = "processing", allowableValues = {"processing", "confirmed", "completed", "cancelled"})
    private String status;

    @Schema(description = "Текущая итоговая сумма заказа в рублях", example = "450000")
    private BigDecimal total;

    @Schema(description = "Телефон клиента", example = "+79991234567")
    private String phone;

    @Schema(description = "Состояние оплаты", example = "false")
    private Boolean isPaid;

    private ClientDto client;

    private DeceasedDto deceased;

    private LocalDate serviceDate;

    private String serviceTime;

    private String serviceAddress;

    private String cemetery;

    private String cemeteryNotes;

    @Schema(description = "Id выбранного места захоронения", example = "2")
    private Long cemeteryPlotId;

    @Schema(description = "Код выбранного места захоронения", example = "A-01-07")
    private String cemeteryPlotCode;

    private List<OrderItemDto> services = new ArrayList<>();

    private List<OrderItemDto> products = new ArrayList<>();

    private List<OrderDocumentDto> documents = new ArrayList<>();
}
