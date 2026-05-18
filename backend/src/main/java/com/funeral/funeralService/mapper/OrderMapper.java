package com.funeral.funeralService.mapper;

import com.funeral.funeralService.dto.*;
import com.funeral.funeralService.entity.*;
import org.springframework.stereotype.Component;

import java.util.List;

@Component
public class OrderMapper {

    public Order toEntity(CreateOrderRequest request) {
        Order order = new Order();

        order.setOrderDate(request.getDate());
        order.setStatus(parseStatus(request.getStatus()));
        order.setTotalAmount(request.getTotal());
        order.setPaid(request.getIsPaid());

        order.setClientName(request.getClient().getName());
        order.setClientPhone(request.getPhone());
        order.setClientEmail(request.getClient().getEmail());

        order.setDeceasedName(request.getDeceased().getName());
        order.setDeceasedDateOfBirth(request.getDeceased().getDateOfBirth());
        order.setDeceasedDateOfDeath(request.getDeceased().getDateOfDeath());

        order.setServices(toServiceItems(request.getServices(), order));
        order.setProducts(toProductsItem(request.getProducts(), order));
        order.setDocuments(toDocumentItems(request.getDocuments(), order));

        return order;
    }

    public OrderResponse toResponse(Order order) {
        OrderResponse response = new OrderResponse();
        ClientDto client = new ClientDto();
        DeceasedDto deceased = new DeceasedDto();

        response.setId(order.getId());
        response.setDate(order.getOrderDate());
        response.setStatus(order.getStatus().name().toLowerCase());
        response.setTotal(order.getTotalAmount());
        response.setPhone(order.getClientPhone());
        response.setIsPaid(order.getPaid());

        client.setName(order.getClientName());
        client.setPhone(order.getClientPhone());
        client.setEmail(order.getClientEmail());

        deceased.setName(order.getDeceasedName());
        deceased.setDateOfBirth(order.getDeceasedDateOfBirth());
        deceased.setDateOfDeath(order.getDeceasedDateOfDeath());
        response.setDeceased(deceased);

        response.setServices(
                order.getServices().stream()
                        .map(this::toOrderItemDto)
                        .toList()
        );

        response.setProducts(
                order.getProducts().stream()
                        .map(this::toOrderItemDto)
                        .toList()
        );

        response.setDocuments(
                order.getDocuments().stream()
                        .map(this::toDocumentDto)
                        .toList()
        );

        return response;
    }

    public List<OrderResponse> toResponseList(List<Order> orders) {
        return orders.stream()
                .map(this::toResponse)
                .toList();
    }

    public OrderStatus toOrderStatus(String status) {
        if (status == null || status.isBlank()) {
            return OrderStatus.PROCESSING;
        }

        return OrderStatus.valueOf(status.toUpperCase());
    }

    private OrderStatus parseStatus(String status) {
        if (status == null || status.isBlank()) {
            return OrderStatus.PROCESSING;
        }

        return OrderStatus.valueOf(status.toUpperCase());
    }

    private List<OrderServiceItem> toServiceItems(List<OrderItemDto> items, Order order) {
        return items.stream()
                .map(item -> {
                    OrderServiceItem entity = new OrderServiceItem();
                    entity.setOrder(order);
                    entity.setName(item.getName());
                    entity.setPrice(item.getPrice());
                    return entity;
                })
                .toList();
    }

    private List<OrderProductItem> toProductsItem(List<OrderItemDto> items, Order order) {
        return items.stream()
                .map(item -> {
                    OrderProductItem entity = new OrderProductItem();
                    entity.setOrder(order);
                    entity.setName(item.getName());
                    entity.setPrice(item.getPrice());
                    return entity;
                })
                .toList();
    }

    private List<OrderDocument> toDocumentItems(List<OrderDocumentRequest> documents, Order order) {
        return documents.stream()
                .map(document -> {
                    OrderDocument entity = new OrderDocument();
                    entity.setOrder(order);
                    entity.setName(document.getName());
                    entity.setType(document.getType());
                    entity.setDocumentDate(document.getDate());
                    entity.setSize(document.getSize());
                    return entity;
                })
                .toList();
    }

    private OrderItemDto toOrderItemDto(OrderServiceItem item) {
        OrderItemDto dto = new OrderItemDto();
        dto.setName(item.getName());
        dto.setPrice(item.getPrice());
        return dto;
    }

    private OrderItemDto toOrderItemDto(OrderProductItem item) {
        OrderItemDto dto = new OrderItemDto();
        dto.setName(item.getName());
        dto.setPrice(item.getPrice());
        return dto;
    }

    private OrderDocumentDto toDocumentDto(OrderDocument document) {
        OrderDocumentDto dto = new OrderDocumentDto();
        dto.setId(document.getId());
        dto.setName(document.getName());
        dto.setType(document.getType());
        dto.setDate(document.getDocumentDate());
        dto.setSize(document.getSize());
        return dto;
    }

}
