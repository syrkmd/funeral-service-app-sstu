package com.funeral.funeralService.repository;

import com.funeral.funeralService.entity.Order;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface OrderRepository extends JpaRepository<Order, String> {
    List<Order> findByClientPhone(String clientPhone);
}
