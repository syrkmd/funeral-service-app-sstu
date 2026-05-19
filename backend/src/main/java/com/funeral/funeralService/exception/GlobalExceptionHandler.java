package com.funeral.funeralService.exception;

import com.funeral.funeralService.exception.dto.ApiError;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.MethodArgumentNotValidException;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.ResponseStatus;
import org.springframework.web.bind.annotation.RestControllerAdvice;

@RestControllerAdvice
public class GlobalExceptionHandler {

    @ResponseStatus(HttpStatus.NOT_FOUND)
    @ExceptionHandler(OrderNotFoundException.class)
    public ApiError handleOrderNotFound(OrderNotFoundException exception) {
        return new ApiError("ORDER_NOT_FOUND", exception.getMessage());
    }

    @ResponseStatus(HttpStatus.NOT_FOUND)
    @ExceptionHandler(OrderDocumentNotFoundException.class)
    public ApiError handleOrderDocumentNotFound(OrderDocumentNotFoundException exception) {
        return new ApiError("ORDER_DOCUMENT_NOT_FOUND", exception.getMessage());
    }

    @ResponseStatus(HttpStatus.BAD_REQUEST)
    @ExceptionHandler(InvalidVerificationCodeException.class)
    public ApiError handleInvalidVerificationCode(InvalidVerificationCodeException exception) {
        return new ApiError("INVALID_VERIFICATION_CODE", exception.getMessage());
    }

    @ResponseStatus(HttpStatus.BAD_REQUEST)
    @ExceptionHandler(MethodArgumentNotValidException.class)
    public ApiError handleValidation(MethodArgumentNotValidException exception) {
        String message = exception.getBindingResult()
                .getFieldErrors()
                .stream()
                .findFirst()
                .map(err -> err.getField() + ": " + err.getDefaultMessage())
                .orElse("Validation error");

        return new ApiError("VALIDATION_ERROR", message);
    }

}
