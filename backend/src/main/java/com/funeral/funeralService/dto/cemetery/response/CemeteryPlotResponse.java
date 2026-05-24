package com.funeral.funeralService.dto.cemetery.response;

import lombok.AllArgsConstructor;
import lombok.Data;

@Data
@AllArgsConstructor
public class CemeteryPlotResponse {

    private Long id;

    private String label;

    private Boolean available;
}
