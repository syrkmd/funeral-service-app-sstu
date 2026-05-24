package com.funeral.funeralService.controller;

import com.funeral.funeralService.dto.account.response.AccountSessionResponse;
import com.funeral.funeralService.dto.account.request.AccountVerifyRequest;
import com.funeral.funeralService.service.AccountService;
import jakarta.servlet.http.HttpSession;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/users/account")
@RequiredArgsConstructor
public class AccountController {

    private final AccountService service;

    @PostMapping("/verify")
    public AccountSessionResponse verify(@Valid @RequestBody AccountVerifyRequest request, HttpSession session) {
        return service.verify(request, session);
    }

    @GetMapping("/session")
    public AccountSessionResponse getSession(HttpSession session) {
        return service.getSession(session);
    }

    @PostMapping("/logout")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void logout(HttpSession session) {
        service.logout(session);
    }
}
