package com.funeral.funeralService.dto.bank.request;

import lombok.AllArgsConstructor;
import lombok.Data;

import java.math.BigDecimal;

@Data
@AllArgsConstructor
public class BankTransferRequest {

    private String fromCardNumber;
    private String fromCvv;
    private String fromExpiryDate;
    private String toCardNumber;
    private BigDecimal amount;
    private String description;
}