package com.funeral.funeralService.entity;

import jakarta.persistence.*;
import lombok.Data;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.util.List;

@Entity
@Data
@Table(name = "orders")
public class Order {

    @Id
    @Column(name = "id", length = 32, nullable = false)
    private String id;

    @Column(name = "order_date", nullable = false)
    private LocalDate orderDate;

    @Enumerated(EnumType.STRING)
    @Column(name = "status", nullable = false)
    private OrderStatus status;

    @Column(name = "total_amount", nullable = false)
    private BigDecimal totalAmount;

    @Column(name = "paid", nullable = false)
    private Boolean paid;

    @Column(name = "client_name", nullable = false)
    private String clientName;

    @Column(name = "client_phone", nullable = false)
    private String clientPhone;

    @Column(name = "client_email")
    private String clientEmail;

    @Column(name = "deceased_name", nullable = false)
    private String deceasedName;

    @Column(name = "deceased_date_of_birth", nullable = false)
    private LocalDate deceasedDateOfBirth;

    @Column(name = "deceased_date_of_death", nullable = false)
    private LocalDate deceasedDateOfDeath;

    @OneToMany(mappedBy = "order", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<OrderServiceItem> services;

    @OneToMany(mappedBy = "order", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<OrderProductItem> products;

    @OneToMany(mappedBy = "order", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<OrderDocument> documents;
}
