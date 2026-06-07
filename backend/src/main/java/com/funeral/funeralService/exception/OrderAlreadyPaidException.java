package com.funeral.funeralService.exception;

public class OrderAlreadyPaidException extends RuntimeException {

    public OrderAlreadyPaidException(String orderId) {
        super("Заказ уже оплачен: " + orderId);
    }
}