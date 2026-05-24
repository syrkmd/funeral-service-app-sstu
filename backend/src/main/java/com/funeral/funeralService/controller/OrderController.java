package com.funeral.funeralService.controller;

import com.funeral.funeralService.dto.order.request.*;
import com.funeral.funeralService.dto.order.response.OrderResponse;
import com.funeral.funeralService.service.OrderService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/orders")
@RequiredArgsConstructor
public class OrderController {

    private final OrderService service;

    @GetMapping
    public List<OrderResponse> getOrders(@RequestParam(required = false) String phone) {
        return service.getOrders(phone);
    }

    @GetMapping("/{id}")
    public OrderResponse getOrderById(@PathVariable String id) {
        return service.getOrderById(id);
    }

    @PostMapping
    @ResponseStatus(HttpStatus.CREATED)
    public OrderResponse createOrder(@Valid @RequestBody CreateOrderRequest request) {
        return service.createOrder(request);
    }

    @PatchMapping("/{id}/status")
    public OrderResponse updateStatus(@PathVariable String id, @Valid @RequestBody UpdateOrderStatusRequest request) {
        return service.updateStatus(id, request);
    }

    @PatchMapping("/{id}/payment")
    public OrderResponse updatePayment(@PathVariable String id, @Valid @RequestBody UpdateOrderPaymentRequest request) {
        return service.updatePayment(id, request);
    }

    @PatchMapping("/{id}/client")
    public OrderResponse updateClient(@PathVariable String id, @Valid @RequestBody UpdateOrderClientRequest request) {
        return service.updateClient(id, request);
    }

    @PatchMapping("/{id}/deceased")
    public OrderResponse updateDeceased(@PathVariable String id, @Valid @RequestBody UpdateOrderDeceasedRequest request) {
        return service.updateDeceased(id, request);
    }

    @PutMapping("/{id}/services")
    public OrderResponse replaceServices(@PathVariable String id, @Valid @RequestBody ReplaceOrderServicesRequest request) {
        return service.replaceServices(id, request);
    }

    @PutMapping("/{id}/products")
    public OrderResponse replaceProducts(@PathVariable String id, @Valid @RequestBody ReplaceOrderProductsRequest request) {
        return service.replaceProducts(id, request);
    }

    @PatchMapping("/{id}/date")
    public OrderResponse updateDate(@PathVariable String id, @Valid @RequestBody UpdateOrderDateRequest request) {
        return service.updateDate(id, request);
    }

    @PatchMapping("/{id}/discount")
    public OrderResponse updateDiscount(@PathVariable String id, @Valid @RequestBody UpdateOrderDiscountRequest request) {
        return service.updateDiscount(id, request);
    }

    @PatchMapping("/{id}/ceremony")
    public OrderResponse updateCeremony(@PathVariable String id, @Valid @RequestBody UpdateOrderCeremonyRequest request) {
        return service.updateCeremony(id, request);
    }

    @PostMapping("/{id}/documents")
    public OrderResponse addDocument(@PathVariable String id, @Valid @RequestBody OrderDocumentRequest request) {
        return service.addDocument(id, request);
    }

    @DeleteMapping("/{id}/documents/{documentId}")
    public OrderResponse removeDocument(@PathVariable String id, @PathVariable Long documentId) {
        return service.removeDocument(id, documentId);
    }

    @DeleteMapping("/{id}")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void deleteOrder(@PathVariable String id) {
        service.deleteOrder(id);
    }

}
