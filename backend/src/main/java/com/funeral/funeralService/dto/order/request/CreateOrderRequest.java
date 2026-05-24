package com.funeral.funeralService.dto.order.request;

import com.funeral.funeralService.dto.order.common.ClientDto;
import com.funeral.funeralService.dto.order.common.DeceasedDto;
import com.funeral.funeralService.dto.order.common.OrderItemDto;
import io.swagger.v3.oas.annotations.media.Schema;
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
@Schema(description = "Запрос создания нового заказа из клиентской формы оформления")
public class CreateOrderRequest {

    @NotNull
    @Schema(description = "Основная дата заказа, выбранная клиентом", example = "2026-05-24")
    private LocalDate date;

    @Schema(description = "Необязательное время создания, переданное frontend", example = "2026-05-24T20:15:00")
    private String createdAt;

    @Schema(description = "Начальный статус заказа. Если не передан, используется processing.", example = "processing", allowableValues = {"processing", "confirmed", "completed", "cancelled"})
    private String status;

    @NotNull
    @PositiveOrZero
    @Schema(description = "Рассчитанная итоговая сумма заказа в рублях", example = "450000")
    private BigDecimal total;

    @NotBlank
    @Schema(description = "Телефон клиента, продублированный для быстрого поиска заказа", example = "+79991234567")
    private String phone;

    @NotNull
    @Schema(description = "Начальное состояние оплаты", example = "false")
    private Boolean isPaid;

    @Valid
    @NotNull
    @Schema(description = "Блок контактных данных клиента")
    private ClientDto client;

    @Valid
    @NotNull
    @Schema(description = "Блок данных умершего")
    private DeceasedDto deceased;

    @Schema(description = "Дата церемонии", example = "2026-05-27")
    private LocalDate serviceDate;

    @Schema(description = "Время церемонии в формате HH:mm", example = "12:30")
    private String serviceTime;

    @Schema(description = "Адрес церемонии", example = "ул. Центральная, 10")
    private String serviceAddress;

    @Schema(description = "Название кладбища", example = "Основное кладбище")
    private String cemetery;

    @Schema(description = "Дополнительные примечания по кладбищу или церемонии", example = "Вход со стороны центральных ворот")
    private String cemeteryNotes;

    @Schema(description = "Id выбранного места захоронения из интеграции с кладбищем", example = "2")
    private Long cemeteryPlotId;

    @Schema(description = "Понятное название выбранного места захоронения", example = "A-13")
    private String cemeteryPlotLabel;

    @Valid
    @Size(min = 1)
    @Schema(description = "Выбранные услуги. Требуется минимум одна услуга.")
    private List<OrderItemDto> services = new ArrayList<>();

    @Valid
    @Schema(description = "Выбранные товары. Может быть пустым списком.")
    private List<OrderItemDto> products = new ArrayList<>();

    @Valid
    @Schema(description = "Необязательные метаданные документов, прикрепляемые при создании заказа")
    private List<OrderDocumentRequest> documents = new ArrayList<>();
}
