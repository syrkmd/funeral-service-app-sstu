# Funeral Service Management System

Веб-система для оформления и администрирования ритуальных услуг. Проект объединяет клиентское React-приложение, Spring Boot backend, PostgreSQL и отдельный Go proxy-service с контролем доступа, rate limiting и мониторингом.

Система позволяет оформить заказ из актуального каталога, выбрать участок кладбища через внешний cemetery-service, оплатить заказ через bank-service и управлять заказами и каталогом из административной панели.

## Возможности системы

### Клиентская часть

- просмотр каталога ритуальных услуг и товаров;
- многошаговое оформление заказа;
- ввод данных клиента и умершего;
- выбор даты, времени и адреса церемонии;
- получение секций и свободных участков из cemetery-service;
- выбор участка захоронения;
- серверная проверка выбранных позиций и расчёт стоимости;
- оплата банковской картой при оформлении заказа;
- возможность отложить оплату и оплатить заказ позднее;
- вход в личный кабинет по телефону и демонстрационному коду подтверждения;
- просмотр заказов, найденных по номеру телефона;
- просмотр состава, статуса, документов и места захоронения заказа.

### Административная часть

- вход по HTTP-сессии;
- просмотр списка заказов и подробной информации о заказе;
- изменение статуса заказа;
- изменение признака оплаты;
- изменение даты заказа;
- изменение данных клиента;
- изменение данных умершего;
- изменение деталей церемонии;
- полная замена услуг и товаров позициями из каталогов;
- применение скидки к неоплаченному заказу;
- добавление и удаление метаданных документов;
- удаление заказа;
- создание и редактирование товаров;
- создание и редактирование ритуальных услуг;
- архивирование и восстановление товаров и услуг;
- просмотр dashboard proxy-service;
- просмотр и управление IP access rules;
- просмотр текущих rate-limit правил, bucket usage и нарушений.

Изменяющие заказы и каталоги endpoints требуют активной административной HTTP-сессии.

### Proxy Service

Go proxy-service работает как reverse proxy перед backend и предоставляет отдельное административное API.

- allow list для разрешённых IP-адресов и диапазонов;
- deny list для блокировки IP-адресов и диапазонов;
- gray list с требованием дополнительной captcha-проверки;
- проверка IP по точному адресу, CIDR и диапазону;
- rate limiting по RPS, RPM, RPH и RPD;
- ограничения соединений и объёма трафика;
- отдельные лимиты для подсетей;
- HTTP response cache и endpoints его очистки;
- runtime-добавление и удаление IP-правил;
- сбор статистики запросов, блокировок, задержек и ошибок upstream;
- Prometheus metrics;
- dashboard endpoints для frontend;
- готовый Grafana dashboard;
- hot reload конфигурации из `config.yaml`.

Rate-limit настройки задаются в `proxy/config.yaml`. Текущий экран Rate Limiting в админ-панели используется прежде всего для live-мониторинга правил и нарушений.

## Архитектура

```text
                         +-----------------------+
                         |  React + TypeScript   |
                         |  Frontend :5173       |
                         +-----------+-----------+
                                     |
                 business REST API   |   proxy admin API / metrics
                                     |
                    +----------------+----------------+
                    |                                 |
                    v                                 v
        +------------------------+       +--------------------------+
        | Funeral Service       |       | Go Proxy Service :8090   |
        | Spring Boot :8080     |<------| reverse proxy, IP access, |
        +-----------+------------+       | rate limits, monitoring   |
                    |                    +-------------+------------+
                    |                                  |
                    v                                  v
        +------------------------+       +--------------------------+
        | PostgreSQL :5432      |       | Prometheus :9090         |
        | funeral_service       |       | Grafana :3000            |
        +------------------------+       +--------------------------+

        Funeral Service :8080 ------ HTTP ------> Bank Service :8082
        Funeral Service :8080 ------ HTTP ------> Cemetery Service :8081
```

В текущем `frontend/.env` бизнес-запросы отправляются прямо в backend через `VITE_API_URL=http://localhost:8080`. Запросы frontend к `/api`, `/metrics`, `/swagger` и `/healthz` в dev-режиме проксируются Vite на Go proxy-service. Сам proxy пересылает все остальные маршруты в настроенный `upstream_url`.

## Технологический стек

### Backend

- Java 17;
- Spring Boot 4;
- Spring Web MVC;
- Spring Data JPA;
- Jakarta Bean Validation;
- PostgreSQL;
- Flyway;
- Lombok;
- Spring `RestClient`;
- Springdoc OpenAPI / Swagger UI;
- Maven Wrapper.

### Frontend

- React 18;
- TypeScript;
- Vite 6;
- React Router;
- Zustand;
- Axios;
- Tailwind CSS 4;
- MUI и Radix UI;
- Lucide React;
- Recharts.

### Proxy

- Go 1.22;
- Gin;
- Prometheus client;
- Grafana;
- Zerolog;
- fsnotify;
- YAML configuration;
- Swaggo / Swagger;
- Docker и Docker Compose.

## Интеграции

### Bank Service

Backend использует `BankClientService` и отправляет запрос:

```text
POST http://localhost:8082/api/transfer/execute
```

Сценарий оплаты:

1. Клиент создаёт заказ или открывает неоплаченный заказ в личном кабинете.
2. Frontend отправляет данные карты в `POST /orders/{id}/pay`.
3. Backend берёт итоговую сумму из сохранённого заказа.
4. Backend формирует перевод на настроенную карту получателя.
5. После успешного ответа bank-service заказ сохраняется с `paid=true`.

Клиент не передаёт сумму платежа. Ошибки карты преобразуются в `400 Bad Request`, а недоступность bank-service — в `502 Bad Gateway`.

Bank-service является внешним приложением и должен быть запущен отдельно на порту `8082`. Его исходный код в текущий репозиторий не входит.

### Cemetery Service

Backend использует `CemeteryClientService`, авторизуется во внешнем сервисе и передаёт JWT в исходящих запросах.

Реализованный сценарий:

1. Frontend получает секции через `GET /cemetery/sections`.
2. После выбора секции frontend получает свободные участки через `GET /cemetery/plots?sectionName=...`.
3. Клиент выбирает участок по его коду.
4. При создании заказа backend вызывает `POST /cemetery/contracts/purchase`.
5. В заказе сохраняются `cemeteryPlotId` и `cemeteryPlotCode`.

Для резервирования обязательна дата церемонии. Cemetery-service является внешним приложением и должен быть запущен отдельно на порту `8081`. Его исходный код в текущий репозиторий не входит.

## Основные API

### Funeral Service API

| Группа | Базовый маршрут | Назначение |
|---|---|---|
| Account | `/users/account` | Подтверждение кода, получение и завершение клиентской сессии |
| Admin | `/admin/session` | Вход, проверка и завершение административной сессии |
| Orders | `/orders` | Создание, поиск, изменение и оплата заказов |
| Catalog | `/catalog` | Товары и категории товаров |
| Funeral Services | `/funeral-services`, `/service-categories` | Ритуальные услуги и их категории |
| Cemetery | `/cemetery` | Секции и свободные участки внешнего cemetery-service |

### Proxy API

| Группа | Маршрут | Назначение |
|---|---|---|
| IP access | `/api/ip_access` | Списки правил и проверка IP |
| Captcha | `/api/captcha/verify` | Подтверждение доступа для gray list |
| Cache | `/api/cache` | Инвалидация proxy cache |
| Metrics | `/api/metrics`, `/metrics` | JSON-метрики и Prometheus metrics |
| Dashboard | `/api/dashboard` | Overview, clients, upstream, rate limits и IP access |
| Health | `/healthz` | Проверка состояния proxy |

## База данных

Основной backend использует PostgreSQL 17:

- база данных: `funeral_service`;
- пользователь: `funeral_user`;
- порт: `5432`;
- Hibernate работает с `ddl-auto=validate`;
- создание и изменение схемы выполняет Flyway.

Миграции находятся в:

```text
backend/src/main/resources/db/migration/
```

Текущие миграции создают таблицы заказов, элементов заказа, документов, каталогов и категорий, добавляют поля церемонии и участка кладбища, а также заполняют стартовый каталог.

## Swagger

После запуска сервисов документация доступна по адресам:

- Funeral Service Swagger UI: [http://localhost:8080/swagger-ui/index.html](http://localhost:8080/swagger-ui/index.html)
- Funeral Service OpenAPI JSON: [http://localhost:8080/v3/api-docs](http://localhost:8080/v3/api-docs)
- Proxy Service Swagger UI: [http://localhost:8090/swagger/index.html](http://localhost:8090/swagger/index.html)
- Proxy Service health check: [http://localhost:8090/healthz](http://localhost:8090/healthz)

## Запуск проекта

### Требования

- Java 17 или новее;
- Docker и Docker Compose;
- Node.js;
- npm;
- Go 1.22, если proxy запускается без Docker;
- отдельно запущенные bank-service на `8082` и cemetery-service на `8081` для полных интеграционных сценариев.

### PostgreSQL

Из корня репозитория:

```bash
cd backend
docker compose up -d postgres
```

Контейнер `funeral-postgres` откроет PostgreSQL на `localhost:5432`.

### Backend

```bash
cd backend
./mvnw spring-boot:run
```

Backend запускается на [http://localhost:8080](http://localhost:8080). При старте Flyway автоматически применяет миграции.

### Proxy

Вариант с Docker, Prometheus и Grafana:

```bash
cd proxy
docker compose up --build -d
```

Будут доступны:

- proxy: [http://localhost:8090](http://localhost:8090)
- Prometheus: [http://localhost:9090](http://localhost:9090)
- Grafana: [http://localhost:3000](http://localhost:3000)

Для локального запуска только proxy:

```bash
cd proxy
go run ./cmd/app
```

Upstream backend задаётся параметром `proxy.upstream_url` в `proxy/config.yaml`.

### Frontend

```bash
cd frontend
npm install
npm run dev
```

Frontend запускается на [http://localhost:5173](http://localhost:5173).

Текущая конфигурация `frontend/.env`:

```dotenv
VITE_API_URL=http://localhost:8080
VITE_USE_MOCK=false
VITE_USE_MOCK_AUTH=false
```

Для отдельного адреса proxy frontend также поддерживает переменную `VITE_PROXY_API_URL`. Если она не задана, dev server Vite направляет proxy-маршруты на `http://localhost:8090`.

## Структура проекта

```text
.
├── frontend/   React-приложение, маршруты, Zustand stores и API clients
├── backend/    Spring Boot REST API, JPA entities, Flyway и интеграционные clients
├── proxy/      Go reverse proxy, IP access, rate limiting, cache и monitoring
├── db/         Зарезервированная директория; схема основной БД находится в Flyway
├── bank/       Локальная служебная директория без исходного кода bank-service
└── Cemetery/   Локальная служебная директория без исходного кода cemetery-service
```

Основные backend-пакеты:

```text
controller/   HTTP endpoints
service/      бизнес-логика и REST-интеграции
repository/   Spring Data JPA repositories
entity/       JPA-модель
dto/          request/response контракты по функциональным областям
mapper/       преобразование Order entity в API response
exception/    доменные исключения и GlobalExceptionHandler
config/       CORS и OpenAPI
```

## Особенности реализации

- Статус нового заказа устанавливается backend как `PROCESSING`.
- Новый заказ всегда создаётся неоплаченным.
- Стоимость заказа вычисляется на сервере по активным позициям каталогов.
- Переданные frontend цены не используются как доверенный источник.
- Услуги и товары сохраняются в заказе как снимок названия и цены.
- Изменение состава заказа выполняется по ID активных позиций каталога.
- Товары и услуги архивируются через `active=false`, без физического удаления.
- Архивные позиции можно восстановить через `PATCH` с `active=true`.
- DTO проверяются Jakarta Bean Validation.
- Ошибки API преобразуются в единый формат через `GlobalExceptionHandler`.
- Клиентская и административная авторизация используют HTTP-сессии и cookie.
- Административные mutation endpoints явно проверяют административную сессию.
- Внешние bank-service и cemetery-service вызываются через Spring `RestClient`.
- Изменение JPA-коллекций заказа выполняется внутри транзакций.
- Proxy хранит runtime-состояние IP access, rate-limit buckets, cache и метрики в памяти.

Демонстрационные данные текущей реализации:

- admin login: `admin`;
- admin password: `admin`;
- код подтверждения клиентского аккаунта: `4821`.

## Потенциальные направления развития

- сохранение идентификатора договора cemetery-service в заказе;
- компенсационный сценарий снятия брони участка при ошибке создания заказа;
- хранение идентификатора банковской транзакции и история платежей;
- идемпотентность внешних платежей и резервирований;
- перенос демонстрационных учётных данных в защищённую конфигурацию;
- постоянное хранение runtime-настроек proxy;
- полноценное изменение rate-limit конфигурации из административной панели;
- расширение автоматических интеграционных тестов;
- дальнейшее развитие Grafana dashboard и alerting.
