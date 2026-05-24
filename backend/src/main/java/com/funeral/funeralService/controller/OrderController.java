package com.funeral.funeralService.controller;

import com.funeral.funeralService.dto.order.request.*;
import com.funeral.funeralService.dto.order.response.OrderResponse;
import com.funeral.funeralService.service.OrderService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/orders")
@RequiredArgsConstructor
@Tag(name = "Заказы", description = "Оформление заказов клиентом и админские endpoints для изменения заказа")
public class OrderController {

    private final OrderService service;

    @GetMapping
    @Operation(
            summary = "Найти заказы",
            description = "Возвращает список заказов, при необходимости фильтрует по телефону клиента. Если заказов нет, возвращает 200 OK и пустой массив."
    )
    @ApiResponse(responseCode = "200", description = "Список заказов возвращён")
    public List<OrderResponse> getOrders(
            @Parameter(description = "Необязательный фильтр по телефону клиента", example = "+79991234567")
            @RequestParam(required = false) String phone
    ) {
        return service.getOrders(phone);
    }

    @GetMapping("/{id}")
    @Operation(summary = "Получить заказ по id", description = "Возвращает один заказ по его бизнес-идентификатору.")
    @ApiResponse(responseCode = "200", description = "Заказ найден")
    @ApiResponse(responseCode = "404", description = "Заказ не найден")
    public OrderResponse getOrderById(
            @Parameter(description = "Id заказа", example = "ORD-F12EEB09")
            @PathVariable String id
    ) {
        return service.getOrderById(id);
    }

    @PostMapping
    @ResponseStatus(HttpStatus.CREATED)
    @Operation(summary = "Создать заказ", description = "Клиентский сценарий оформления заказа. Если выбрано место захоронения, backend резервирует его через интеграционный слой кладбища.")
    @ApiResponse(responseCode = "201", description = "Заказ создан")
    @ApiResponse(responseCode = "400", description = "Ошибка валидации")
    public OrderResponse createOrder(@Valid @RequestBody CreateOrderRequest request) {
        return service.createOrder(request);
    }

    @PatchMapping("/{id}/status")
    @Operation(summary = "Изменить статус заказа", description = "Админский endpoint для изменения жизненного цикла заказа: processing, confirmed, completed или cancelled.")
    @ApiResponse(responseCode = "200", description = "Статус изменён")
    @ApiResponse(responseCode = "404", description = "Заказ не найден")
    public OrderResponse updateStatus(@Parameter(example = "ORD-F12EEB09") @PathVariable String id, @Valid @RequestBody UpdateOrderStatusRequest request) {
        return service.updateStatus(id, request);
    }

    @PatchMapping("/{id}/payment")
    @Operation(summary = "Изменить статус оплаты", description = "Админский endpoint для отметки заказа как оплаченного или неоплаченного.")
    @ApiResponse(responseCode = "200", description = "Статус оплаты изменён")
    @ApiResponse(responseCode = "404", description = "Заказ не найден")
    public OrderResponse updatePayment(@Parameter(example = "ORD-F12EEB09") @PathVariable String id, @Valid @RequestBody UpdateOrderPaymentRequest request) {
        return service.updatePayment(id, request);
    }

    @PatchMapping("/{id}/client")
    @Operation(summary = "Изменить данные клиента", description = "Админский endpoint для исправления имени, телефона или email клиента после создания заказа.")
    @ApiResponse(responseCode = "200", description = "Данные клиента изменены")
    @ApiResponse(responseCode = "404", description = "Заказ не найден")
    public OrderResponse updateClient(@Parameter(example = "ORD-F12EEB09") @PathVariable String id, @Valid @RequestBody UpdateOrderClientRequest request) {
        return service.updateClient(id, request);
    }

    @PatchMapping("/{id}/deceased")
    @Operation(summary = "Изменить данные умершего", description = "Админский endpoint для исправления данных умершего человека.")
    @ApiResponse(responseCode = "200", description = "Данные умершего изменены")
    @ApiResponse(responseCode = "404", description = "Заказ не найден")
    public OrderResponse updateDeceased(@Parameter(example = "ORD-F12EEB09") @PathVariable String id, @Valid @RequestBody UpdateOrderDeceasedRequest request) {
        return service.updateDeceased(id, request);
    }

    @PutMapping("/{id}/services")
    @Operation(summary = "Заменить услуги заказа", description = "Админский endpoint полностью заменяет выбранные услуги заказа по id из каталога услуг.")
    @ApiResponse(responseCode = "200", description = "Услуги заменены, сумма заказа пересчитана")
    @ApiResponse(responseCode = "404", description = "Заказ или услуга не найдены")
    public OrderResponse replaceServices(@Parameter(example = "ORD-F12EEB09") @PathVariable String id, @Valid @RequestBody ReplaceOrderServicesRequest request) {
        return service.replaceServices(id, request);
    }

    @PutMapping("/{id}/products")
    @Operation(summary = "Заменить товары заказа", description = "Админский endpoint полностью заменяет выбранные товары заказа по id из каталога товаров.")
    @ApiResponse(responseCode = "200", description = "Товары заменены, сумма заказа пересчитана")
    @ApiResponse(responseCode = "404", description = "Заказ или товар не найдены")
    public OrderResponse replaceProducts(@Parameter(example = "ORD-F12EEB09") @PathVariable String id, @Valid @RequestBody ReplaceOrderProductsRequest request) {
        return service.replaceProducts(id, request);
    }

    @PatchMapping("/{id}/date")
    @Operation(summary = "Изменить дату заказа", description = "Админский endpoint для изменения основной даты заказа.")
    @ApiResponse(responseCode = "200", description = "Дата заказа изменена")
    @ApiResponse(responseCode = "404", description = "Заказ не найден")
    public OrderResponse updateDate(@Parameter(example = "ORD-F12EEB09") @PathVariable String id, @Valid @RequestBody UpdateOrderDateRequest request) {
        return service.updateDate(id, request);
    }

    @PatchMapping("/{id}/discount")
    @Operation(summary = "Применить скидку", description = "Админский endpoint для применения ручной скидки к итоговой сумме заказа.")
    @ApiResponse(responseCode = "200", description = "Скидка применена")
    @ApiResponse(responseCode = "404", description = "Заказ не найден")
    public OrderResponse updateDiscount(@Parameter(example = "ORD-F12EEB09") @PathVariable String id, @Valid @RequestBody UpdateOrderDiscountRequest request) {
        return service.updateDiscount(id, request);
    }

    @PatchMapping("/{id}/ceremony")
    @Operation(summary = "Изменить детали церемонии", description = "Админский endpoint для изменения даты, времени, адреса церемонии и примечаний по кладбищу.")
    @ApiResponse(responseCode = "200", description = "Детали церемонии изменены")
    @ApiResponse(responseCode = "404", description = "Заказ не найден")
    public OrderResponse updateCeremony(@Parameter(example = "ORD-F12EEB09") @PathVariable String id, @Valid @RequestBody UpdateOrderCeremonyRequest request) {
        return service.updateCeremony(id, request);
    }

    @PostMapping("/{id}/documents")
    @Operation(summary = "Добавить документ к заказу", description = "Админский endpoint для добавления метаданных документа к заказу.")
    @ApiResponse(responseCode = "200", description = "Документ добавлен")
    @ApiResponse(responseCode = "404", description = "Заказ не найден")
    public OrderResponse addDocument(@Parameter(example = "ORD-F12EEB09") @PathVariable String id, @Valid @RequestBody OrderDocumentRequest request) {
        return service.addDocument(id, request);
    }

    @DeleteMapping("/{id}/documents/{documentId}")
    @Operation(summary = "Удалить документ из заказа", description = "Админский endpoint для удаления одного документа из заказа.")
    @ApiResponse(responseCode = "200", description = "Документ удалён")
    @ApiResponse(responseCode = "404", description = "Заказ или документ не найдены")
    public OrderResponse removeDocument(
            @Parameter(example = "ORD-F12EEB09") @PathVariable String id,
            @Parameter(description = "Id документа", example = "1") @PathVariable Long documentId
    ) {
        return service.removeDocument(id, documentId);
    }

    @DeleteMapping("/{id}")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    @Operation(summary = "Удалить заказ", description = "Админский endpoint для удаления заказа из системы.")
    @ApiResponse(responseCode = "204", description = "Заказ удалён")
    @ApiResponse(responseCode = "404", description = "Заказ не найден")
    public void deleteOrder(@Parameter(example = "ORD-F12EEB09") @PathVariable String id) {
        service.deleteOrder(id);
    }

}
