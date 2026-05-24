package com.funeral.funeralService.controller;

import com.funeral.funeralService.dto.cemetery.response.CemeteryPlotResponse;
import com.funeral.funeralService.service.CemeteryClientService;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@RestController
@RequestMapping("/cemetery")
@RequiredArgsConstructor
public class CemeteryController {

    private final CemeteryClientService service;

    @GetMapping("/plots")
    public List<CemeteryPlotResponse> getPlots() {
        return service.getPlots();
    }
}
