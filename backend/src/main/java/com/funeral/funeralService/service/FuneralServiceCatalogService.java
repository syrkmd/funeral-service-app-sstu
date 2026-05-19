package com.funeral.funeralService.service;

import com.funeral.funeralService.dto.FuneralServiceDto;
import com.funeral.funeralService.dto.ServiceCategoryDto;
import com.funeral.funeralService.entity.FuneralService;
import com.funeral.funeralService.entity.ServiceCategory;
import com.funeral.funeralService.repository.FuneralServiceRepository;
import com.funeral.funeralService.repository.ServiceCategoryRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
@RequiredArgsConstructor
public class FuneralServiceCatalogService {

    private final FuneralServiceRepository funeralServiceRepository;
    private final ServiceCategoryRepository serviceCategoryRepository;

    public List<FuneralServiceDto> getServices() {
        return funeralServiceRepository.findByActiveTrueOrderBySortOrderAsc()
                .stream()
                .map(this::toServiceDto)
                .toList();
    }

    public List<ServiceCategoryDto> getCategories() {
        return serviceCategoryRepository.findByActiveTrueOrderBySortOrderAsc()
                .stream()
                .map(this::toCategoryDto)
                .toList();
    }

    private FuneralServiceDto toServiceDto(FuneralService service) {
        return new FuneralServiceDto(
                service.getId(),
                service.getTitle(),
                service.getDescription(),
                service.getPrice(),
                toCategoryDto(service.getCategory())
        );
    }

    private ServiceCategoryDto toCategoryDto(ServiceCategory category) {
        if (category == null) {
            return null;
        }

        return new ServiceCategoryDto(
                category.getId(),
                category.getName()
        );
    }
}
