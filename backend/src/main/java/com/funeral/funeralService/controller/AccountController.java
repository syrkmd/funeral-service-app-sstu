package com.funeral.funeralService.controller;

import com.funeral.funeralService.dto.AccountSessionResponse;
import com.funeral.funeralService.dto.AccountVerifyRequest;
import com.funeral.funeralService.service.AccountService;
import jakarta.servlet.http.HttpSession;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/users/account")
@RequiredArgsConstructor
public class AccountController {

    private final AccountService service;

    @PostMapping("/verify")
    public ResponseEntity<AccountSessionResponse> verify(@Valid @RequestBody AccountVerifyRequest request, HttpSession session) {
        return ResponseEntity.ok(service.verify(request, session));
    }

    @GetMapping("/session")
    public ResponseEntity<AccountSessionResponse> getSession(HttpSession session) {
        return ResponseEntity.ok(service.getSession(session));
    }

    @PostMapping("/logout")
    public ResponseEntity<Void> logout(HttpSession session) {
        service.logout(session);
        return ResponseEntity.noContent().build();
    }
}
