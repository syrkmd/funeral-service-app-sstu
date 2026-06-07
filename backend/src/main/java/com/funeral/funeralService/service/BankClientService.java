package com.funeral.funeralService.service;

import com.funeral.funeralService.dto.bank.request.BankTransferRequest;
import com.funeral.funeralService.dto.order.request.PayOrderRequest;
import com.funeral.funeralService.exception.BankPaymentException;
import com.funeral.funeralService.exception.BankServiceUnavailableException;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;
import org.springframework.web.client.HttpClientErrorException;
import org.springframework.web.client.HttpServerErrorException;
import org.springframework.web.client.ResourceAccessException;
import org.springframework.web.client.RestClient;

import java.math.BigDecimal;

@Service
public class BankClientService {

    private final RestClient restClient;
    private final String receiverCardNumber;

    public BankClientService(
            @Value("${bank.service.url}") String bankServiceUrl,
            @Value("${bank.receiver-card-number}") String receiverCardNumber
    ) {
        this.restClient = RestClient.builder()
                .baseUrl(bankServiceUrl)
                .build();
        this.receiverCardNumber = receiverCardNumber;
    }

    public void pay(
            String orderId,
            BigDecimal amount,
            PayOrderRequest request
    ) {
        BankTransferRequest transfer = new BankTransferRequest(
                request.getCardNumber(),
                request.getCvv(),
                request.getExpiryDate(),
                receiverCardNumber,
                amount,
                "Оплата заказа " + orderId
        );

        try {
            restClient.post()
                    .uri("/api/transfer/execute")
                    .body(transfer)
                    .retrieve()
                    .toBodilessEntity();
        } catch (HttpClientErrorException exception) {
            throw new BankPaymentException();
        } catch (HttpServerErrorException | ResourceAccessException exception) {
            throw new BankServiceUnavailableException();
        }
    }
}
