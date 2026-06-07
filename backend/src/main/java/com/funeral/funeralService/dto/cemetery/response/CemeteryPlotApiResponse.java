package com.funeral.funeralService.dto.cemetery.response;

import lombok.Data;

@Data
public class CemeteryPlotApiResponse {

    private Long id;

    private String code;

    private String sectionName;

    private String status;

    private Double areaM2;
}
