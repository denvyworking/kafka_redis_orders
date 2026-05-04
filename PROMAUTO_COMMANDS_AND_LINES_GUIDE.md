# Пошаговый гайд: `prometheus` + `promauto` в Go (каждая команда и каждая строка)

Этот файл — отдельная шпаргалка именно под твой проект.

---

## 1) Команды установки: разбор каждого слова

### Команда A

```bash
go get github.com/prometheus/client_golang/prometheus
```

Разбор по словам:
- `go` — CLI языка Go.
- `get` — команда для добавления/обновления зависимости в модуле.
- `github.com/prometheus/client_golang/prometheus` — путь к пакету `prometheus` внутри модуля `client_golang`.

Что произойдёт:
- Go добавит (или обновит) модуль в `go.mod`.
- Хеши попадут в `go.sum`.

---

### Команда B

```bash
go get github.com/prometheus/client_golang/prometheus/promauto
```

Разбор по словам:
- `go` — CLI Go.
- `get` — скачать/обновить зависимость в текущем модуле.
- `github.com/prometheus/client_golang/prometheus/promauto` — пакет `promauto` (авто-регистрация метрик).

Важно:
- Это тот же модуль `client_golang`, просто другой пакет.
- Часто достаточно одной команды:

```bash
go get github.com/prometheus/client_golang@latest
```

---

## 2) Где это используется в твоём коде

Файл: `internal/api/server.go`.

Ты используешь схему:
1. создать `registry`;
2. создать метрики через `promauto.With(registry)`;
3. отдать `/metrics` через `promhttp.HandlerFor(registry, ...)`.

Это правильный production-friendly подход: не глобальный реестр, а локальный/явный.

---

## 3) Разбор кода: каждая строка и зачем она нужна

Ниже важный фрагмент (упрощённо):

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)
```

Построчно:
- `prometheus` — типы метрик и опции (`CounterOpts`, `HistogramOpts`, buckets).
- `promauto` — удобное создание метрик с авто-регистрацией.
- `promhttp` — HTTP handler для endpoint `/metrics`.

---

```go
registry := prometheus.NewRegistry()
```

Построчно:
- `registry` — контейнер метрик.
- `NewRegistry()` — создаёт отдельный реестр только для этого сервера.
- Зачем: контроль, тестируемость, отсутствие конфликтов с глобальным реестром.

---

```go
requestsTotal := promauto.With(registry).NewCounterVec(
    prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total amount of HTTP requests grouped by route, method and status code.",
    },
    []string{"route", "method", "status"},
)
```

Построчно:
- `promauto.With(registry)` — говорим: регистрировать метрику именно в `registry`.
- `NewCounterVec` — создаём счётчик с labels.
- `CounterOpts` — конфиг метрики.
- `Name` — имя, которое увидишь в Prometheus/Grafana.
- `Help` — описание метрики.
- `[]string{"route", "method", "status"}` — список labels.

Смысл метрики:
- сколько HTTP-запросов всего, с разбивкой по пути, методу и статусу.

---

```go
requestDuration := promauto.With(registry).NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "http_request_duration_seconds",
        Help:    "HTTP request latency in seconds grouped by route and method.",
        Buckets: prometheus.DefBuckets,
    },
    []string{"route", "method"},
)
```

Построчно:
- `NewHistogramVec` — гистограмма с labels.
- `Name` — имя метрики latency.
- `Help` — описание.
- `Buckets: prometheus.DefBuckets` — стандартные интервалы времени.
- labels `route`, `method` — разбивка задержек по endpoint и HTTP методу.

Смысл метрики:
- измерение распределения времени ответа (а не только среднего).

---

```go
s.mux.Handle("/metrics", promhttp.HandlerFor(s.registry, promhttp.HandlerOpts{}))
```

Построчно:
- `Handle("/metrics", ...)` — регистрируем endpoint метрик.
- `HandlerFor(s.registry, ...)` — отдаём именно метрики из твоего `registry`.
- `HandlerOpts{}` — дополнительные опции handler (сейчас дефолт).

Результат:
- Prometheus может читать метрики по URL `/metrics`.

---

```go
status := strconv.Itoa(recorder.status)
s.requestsTotal.WithLabelValues(route, r.Method, status).Inc()
s.requestDuration.WithLabelValues(route, r.Method).Observe(time.Since(startedAt).Seconds())
```

Построчно:
- `strconv.Itoa(...)` — переводим HTTP статус из числа в строку для label.
- `WithLabelValues(...).Inc()` — увеличиваем счётчик на 1 для конкретной комбинации labels.
- `Observe(...)` — отправляем время выполнения в histogram.

Итог:
- на каждый запрос пишутся 2 факта: "один запрос пришёл" и "запрос занял N секунд".

---

## 4) Мини-практика: чтобы закрепить

### Шаг 1. Запусти API

```bash
go run cmd/order-api/main.go
```

Что значит:
- `go run` — собрать и запустить.
- `cmd/order-api/main.go` — точка входа сервиса.

---

### Шаг 2. Сгенерируй трафик

```bash
curl http://localhost:8080/health
curl http://localhost:8080/order/non-existent-id
```

Что значит:
- `curl` — HTTP-запрос из терминала.
- первый вызов обычно даст `200`, второй — `404`.

---

### Шаг 3. Проверь метрики сырым текстом

```bash
curl http://localhost:8080/metrics
```

Что искать:
- `http_requests_total`
- `http_request_duration_seconds_bucket`

---

### Шаг 4. Проверь в Prometheus UI

Открыть: `http://localhost:9090`

Запросы:

```promql
sum(rate(http_requests_total[1m]))
sum by (status) (rate(http_requests_total[1m]))
```

Смысл:
- общий RPS;
- RPS по HTTP-статусам.

---

## 5) Что выучить «в совершенстве» (дорожная карта)

1. Типы метрик и когда что применять.
2. Labels и кардинальность (не использовать user_id/order_id/email в labels).
3. PromQL: `rate`, `increase`, `sum by`, фильтры по labels.
4. Работа с latency через `histogram_quantile`.
5. Связка "кодовая метрика -> Prometheus -> Grafana panel -> алерт".

---

## 6) Частые вопросы

### Нужно ли каждый раз делать `go get`?
Нет. Обычно один раз добавил зависимость, дальше просто пишешь код и коммитишь `go.mod`/`go.sum`.

### Когда делать `go get` снова?
- когда добавляешь новую библиотеку;
- когда хочешь обновить версию библиотеки.

### `promauto` обязателен?
Нет. Это удобный слой. Можно всё делать через `prometheus.New*` + ручную регистрацию.

### Почему в проекте лучше `promauto.With(registry)`, а не просто `promauto.New*`?
Потому что так ты явно контролируешь реестр и избегаешь проблем с глобальным реестром в тестах и сложных сервисах.
