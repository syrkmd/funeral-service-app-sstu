package com.funeral.funeralService.service;

import com.funeral.funeralService.dto.*;
import com.funeral.funeralService.entity.Order;
import com.funeral.funeralService.entity.OrderDocument;
import com.funeral.funeralService.exception.OrderDocumentNotFoundException;
import com.funeral.funeralService.exception.OrderNotFoundException;
import com.funeral.funeralService.mapper.OrderMapper;
import com.funeral.funeralService.repository.OrderRepository;
import lombok.AllArgsConstructor;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.UUID;

@Service
@AllArgsConstructor
public class OrderService {

    private OrderMapper mapper;
    private OrderRepository repository;

    public List<OrderResponse> getOrders(String phone) {
        List<Order> orders;

        if (phone != null && !phone.isBlank()) {
            orders = repository.findByClientPhone(phone);
        } else {
            orders = repository.findAll();
        }

        return mapper.toResponseList(orders);
    }

    public OrderResponse getOrderById(String id) {
        Order order = findOrder(id);
        return mapper.toResponse(order);
    }

    public OrderResponse createOrder(CreateOrderRequest request) {
        Order order = mapper.toEntity(request);
        order.setId(generateOrderId());

        Order savedOrder = repository.save(order);

        return mapper.toResponse(savedOrder);
    }

    public OrderResponse updateStatus(String id, UpdateOrderStatusRequest request) {
        Order order = findOrder(id);

        order.setStatus(mapper.toOrderStatus(request.getStatus()));

        Order savedOrder = repository.save(order);

        return mapper.toResponse(savedOrder);
    }

    public OrderResponse updatePayment(String id, UpdatePaymentRequest request) {
        Order order = findOrder(id);

        order.setPaid(request.getIsPaid());

        Order savedOrder = repository.save(order);

        return mapper.toResponse(savedOrder);
    }

    public OrderResponse addDocument(String orderId, OrderDocumentRequest request) {
        Order order = findOrder(orderId);
        OrderDocument document = new OrderDocument();

        document.setOrder(order);
        document.setName(request.getName());
        document.setType(request.getType());
        document.setDocumentDate(request.getDate());
        document.setSize(request.getSize());

        order.getDocuments().add(document);

        Order savedOrder = repository.save(order);

        return mapper.toResponse(savedOrder);
    }

    public OrderResponse removeDocument(String orderId, Long documentId) {
        Order order = findOrder(orderId);

        boolean removed = order.getDocuments().removeIf(
                document -> document.getId().equals(documentId)
        );

        if (!removed) {
            throw new OrderDocumentNotFoundException(documentId);
        }

        Order savedOrder = repository.save(order);

        return mapper.toResponse(savedOrder);
    }


    private String generateOrderId() {
        return "ORD-" + UUID.randomUUID()
                .toString()
                .replace("-","")
                .substring(0, 8)
                .toUpperCase();
    }

    private Order findOrder(String id) {
        return repository.findById(id)
                .orElseThrow(() -> new OrderNotFoundException(id));
    }
}
