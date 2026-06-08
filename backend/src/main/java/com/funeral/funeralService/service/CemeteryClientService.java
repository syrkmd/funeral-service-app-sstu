package com.funeral.funeralService.service;

import com.funeral.funeralService.dto.cemetery.request.CemeteryLoginRequest;
import com.funeral.funeralService.dto.cemetery.request.PurchasePlotRequest;
import com.funeral.funeralService.dto.cemetery.response.*;
import com.funeral.funeralService.exception.CemeteryIntegrationException;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpHeaders;
import org.springframework.http.client.SimpleClientHttpRequestFactory;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestClient;

import java.util.Arrays;
import java.util.List;

@Service
public class CemeteryClientService {

    private final RestClient restClient;
    private final String login;
    private final String password;
    private final boolean cemeteryIntegrationEnabled;

    public CemeteryClientService(
            @Value("${cemetery.service.url}") String cemeteryServiceUrl,
            @Value("${cemetery.service.login}") String login,
            @Value("${cemetery.service.password}") String password,
            @Value("${integration.cemetery.enabled}") boolean cemeteryIntegrationEnabled
    ) {

        SimpleClientHttpRequestFactory factory = new SimpleClientHttpRequestFactory();

        factory.setConnectTimeout(5000);
        factory.setReadTimeout(7000);

        this.restClient = RestClient.builder()
                .baseUrl(cemeteryServiceUrl)
                .requestFactory(factory)
                .build();
        this.login = login;
        this.password = password;
        this.cemeteryIntegrationEnabled = cemeteryIntegrationEnabled;
    }

    public List<CemeteryPlotResponse> getPlots(String sectionName) {
        if (!cemeteryIntegrationEnabled) {
            return getDemoPlots(sectionName);
        }

        String token = login();

        CemeteryPlotApiResponse[] plots = restClient.get()
                .uri(uriBuilder -> uriBuilder
                        .path("/cemetery/plots/free")
                        .queryParam("sectionName", sectionName)
                        .build())
                .header(HttpHeaders.AUTHORIZATION, "Bearer " + token)
                .retrieve()
                .body(CemeteryPlotApiResponse[].class);

        if (plots == null) {
            return List.of();
        }

        return Arrays.stream(plots)
                .map(plot -> new CemeteryPlotResponse(
                        plot.getId(),
                        plot.getCode(),
                        "FREE".equals(plot.getStatus())
                ))
                .toList();
    }

    public List<CemeterySectionResponse> getSections() {
        if (!cemeteryIntegrationEnabled) {
            return getDemoSections();
        }

        String token = login();

        CemeterySectionResponse[] sections = restClient.get()
                .uri("/cemetery/sections")
                .header(HttpHeaders.AUTHORIZATION, "Bearer " + token)
                .retrieve()
                .body(CemeterySectionResponse[].class);

        return sections == null ? List.of() : Arrays.asList(sections);
    }

    public CemeteryContractResponse reservePlot(PurchasePlotRequest request) {
        String token = login();

        CemeteryContractResponse response = restClient.post()
                .uri("/cemetery/contracts/purchase")
                .header(HttpHeaders.AUTHORIZATION, "Bearer " + token)
                .body(request)
                .retrieve()
                .body(CemeteryContractResponse.class);

        if (response == null) {
            throw new CemeteryIntegrationException(
                    "Cemetery-service не вернул данные договора"
            );
        }

        return response;
    }

    private String login() {
        CemeteryLoginResponse response = restClient.post()
                .uri("/auth/login")
                .body(new CemeteryLoginRequest(login, password))
                .retrieve()
                .body(CemeteryLoginResponse.class);

        if (response == null || response.getToken() == null) {
            throw new CemeteryIntegrationException(
                    "Cemetery-service не вернул токен"
            );
        }

        return response.getToken();
    }

    private List<CemeterySectionResponse> getDemoSections() {
        return List.of(
                createDemoSection(1L, "A", "Центральная секция"),
                createDemoSection(2L, "B", "Северная секция"),
                createDemoSection(3L, "C", "Южная секция")
        );
    }

    private List<CemeteryPlotResponse> getDemoPlots(String sectionName) {
        String normalizedSection = sectionName == null
                ? ""
                : sectionName.trim().toUpperCase();

        return switch (normalizedSection) {
            case "A" -> List.of(
                    new CemeteryPlotResponse(1L, "A-101", true),
                    new CemeteryPlotResponse(2L, "A-102", true),
                    new CemeteryPlotResponse(3L, "A-103", false)
            );
            case "B" -> List.of(
                    new CemeteryPlotResponse(4L, "B-201", true),
                    new CemeteryPlotResponse(5L, "B-202", false),
                    new CemeteryPlotResponse(6L, "B-203", true)
            );
            case "C" -> List.of(
                    new CemeteryPlotResponse(7L, "C-301", true),
                    new CemeteryPlotResponse(8L, "C-302", true),
                    new CemeteryPlotResponse(9L, "C-303", false)
            );
            default -> List.of();
        };
    }

    private CemeterySectionResponse createDemoSection(
            Long id,
            String name,
            String description
    ) {
        CemeterySectionResponse section = new CemeterySectionResponse();
        section.setId(id);
        section.setName(name);
        section.setDescription(description);
        return section;
    }
}
