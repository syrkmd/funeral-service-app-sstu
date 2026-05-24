package com.funeral.funeralService.service;

import com.funeral.funeralService.dto.account.response.AccountSessionResponse;
import com.funeral.funeralService.dto.account.request.AccountVerifyRequest;
import com.funeral.funeralService.exception.InvalidVerificationCodeException;
import jakarta.servlet.http.HttpSession;
import org.springframework.stereotype.Service;

@Service
public class AccountService {

    private static final String ACCOUNT_PHONE_SESSION_KEY = "accountPhone";
    private static final String DEMO_CODE = "4821";

    public AccountSessionResponse verify(AccountVerifyRequest request, HttpSession session) {
        if (!DEMO_CODE.equals(request.getCode())) {
            throw new InvalidVerificationCodeException();
        }

        session.setAttribute(ACCOUNT_PHONE_SESSION_KEY, request.getPhone());

        return new AccountSessionResponse(true, request.getPhone());
    }

    public AccountSessionResponse getSession(HttpSession session) {
        String phone = (String) session.getAttribute(ACCOUNT_PHONE_SESSION_KEY);

        if (phone == null) {
            return new AccountSessionResponse(false, null);
        }

        return new AccountSessionResponse(true, phone);
    }

    public void logout(HttpSession session) {
        session.removeAttribute(ACCOUNT_PHONE_SESSION_KEY);
    }
}
