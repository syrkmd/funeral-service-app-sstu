package com.funeral.funeralService.controller;

import com.funeral.funeralService.dto.cemetery.response.CemeteryPlotResponse;
import com.funeral.funeralService.dto.cemetery.response.CemeterySectionResponse;
import com.funeral.funeralService.service.CemeteryClientService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.tags.Tag;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@RestController
@RequestMapping("/cemetery")
@RequiredArgsConstructor
@Tag(
        name = "Интеграция с кладбищем",
        description = "Интеграция с cemetery-service для получения секций и свободных участков. " +
                "При integration.cemetery.enabled=false используются локальные демонстрационные данные."
)
public class CemeteryController {

    private final CemeteryClientService service;

    @GetMapping("/plots")
    @Operation(
            summary = "Получить места захоронения",
            description = "Возвращает места выбранной секции. При включённой интеграции данные " +
                    "загружаются из cemetery-service, при integration.cemetery.enabled=false " +
                    "возвращается локальный демонстрационный список."
    )
    @ApiResponse(responseCode = "200", description = "Список мест захоронения возвращён")
    public List<CemeteryPlotResponse> getPlots(@RequestParam String sectionName) {
        return service.getPlots(sectionName);
    }

    @GetMapping("/sections")
    @Operation(
            summary = "Получить секции кладбища",
            description = "Возвращает секции кладбища. При включённой интеграции данные загружаются " +
                    "из cemetery-service, при integration.cemetery.enabled=false возвращаются " +
                    "локальные демонстрационные секции."
    )
    @ApiResponse(responseCode = "200", description = "Список секций возвращён")
    public List<CemeterySectionResponse> getSections() {
        return service.getSections();
    }
}
