# Grafana + Prometheus в этом проекте (обучение с нуля)

## 0) Что мы добавили в проект

- `prometheus` в `docker-compose.yml` — сервис, который регулярно забирает метрики с API.
- `grafana` в `docker-compose.yml` — интерфейс, где строятся графики и дашборды.
- `configs/prometheus.yml` — файл, где написано **что** Prometheus должен опрашивать.
- `/metrics` в API (`internal/api/server.go`) — endpoint, который отдаёт метрики.

---

## 1) Словарь простыми словами

- **Метрика (metric)** — число, которое меняется во времени. Пример: сколько запросов пришло.
- **Prometheus** — "сборщик" метрик. Сам ходит по URL и собирает числа.
- **Grafana** — "экран" для графиков. Берёт данные из Prometheus и рисует их.
- **Scrape** — опрос endpoint-а Prometheus-ом ("сходить и забрать метрики").
- **Target** — адрес, который Prometheus опрашивает.
- **Endpoint `/metrics`** — URL, где приложение отдаёт метрики в текстовом формате.
- **Counter** — счётчик, который только растёт (например, число запросов).
- **Histogram** — метрика времени/размеров, чтобы видеть распределение (например, latency).
- **Dashboard** — набор графиков в Grafana.
- **Panel** — один график внутри dashboard.

---

## 2) Почему target = `host.docker.internal:8080`

Prometheus работает в Docker-контейнере, а ваш Go API обычно запускается локально командой `go run` на хост-машине.

- `localhost` внутри контейнера = сам контейнер, а не ваш компьютер.
- `host.docker.internal` — специальное имя хоста (вашего ПК) из контейнера.

Поэтому Prometheus ходит на `host.docker.internal:8080/metrics`.

---

## 3) Пошаговый запуск (каждая команда объяснена)

### Шаг 1. Поднять инфраструктуру

```bash
docker compose up -d
```

Разбор команды:
- `docker compose` — управление сервисами из `docker-compose.yml`.
- `up` — создать и запустить сервисы.
- `-d` (`detached`) — запустить в фоне, чтобы терминал не блокировался.

Что поднимется:
- Kafka, ZooKeeper, Redis, Prometheus, Grafana.

### Шаг 2. Запустить API

```bash
go run cmd/order-api/main.go
```

Разбор:
- `go run` — компилирует и сразу запускает Go-программу.
- `cmd/order-api/main.go` — точка входа HTTP API.

После запуска API должны работать endpoints:
- `http://localhost:8080/health`
- `http://localhost:8080/metrics`

### Шаг 3. Проверить, что метрики реально отдаются

```bash
curl http://localhost:8080/metrics
```

Разбор:
- `curl` — утилита для HTTP-запросов из терминала.
- Ответ — длинный текст со строками метрик (`http_requests_total`, `http_request_duration_seconds`, ...).

### Шаг 4. Проверить Prometheus

Открыть в браузере:
- `http://localhost:9090`

Потом:
- `Status` → `Targets`
- у `order-api` должно быть состояние `UP`.

Если `DOWN`:
- проверьте, что API запущен;
- проверьте `http://localhost:8080/metrics`;
- проверьте, что Docker Desktop запущен.

### Шаг 5. Подключить Prometheus как источник в Grafana

Открыть в браузере:
- `http://localhost:4000`

Логин/пароль по умолчанию (из compose):
- login: `admin`
- password: `admin`

Дальше в Grafana:
1. `Connections` → `Data sources`
2. `Add data source` → `Prometheus`
3. URL: `http://prometheus:9090`
4. `Save & test`

Почему URL такой:
- Grafana тоже в Docker, и обращается к Prometheus по имени сервиса `prometheus` внутри Docker-сети.

### Шаг 6. Сделать первый график

1. `Dashboards` → `New` → `New dashboard` → `Add visualization`
2. Выбрать source: `Prometheus`
3. Вставить запрос:

```promql
sum(rate(http_requests_total[1m]))
```

Что означает выражение:
- `http_requests_total` — общий счётчик запросов.
- `rate(...[1m])` — средняя скорость роста счётчика за последнюю минуту.
- `sum(...)` — суммирование по всем меткам.

Итог: график "сколько запросов в секунду сейчас".

---

## 4) Полезные PromQL-запросы для тренировки

- RPS (запросы/сек):

```promql
sum(rate(http_requests_total[1m]))
```

- RPS по route:

```promql
sum by (route) (rate(http_requests_total[1m]))
```

- Ошибки 5xx в сек:

```promql
sum(rate(http_requests_total{status=~"5.."}[1m]))
```

- 95-й перцентиль latency:

```promql
histogram_quantile(0.95, sum by (le, route) (rate(http_request_duration_seconds_bucket[5m])))
```

---

## 5) План обучения на 7 дней (понятный и короткий)

### День 1: База
- Поднять стек, открыть все UI, убедиться что target `UP`.
- Понять роли: кто собирает (Prometheus), кто рисует (Grafana).

### День 2: Метрики API
- Вызвать `/health` и `/order/*`, посмотреть как растёт `http_requests_total`.
- Понять разницу между `counter` и `histogram`.

### День 3: PromQL основы
- Освоить `sum`, `rate`, фильтры `{status="200"}`.
- Построить 2 графика: общий RPS и RPS по route.

### День 4: Latency
- Освоить `histogram_quantile`.
- Построить p50 и p95 latency по route.

### День 5: Ошибки
- Сделать панель с 4xx и 5xx.
- Понять, как находить деградацию по росту ошибок.

### День 6: Дашборд
- Собрать один dashboard: RPS, ошибки, p95, health.
- Подписать панели человеко-понятными названиями.

### День 7: Повторение
- С нуля поднять всё по памяти.
- Объяснить самому себе путь данных: API → Prometheus → Grafana.

---

## 6) Остановка и очистка

Остановить сервисы:

```bash
docker compose down
```

Остановить и удалить тома (включая данные Grafana/Redis):

```bash
docker compose down -v
```

Разбор:
- `down` — остановить и удалить контейнеры проекта.
- `-v` — удалить volumes (данные тоже удалятся).
