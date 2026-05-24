package com.funeral.funeralService.service;

import com.funeral.funeralService.dto.cemetery.request.ReservePlotRequest;
import com.funeral.funeralService.dto.order.request.*;
import com.funeral.funeralService.dto.order.response.OrderResponse;
import com.funeral.funeralService.entity.Order;
import com.funeral.funeralService.entity.OrderDocument;
import com.funeral.funeralService.entity.OrderProductItem;
import com.funeral.funeralService.entity.OrderServiceItem;
import com.funeral.funeralService.exception.CatalogProductNotFoundException;
import com.funeral.funeralService.exception.FuneralServiceNotFoundException;
import com.funeral.funeralService.exception.OrderDocumentNotFoundException;
import com.funeral.funeralService.exception.OrderNotFoundException;
import com.funeral.funeralService.mapper.OrderMapper;
import com.funeral.funeralService.repository.CatalogProductRepository;
import com.funeral.funeralService.repository.FuneralServiceRepository;
import com.funeral.funeralService.repository.OrderRepository;
import lombok.AllArgsConstructor;
import org.springframework.stereotype.Service;

import java.math.BigDecimal;
import java.util.List;
import java.util.UUID;

@Service
@AllArgsConstructor
public class OrderService {

    private OrderMapper mapper;
    private OrderRepository repository;
    private CemeteryClientService cemeteryClientService;
    private FuneralServiceRepository funeralServiceRepository;
    private CatalogProductRepository catalogProductRepository;

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

        if (request.getCemeteryPlotId() != null) {
            ReservePlotRequest reservePlotRequest = new ReservePlotRequest();
            reservePlotRequest.setPlotId(request.getCemeteryPlotId());
            reservePlotRequest.setOrderId(order.getId());
            cemeteryClientService.reservePlot(reservePlotRequest);
        }

        Order savedOrder = repository.save(order);

        return mapper.toResponse(savedOrder);
    }

    public OrderResponse updateStatus(String id, UpdateOrderStatusRequest request) {
        Order order = findOrder(id);

        order.setStatus(mapper.toOrderStatus(request.getStatus()));

        Order savedOrder = repository.save(order);

        return mapper.toResponse(savedOrder);
    }

    public OrderResponse updatePayment(String id, UpdateOrderPaymentRequest request) {
        Order order = findOrder(id);

        order.setPaid(request.getIsPaid());

        Order savedOrder = repository.save(order);

        return mapper.toResponse(savedOrder);
    }

    public OrderResponse updateClient(String id, UpdateOrderClientRequest request) {
        Order order = findOrder(id);

        order.setClientName(request.getName());
        order.setClientPhone(request.getPhone());
        order.setClientEmail(request.getEmail());

        Order savedOrder = repository.save(order);

        return mapper.toResponse(savedOrder);
    }

    public OrderResponse updateDeceased(String id, UpdateOrderDeceasedRequest request) {
        Order order = findOrder(id);

        order.setDeceasedName(request.getName());
        order.setDeceasedDateOfBirth(request.getDateOfBirth());
        order.setDeceasedDateOfDeath(request.getDateOfDeath());

        Order savedOrder = repository.save(order);

        return mapper.toResponse(savedOrder);
    }

    public OrderResponse replaceServices(String id, ReplaceOrderServicesRequest request) {
        Order order = findOrder(id);

        order.getServices().clear();
        order.getServices().addAll(toServiceItems(request.getServiceIds(), order));
        recalculateTotal(order);

        Order savedOrder = repository.save(order);

        return mapper.toResponse(savedOrder);
    }

    public OrderResponse replaceProducts(String id, ReplaceOrderProductsRequest request) {
        Order order = findOrder(id);

        order.getProducts().clear();
        order.getProducts().addAll(toProductItems(request.getProductIds(), order));
        recalculateTotal(order);

        Order savedOrder = repository.save(order);

        return mapper.toResponse(savedOrder);
    }

    public OrderResponse updateDate(String id, UpdateOrderDateRequest request) {
        Order order = findOrder(id);

        order.setOrderDate(request.getDate());

        Order savedOrder = repository.save(order);

        return mapper.toResponse(savedOrder);
    }

    public OrderResponse updateDiscount(String id, UpdateOrderDiscountRequest request) {
        Order order = findOrder(id);

        order.setTotalAmount(order.getTotalAmount().subtract(request.getDiscountAmount()));

        Order savedOrder = repository.save(order);

        return mapper.toResponse(savedOrder);
    }

    public OrderResponse updateCeremony(String id, UpdateOrderCeremonyRequest request) {
        Order order = findOrder(id);

        order.setServiceDate(request.getServiceDate());
        order.setServiceTime(request.getServiceTime());
        order.setServiceAddress(request.getServiceAddress());
        order.setCemetery(request.getCemetery());
        order.setCemeteryNotes(request.getCemeteryNotes());

        Order savedOrder = repository.save(order);

        return mapper.toResponse(savedOrder);
    }

    public void deleteOrder(String id) {
        Order order = findOrder(id);

        repository.deleteById(order.getId());
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

    private void recalculateTotal(Order order) {
        BigDecimal servicesTotal = order.getServices().stream()
                .map(OrderServiceItem::getPrice)
                .reduce(BigDecimal.ZERO, BigDecimal::add);

        BigDecimal productsTotal = order.getProducts().stream()
                .map(OrderProductItem::getPrice)
                .reduce(BigDecimal.ZERO, BigDecimal::add);

        order.setTotalAmount(servicesTotal.add(productsTotal));
    }

    private List<OrderServiceItem> toServiceItems(List<Long> serviceIds, Order order) {
        return serviceIds.stream()
                .map(serviceId -> {
                    com.funeral.funeralService.entity.FuneralService service = funeralServiceRepository.findById(serviceId)
                            .filter(item -> Boolean.TRUE.equals(item.getActive()))
                            .orElseThrow(() -> new FuneralServiceNotFoundException(serviceId));

                    OrderServiceItem item = new OrderServiceItem();
                    item.setOrder(order);
                    item.setName(service.getTitle());
                    item.setPrice(service.getPrice());
                    return item;
                })
                .toList();
    }

    private List<OrderProductItem> toProductItems(List<Long> productIds, Order order) {
        return productIds.stream()
                .map(productId -> {
                    com.funeral.funeralService.entity.CatalogProduct product = catalogProductRepository.findById(productId)
                            .filter(item -> Boolean.TRUE.equals(item.getActive()))
                            .orElseThrow(() -> new CatalogProductNotFoundException(productId));

                    OrderProductItem item = new OrderProductItem();
                    item.setOrder(order);
                    item.setName(product.getTitle());
                    item.setPrice(product.getPrice());
                    return item;
                })
                .toList();
    }
}
