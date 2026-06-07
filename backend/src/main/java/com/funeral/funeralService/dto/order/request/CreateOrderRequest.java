package com.funeral.funeralService.dto.order.request;

import com.funeral.funeralService.dto.order.common.ClientDto;
import com.funeral.funeralService.dto.order.common.DeceasedDto;
import com.fasterxml.jackson.annotation.JsonAlias;
import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.Valid;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Pattern;
import jakarta.validation.constraints.Positive;
import jakarta.validation.constraints.Size;
import lombok.Data;

import java.time.LocalDate;
import java.util.ArrayList;
import java.util.List;

@Data
@Schema(description = "Запрос создания нового заказа из клиентской формы оформления")
public class CreateOrderRequest {

    @NotNull
    @Schema(description = "Основная дата заказа, выбранная клиентом", example = "2026-05-24")
    private LocalDate date;

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

    @Pattern(regexp = "^([01]\\d|2[0-3]):[0-5]\\d$", message = "Время должно быть в формате HH:mm")
    @Schema(description = "Время церемонии в формате HH:mm", example = "12:30")
    private String serviceTime;

    @Schema(description = "Адрес церемонии", example = "ул. Центральная, 10")
    private String serviceAddress;

    @Schema(description = "Название кладбища", example = "Основное кладбище")
    private String cemetery;

    @Schema(description = "Дополнительные примечания по кладбищу или церемонии", example = "Вход со стороны центральных ворот")
    private String cemeteryNotes;

    @Schema(description = "Id выбранного места захоронения из интеграции с кладбищем", example = "2")
    @Positive
    private Long cemeteryPlotId;

    @JsonAlias("cemeteryPlotLabel")
    @Schema(description = "Код выбранного места захоронения", example = "A-01-07")
    private String cemeteryPlotCode;

    @Valid
    @Size(min = 1)
    @Schema(description = "Выбранные услуги. Backend сверяет названия с активным каталогом и использует серверные цены.")
    private List<OrderCatalogSelectionRequest> services = new ArrayList<>();

    @Valid
    @Schema(description = "Выбранные товары. Backend сверяет названия с активным каталогом и использует серверные цены.")
    private List<OrderCatalogSelectionRequest> products = new ArrayList<>();

    @Valid
    @Schema(description = "Необязательные метаданные документов, прикрепляемые при создании заказа")
    private List<OrderDocumentRequest> documents = new ArrayList<>();
}
