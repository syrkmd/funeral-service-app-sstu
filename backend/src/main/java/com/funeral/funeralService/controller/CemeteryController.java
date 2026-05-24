package com.funeral.funeralService.controller;

import com.funeral.funeralService.dto.cemetery.response.CemeteryPlotResponse;
import com.funeral.funeralService.service.CemeteryClientService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.tags.Tag;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@RestController
@RequestMapping("/cemetery")
@RequiredArgsConstructor
@Tag(name = "Интеграция с кладбищем", description = "Заглушка интеграции с сервисом кладбища, подготовленная для будущей замены на внешний сервис")
public class CemeteryController {

    private final CemeteryClientService service;

    @GetMapping("/plots")
    @Operation(summary = "Получить места захоронения", description = "Возвращает доступные и недоступные места захоронения из fake-сервиса интеграции.")
    @ApiResponse(responseCode = "200", description = "Список мест захоронения возвращён")
    public List<CemeteryPlotResponse> getPlots() {
        return service.getPlots();
    }
}
