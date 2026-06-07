package com.funeral.funeralService.controller;

import com.funeral.funeralService.dto.catalog.request.CreateFuneralServiceRequest;
import com.funeral.funeralService.dto.catalog.request.UpdateFuneralServiceRequest;
import com.funeral.funeralService.dto.catalog.response.FuneralServiceDto;
import com.funeral.funeralService.dto.catalog.response.ServiceCategoryDto;
import com.funeral.funeralService.service.AdminSessionService;
import com.funeral.funeralService.service.FuneralServiceCatalogService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import jakarta.servlet.http.HttpSession;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequiredArgsConstructor
@Tag(name = "Каталог услуг", description = "Каталог ритуальных услуг для выбора при заказе и админского управления")
public class FuneralServiceController {

    private final FuneralServiceCatalogService service;
    private final AdminSessionService adminSessionService;

    @GetMapping("/funeral-services")
    @Operation(summary = "Получить ритуальные услуги", description = "Возвращает активные услуги для выбора при оформлении заказа. Админка может передать includeInactive=true, чтобы увидеть архивные услуги.")
    @ApiResponse(responseCode = "200", description = "Список услуг возвращён")
    public List<FuneralServiceDto> getServices(
            @Parameter(description = "Админский флаг для отображения архивных услуг", example = "true")
            @RequestParam(defaultValue = "false") boolean includeInactive
    ) {
        return service.getServices(includeInactive);
    }

    @PostMapping("/funeral-services")
    @ResponseStatus(HttpStatus.CREATED)
    @Operation(summary = "Создать ритуальную услугу", description = "Админский endpoint для создания новой услуги или пакета услуг.")
    @ApiResponse(responseCode = "201", description = "Услуга создана")
    @ApiResponse(responseCode = "400", description = "Ошибка валидации")
    @ApiResponse(responseCode = "404", description = "Категория услуги не найдена")
    public FuneralServiceDto createService(@Valid @RequestBody CreateFuneralServiceRequest request, HttpSession session) {
        adminSessionService.requireAuthenticated(session);
        return service.createService(request);
    }

    @PatchMapping("/funeral-services/{id}")
    @Operation(summary = "Изменить ритуальную услугу", description = "Админский endpoint для частичного изменения услуги. Также позволяет восстановить архивную услугу через active=true.")
    @ApiResponse(responseCode = "200", description = "Услуга изменена")
    @ApiResponse(responseCode = "400", description = "Ошибка валидации")
    @ApiResponse(responseCode = "404", description = "Услуга или категория не найдены")
    public FuneralServiceDto updateService(
            @Parameter(description = "Id услуги", example = "1") @PathVariable Long id,
            @Valid @RequestBody UpdateFuneralServiceRequest request,
            HttpSession session
    ) {
        adminSessionService.requireAuthenticated(session);
        return service.updateService(id, request);
    }

    @DeleteMapping("/funeral-services/{id}")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    @Operation(summary = "Архивировать ритуальную услугу", description = "Админский soft delete endpoint. Услуга не удаляется физически, а переводится в active=false.")
    @ApiResponse(responseCode = "204", description = "Услуга архивирована")
    @ApiResponse(responseCode = "404", description = "Услуга не найдена")
    public void archiveService(@Parameter(description = "Id услуги", example = "1") @PathVariable Long id, HttpSession session) {
        adminSessionService.requireAuthenticated(session);
        service.archiveService(id);
    }

    @GetMapping("/service-categories")
    @Operation(summary = "Получить категории услуг", description = "Возвращает активные категории ритуальных услуг.")
    @ApiResponse(responseCode = "200", description = "Категории услуг возвращены")
    public List<ServiceCategoryDto> getCategories() {
        return service.getCategories();
    }
}
