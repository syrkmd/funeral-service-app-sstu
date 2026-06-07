package com.funeral.funeralService.dto.cemetery.response;

import lombok.Data;

import java.time.LocalDate;

@Data
public class CemeteryContractResponse {
    private Long id;

    private String plotCode;

    private String ownerName;

    private String phone;

    private LocalDate startDate;

    private LocalDate endDate;

    private String status;
}