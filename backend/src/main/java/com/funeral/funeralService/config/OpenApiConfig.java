package com.funeral.funeralService.config;

import io.swagger.v3.oas.models.OpenAPI;
import io.swagger.v3.oas.models.info.Info;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class OpenApiConfig {

    @Bean
    public OpenAPI funeralServiceOpenApi() {
        return new OpenAPI()
                .info(new Info()
                        .title("API системы ритуальных услуг")
                        .version("0.0.1")
                        .description("""
                                REST API для системы управления ритуальными услугами.
                                API включает оформление заказов, админское изменение заказов,
                                управление каталогом услуг и товаров, авторизацию через сессии,
                                подтверждение аккаунта и интеграцию с выбором места захоронения.
                                """));
    }
}
