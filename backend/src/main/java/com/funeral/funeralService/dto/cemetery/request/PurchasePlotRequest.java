package com.funeral.funeralService.dto.cemetery.request;

import lombok.Data;

import java.time.LocalDate;

@Data
public class PurchasePlotRequest {

    private String plotCode;

    private String ownerName;

    private String phone;

    private LocalDate startDate;
}