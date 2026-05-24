package com.funeral.funeralService.service;

import com.funeral.funeralService.dto.cemetery.request.ReservePlotRequest;
import com.funeral.funeralService.dto.cemetery.response.CemeteryPlotResponse;
import com.funeral.funeralService.dto.cemetery.response.ReservePlotResponse;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.UUID;

@Service
public class CemeteryClientService {

    public List<CemeteryPlotResponse> getPlots() {
        return List.of(
                new CemeteryPlotResponse(1L, "A-12", true),
                new CemeteryPlotResponse(2L, "A-13", true),
                new CemeteryPlotResponse(3L, "A-14", false)
        );
    }

    public ReservePlotResponse reservePlot(ReservePlotRequest request) {
        return new ReservePlotResponse(true, "RES-" + UUID.randomUUID()
                .toString()
                .replace("-", "")
                .substring(0, 8)
                .toUpperCase());
    }
}
