package com.funeral.funeralService.dto.admin.response;

import lombok.AllArgsConstructor;
import lombok.Data;

@Data
@AllArgsConstructor
public class AdminSessionResponse {

    private Boolean isAuthenticated;
}
