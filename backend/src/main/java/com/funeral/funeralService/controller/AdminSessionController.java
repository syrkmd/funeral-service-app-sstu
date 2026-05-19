package com.funeral.funeralService.controller;

import com.funeral.funeralService.dto.AdminLoginRequest;
import com.funeral.funeralService.dto.AdminSessionResponse;
import com.funeral.funeralService.service.AdminSessionService;
import jakarta.servlet.http.HttpSession;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/admin/session")
@RequiredArgsConstructor
public class AdminSessionController {

    private final AdminSessionService service;

    @PostMapping
    public ResponseEntity<AdminSessionResponse> login(@Valid @RequestBody AdminLoginRequest request, HttpSession session) {
        return ResponseEntity.ok(service.login(request, session));
    }

    @GetMapping
    public ResponseEntity<AdminSessionResponse> getSession(HttpSession session) {
        return ResponseEntity.ok(service.getSession(session));
    }

    @DeleteMapping
    public ResponseEntity<Void> logout(HttpSession session) {
        service.logout(session);
        return ResponseEntity.noContent().build();
    }
}
