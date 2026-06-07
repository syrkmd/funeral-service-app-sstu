package com.funeral.funeralService.dto.cemetery.response;

import lombok.Data;

@Data
public class CemeteryLoginResponse {

    private String token;

    private String role;
}