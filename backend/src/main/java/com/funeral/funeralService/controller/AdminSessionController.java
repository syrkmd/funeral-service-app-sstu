package com.funeral.funeralService.controller;

import com.funeral.funeralService.dto.admin.request.AdminLoginRequest;
import com.funeral.funeralService.dto.admin.response.AdminSessionResponse;
import com.funeral.funeralService.service.AdminSessionService;
import jakarta.servlet.http.HttpSession;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/admin/session")
@RequiredArgsConstructor
public class AdminSessionController {

    private final AdminSessionService service;

    @PostMapping
    public AdminSessionResponse login(@Valid @RequestBody AdminLoginRequest request, HttpSession session) {
        return service.login(request, session);
    }

    @GetMapping
    public AdminSessionResponse getSession(HttpSession session) {
        return service.getSession(session);
    }

    @DeleteMapping
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void logout(HttpSession session) {
        service.logout(session);
    }
}
