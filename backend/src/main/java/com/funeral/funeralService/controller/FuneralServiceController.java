package com.funeral.funeralService.controller;

import com.funeral.funeralService.dto.FuneralServiceDto;
import com.funeral.funeralService.dto.ServiceCategoryDto;
import com.funeral.funeralService.service.FuneralServiceCatalogService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@RestController
@RequiredArgsConstructor
public class FuneralServiceController {

    private final FuneralServiceCatalogService service;

    @GetMapping("/funeral-services")
    public ResponseEntity<List<FuneralServiceDto>> getServices() {
        return ResponseEntity.ok(service.getServices());
    }

    @GetMapping("/service-categories")
    public ResponseEntity<List<ServiceCategoryDto>> getCategories() {
        return ResponseEntity.ok(service.getCategories());
    }
}
