package com.funeral.funeralService.controller;

import com.funeral.funeralService.dto.*;
import com.funeral.funeralService.service.OrderService;
import jakarta.validation.Valid;
import lombok.AllArgsConstructor;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/orders")
@RequiredArgsConstructor
public class OrderController {

    private final OrderService service;

    @GetMapping
    public ResponseEntity<List<OrderResponse>> getOrders(@RequestParam(required = false) String phone) {
        return ResponseEntity.ok(service.getOrders(phone));
    }

    @GetMapping("/{id}")
    public ResponseEntity<OrderResponse> getOrderById(@PathVariable String id) {
        return ResponseEntity.ok(service.getOrderById(id));
    }

    @PostMapping
    public ResponseEntity<OrderResponse> createOrder(@Valid @RequestBody CreateOrderRequest request) {
        return ResponseEntity
                .status(HttpStatus.CREATED)
                .body(service.createOrder(request));
    }

    @PatchMapping("/{id}/status")
    public ResponseEntity<OrderResponse> updateStatus(@PathVariable String id, @Valid @RequestBody UpdateOrderStatusRequest request) {
        return ResponseEntity.ok(service.updateStatus(id, request));
    }

    @PatchMapping("/{id}/payment")
    public ResponseEntity<OrderResponse> updatePayment(@PathVariable String id, @Valid @RequestBody UpdatePaymentRequest request) {
        return ResponseEntity.ok(service.updatePayment(id, request));
    }

    @PostMapping("/{id}/documents")
    public ResponseEntity<OrderResponse> addDocument(@PathVariable String id, @Valid @RequestBody OrderDocumentRequest request) {
        return ResponseEntity.ok(service.addDocument(id, request));
    }

    @DeleteMapping("/{id}/documents/{documentId}")
    public ResponseEntity<OrderResponse> removeDocument(@PathVariable String id, @PathVariable Long documentId) {
        return ResponseEntity.ok(service.removeDocument(id, documentId));
    }

}
