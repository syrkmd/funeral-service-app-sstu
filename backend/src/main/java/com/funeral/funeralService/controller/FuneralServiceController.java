package com.funeral.funeralService.controller;

import com.funeral.funeralService.dto.catalog.request.CreateFuneralServiceRequest;
import com.funeral.funeralService.dto.catalog.request.UpdateFuneralServiceRequest;
import com.funeral.funeralService.dto.catalog.response.FuneralServiceDto;
import com.funeral.funeralService.dto.catalog.response.ServiceCategoryDto;
import com.funeral.funeralService.service.FuneralServiceCatalogService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequiredArgsConstructor
public class FuneralServiceController {

    private final FuneralServiceCatalogService service;

    @GetMapping("/funeral-services")
    public List<FuneralServiceDto> getServices(@RequestParam(defaultValue = "false") boolean includeInactive) {
        return service.getServices(includeInactive);
    }

    @PostMapping("/funeral-services")
    @ResponseStatus(HttpStatus.CREATED)
    public FuneralServiceDto createService(@Valid @RequestBody CreateFuneralServiceRequest request) {
        return service.createService(request);
    }

    @PatchMapping("/funeral-services/{id}")
    public FuneralServiceDto updateService(@PathVariable Long id, @Valid @RequestBody UpdateFuneralServiceRequest request) {
        return service.updateService(id, request);
    }

    @DeleteMapping("/funeral-services/{id}")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void archiveService(@PathVariable Long id) {
        service.archiveService(id);
    }

    @GetMapping("/service-categories")
    public List<ServiceCategoryDto> getCategories() {
        return service.getCategories();
    }
}
