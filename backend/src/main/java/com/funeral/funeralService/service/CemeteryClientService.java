package com.funeral.funeralService.service;

import com.funeral.funeralService.dto.cemetery.request.CemeteryLoginRequest;
import com.funeral.funeralService.dto.cemetery.request.PurchasePlotRequest;
import com.funeral.funeralService.dto.cemetery.response.*;
import com.funeral.funeralService.exception.CemeteryIntegrationException;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpHeaders;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestClient;

import java.util.Arrays;
import java.util.List;

@Service
public class CemeteryClientService {

    private final RestClient restClient;
    private final String login;
    private final String password;

    public CemeteryClientService(
            @Value("${cemetery.service.url}") String cemeteryServiceUrl,
            @Value("${cemetery.service.login}") String login,
            @Value("${cemetery.service.password}") String password
    ) {
        this.restClient = RestClient.builder()
                .baseUrl(cemeteryServiceUrl)
                .build();
        this.login = login;
        this.password = password;
    }

    public List<CemeteryPlotResponse> getPlots(String sectionName) {

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
}
