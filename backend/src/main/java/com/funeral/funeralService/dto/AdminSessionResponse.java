package com.funeral.funeralService.dto;

import lombok.AllArgsConstructor;
import lombok.Data;

@Data
@AllArgsConstructor
public class AdminSessionResponse {

    private Boolean isAuthenticated;
}
