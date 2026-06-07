package com.funeral.funeralService.exception;

public class BankPaymentException extends RuntimeException {

    public BankPaymentException() {
        super("Банк отклонил платёж. Проверьте данные карты и доступный баланс.");
    }
}
