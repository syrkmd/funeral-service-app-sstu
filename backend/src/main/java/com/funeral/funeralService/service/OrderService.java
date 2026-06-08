package com.funeral.funeralService.service;

import com.funeral.funeralService.dto.cemetery.request.PurchasePlotRequest;
import com.funeral.funeralService.dto.order.request.*;
import com.funeral.funeralService.dto.order.response.OrderResponse;
import com.funeral.funeralService.entity.*;
import com.funeral.funeralService.exception.*;
import com.funeral.funeralService.mapper.OrderMapper;
import com.funeral.funeralService.repository.CatalogProductRepository;
import com.funeral.funeralService.repository.FuneralServiceRepository;
import com.funeral.funeralService.repository.OrderRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.util.List;
import java.util.UUID;

@Service
@RequiredArgsConstructor
public class OrderService {

    private final OrderMapper mapper;
    private final OrderRepository repository;
    private final CemeteryClientService cemeteryClientService;
    private final BankClientService bankService;
    private final FuneralServiceRepository funeralServiceRepository;
    private final CatalogProductRepository catalogProductRepository;

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
        order.setStatus(OrderStatus.PROCESSING);
        order.setPaid(false);
        order.setServices(toServiceItemsByCatalog(request.getServices(), order));
        order.setProducts(toProductItemsByCatalog(request.getProducts(), order));
        recalculateTotal(order);

        if (request.getCemeteryPlotCode() != null
                && !request.getCemeteryPlotCode().isBlank()) {

            if (request.getServiceDate() == null) {
                throw new InvalidCemeteryReservationException();
            }

            PurchasePlotRequest purchaseRequest = new PurchasePlotRequest();
            purchaseRequest.setPlotCode(request.getCemeteryPlotCode());
            purchaseRequest.setOwnerName(request.getClient().getName());
            purchaseRequest.setPhone(request.getClient().getPhone());
            purchaseRequest.setStartDate(request.getServiceDate());

            cemeteryClientService.reservePlot(purchaseRequest);
        }

        Order savedOrder = repository.save(order);

        return mapper.toResponse(savedOrder);
    }

    @Transactional
    public OrderResponse updateStatus(String id, UpdateOrderStatusRequest request) {
        Order order = findOrder(id);

        order.setStatus(mapper.toOrderStatus(request.getStatus()));

        Order savedOrder = repository.save(order);

        return mapper.toResponse(savedOrder);
    }

    @Transactional
    public OrderResponse updatePayment(String id, UpdateOrderPaymentRequest request) {
        Order order = findOrder(id);

        order.setPaid(request.getIsPaid());

        Order savedOrder = repository.save(order);

        return mapper.toResponse(savedOrder);
    }

    @Transactional
    public OrderResponse updateClient(String id, UpdateOrderClientRequest request) {
        Order order = findOrder(id);

        order.setClientName(request.getName());
        order.setClientPhone(request.getPhone());
        order.setClientEmail(request.getEmail());

        Order savedOrder = repository.save(order);

        return mapper.toResponse(savedOrder);
    }

    @Transactional
    public OrderResponse updateDeceased(String id, UpdateOrderDeceasedRequest request) {
        Order order = findOrder(id);

        order.setDeceasedName(request.getName());
        order.setDeceasedDateOfBirth(request.getDateOfBirth());
        order.setDeceasedDateOfDeath(request.getDateOfDeath());

        Order savedOrder = repository.save(order);

        return mapper.toResponse(savedOrder);
    }

    @Transactional
    public OrderResponse replaceServices(String id, ReplaceOrderServicesRequest request) {
        Order order = findOrder(id);

        order.getServices().clear();
        order.getServices().addAll(toServiceItems(request.getServiceIds(), order));
        recalculateTotal(order);

        Order savedOrder = repository.save(order);

        return mapper.toResponse(savedOrder);
    }

    @Transactional
    public OrderResponse replaceProducts(String id, ReplaceOrderProductsRequest request) {
        Order order = findOrder(id);

        order.getProducts().clear();
        order.getProducts().addAll(toProductItems(request.getProductIds(), order));
        recalculateTotal(order);

        Order savedOrder = repository.save(order);

        return mapper.toResponse(savedOrder);
    }

    @Transactional
    public OrderResponse updateDate(String id, UpdateOrderDateRequest request) {
        Order order = findOrder(id);

        order.setOrderDate(request.getDate());

        Order savedOrder = repository.save(order);

        return mapper.toResponse(savedOrder);
    }

    @Transactional
    public OrderResponse updateDiscount(String id, UpdateOrderDiscountRequest request) {
        Order order = findOrder(id);

        if (Boolean.TRUE.equals(order.getPaid())) {
            throw new OrderAlreadyPaidException(order.getId());
        }

        if (request.getDiscountAmount().compareTo(order.getTotalAmount()) > 0) {
            throw new InvalidDiscountException();
        }

        order.setTotalAmount(order.getTotalAmount().subtract(request.getDiscountAmount()));

        Order savedOrder = repository.save(order);

        return mapper.toResponse(savedOrder);
    }

    @Transactional
    public OrderResponse updateCeremony(String id, UpdateOrderCeremonyRequest request) {
        Order order = findOrder(id);

        order.setServiceDate(request.getServiceDate());
        order.setServiceTime(request.getServiceTime());
        order.setServiceAddress(request.getServiceAddress());
        order.setCemeteryNotes(request.getCemeteryNotes());

        Order savedOrder = repository.save(order);

        return mapper.toResponse(savedOrder);
    }

    public void deleteOrder(String id) {
        Order order = findOrder(id);

        repository.deleteById(order.getId());
    }

    @Transactional
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

    @Transactional
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

    public OrderResponse payOrder(String id, PayOrderRequest request) {
        Order order = findOrder(id);

        if (Boolean.TRUE.equals(order.getPaid())) {
            throw new OrderAlreadyPaidException(order.getId());
        }

        bankService.pay(
                order.getId(),
                order.getTotalAmount(),
                request
        );

        order.setPaid(true);

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
                    FuneralService service = funeralServiceRepository.findById(serviceId)
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
                    CatalogProduct product = catalogProductRepository.findById(productId)
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

    private List<OrderServiceItem> toServiceItemsByCatalog(List<OrderCatalogSelectionRequest> requestedItems, Order order) {
        return requestedItems.stream()
                .map(requestedItem -> {
                    FuneralService service =
                            funeralServiceRepository.findFirstByTitleAndActiveTrue(requestedItem.getName())
                                    .orElseThrow(() -> new FuneralServiceNotFoundException(requestedItem.getName()));

                    OrderServiceItem item = new OrderServiceItem();
                    item.setOrder(order);
                    item.setName(service.getTitle());
                    item.setPrice(service.getPrice());
                    return item;
                })
                .toList();
    }

    private List<OrderProductItem> toProductItemsByCatalog(List<OrderCatalogSelectionRequest> requestedItems, Order order) {
        return requestedItems.stream()
                .map(requestedItem -> {
                    CatalogProduct product =
                            catalogProductRepository.findFirstByTitleAndActiveTrue(requestedItem.getName())
                                    .orElseThrow(() -> new CatalogProductNotFoundException(requestedItem.getName()));

                    OrderProductItem item = new OrderProductItem();
                    item.setOrder(order);
                    item.setName(product.getTitle());
                    item.setPrice(product.getPrice());
                    return item;
                })
                .toList();
    }
}
