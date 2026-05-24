package com.funeral.funeralService.service;

import com.funeral.funeralService.dto.catalog.request.CreateFuneralServiceRequest;
import com.funeral.funeralService.dto.catalog.request.UpdateFuneralServiceRequest;
import com.funeral.funeralService.dto.catalog.response.FuneralServiceDto;
import com.funeral.funeralService.dto.catalog.response.ServiceCategoryDto;
import com.funeral.funeralService.entity.FuneralService;
import com.funeral.funeralService.entity.ServiceCategory;
import com.funeral.funeralService.exception.FuneralServiceNotFoundException;
import com.funeral.funeralService.exception.ServiceCategoryNotFoundException;
import com.funeral.funeralService.repository.FuneralServiceRepository;
import com.funeral.funeralService.repository.ServiceCategoryRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.util.List;

import static com.funeral.funeralService.util.PatchUtils.applyIfPresent;

@Service
@RequiredArgsConstructor
public class FuneralServiceCatalogService {

    private final FuneralServiceRepository funeralServiceRepository;
    private final ServiceCategoryRepository serviceCategoryRepository;

    public List<FuneralServiceDto> getServices(boolean includeInactive) {
        List<FuneralService> services = includeInactive
                ? funeralServiceRepository.findAllByOrderBySortOrderAsc()
                : funeralServiceRepository.findByActiveTrueOrderBySortOrderAsc();

        return services
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

    public FuneralServiceDto createService(CreateFuneralServiceRequest request) {
        FuneralService service = new FuneralService();
        ServiceCategory category = findCategory(request.getCategoryId());

        service.setTitle(request.getTitle());
        service.setDescription(request.getDescription());
        service.setPrice(request.getPrice());
        service.setCategory(category);
        service.setActive(request.getActive() != null ? request.getActive() : true);
        service.setSortOrder(request.getSortOrder());

        return toServiceDto(funeralServiceRepository.save(service));
    }

    public FuneralServiceDto updateService(Long id, UpdateFuneralServiceRequest request) {
        FuneralService service = findService(id);

        applyIfPresent(request.getTitle(), service::setTitle);
        applyIfPresent(request.getDescription(), service::setDescription);
        applyIfPresent(request.getPrice(), service::setPrice);
        applyIfPresent(request.getCategoryId(), categoryId -> service.setCategory(findCategory(categoryId)));
        applyIfPresent(request.getActive(), service::setActive);
        applyIfPresent(request.getSortOrder(), service::setSortOrder);

        return toServiceDto(funeralServiceRepository.save(service));
    }

    public void archiveService(Long id) {
        FuneralService service = findService(id);
        service.setActive(false);
        funeralServiceRepository.save(service);
    }

    private FuneralServiceDto toServiceDto(FuneralService service) {
        return new FuneralServiceDto(
                service.getId(),
                service.getTitle(),
                service.getDescription(),
                service.getPrice(),
                toCategoryDto(service.getCategory()),
                service.getActive(),
                service.getSortOrder()
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

    private FuneralService findService(Long id) {
        return funeralServiceRepository.findById(id)
                .orElseThrow(() -> new FuneralServiceNotFoundException(id));
    }

    private ServiceCategory findCategory(Long id) {
        return serviceCategoryRepository.findById(id)
                .orElseThrow(() -> new ServiceCategoryNotFoundException(id));
    }
}
