# Нагрузочное тестирование

Образ k6 — `grafana/k6:latest`.

Задача стояла не в том, чтобы выжать максимум RPS, а в двух практических вопросах. Выдержит ли API смешанную нагрузку ~150 запросов/с с приемлемой задержкой? И как поведёт себя при превышении лимита — упадёт с 5xx или корректно ответит 429?

## Итог

| Сценарий | Как запускали | Объём | p95 | Ошибки | Вердикт |
|---|---|---:|---:|---|---|
| **S6 mixed** | `make loadtest` | 17 405 запросов, **144.9/s** | **11.6 ms** | 0% failed, 0× 5xx | **PASS** |
| **S7 spike** | `make loadtest-spike-default` | 9 549 запросов, **191/s** | 4.9 ms на 200 / 0.8 ms на 429 | 0× 5xx, **56.7% 429** | **PASS** (429 здесь ожидаемы) |

Смешанный сценарий проходит порог p95 < 250 ms с большим запасом. Всплеск не валит API: лимитер отсекает лишнее, успешные ответы остаются быстрыми.

## Артефакты

Сырые логи и JSON оставлены в репозитории [loadtest](/docs/loadtest/), чтобы можно было глянуть.

| Файл | Что внутри |
|---|---|
| [loadtest/env-snapshot.txt](loadtest/env-snapshot.txt) | `docker compose ps` и `/health/ready` на момент прогона |
| [loadtest/mixed-console.log](loadtest/mixed-console.log) | Полный stdout k6, mixed |
| [loadtest/mixed-summary.json](loadtest/mixed-summary.json) | Метрики и checks в JSON, mixed |
| [loadtest/mixed-k6-summary.txt](loadtest/mixed-k6-summary.txt) | Выдержка THRESHOLDS и TOTAL RESULTS, mixed |
| [loadtest/spike-console.log](loadtest/spike-console.log) | Полный stdout k6, spike |
| [loadtest/spike-summary.json](loadtest/spike-summary.json) | Метрики и checks в JSON, spike |
| [loadtest/spike-k6-summary.txt](loadtest/spike-k6-summary.txt) | Выдержка THRESHOLDS и TOTAL RESULTS, spike |

Скрипты: `test/k6/mixed.js`, `test/k6/spike.js`.

## Окружение

Стек тот же: один инстанс API, Postgres 16, симулятор заведения в `auto_accept`. Пул приложения — 20 соединений, у Postgres `max_connections=100`. В сиде — демо-заведения и тысяча покупателей с UUID вида `aaaaaaaa-aaaa-4aaa-8aaa-*`, чтобы запись в корзину не билась об одного клиента.

Сценарии нарочно стоят на разной конфигурации лимитов.

- **S6.** Поверх Compose наложен [`docker-compose.loadtest.yml`](../docker-compose.loadtest.yml): `RATE_LIMIT_RPS_PUBLIC=500`, запись 150, партнёр 200. Иначе на цели 150 RPS мы измеряли бы 429, а не приложение.
- **S7.** Оверлея нет, публичный лимит остаётся 100 RPS. Цель спайка — 300. Здесь как раз интересно, как ведёт себя ограничитель.

Снимок контейнеров: [env-snapshot.txt](loadtest/env-snapshot.txt).

## S6 — смешанная нагрузка

Две минуты, цель около 150 RPS. Скрипт чередует витрину, меню, корзину, оформление заказа и очередь партнёра — ближе к живому трафику, чем только GET /venues.

Фрагмент из [mixed-k6-summary.txt](loadtest/mixed-k6-summary.txt):

```
  THRESHOLDS
    http_req_duration   ✓ 'p(95)<250'  p(95)=11.61ms
    http_req_failed     ✓ 'rate<0.01'  rate=0.00%

  TOTAL RESULTS
    checks_total.......: 17405   144.86/s
    checks_succeeded...: 100.00% 17405/17405

    http_req_duration..: avg=4.97ms  p(95)=11.61ms  max=474ms
    http_req_failed....: 0.00%  0/17405
    http_reqs..........: 17405  144.86/s
    iterations.........: 15447  128.57/s
    vus_max............: 87
```

Все проверки эндпоинтов из [mixed-summary.json](loadtest/mixed-summary.json) зелёные:

| Эндпоинт / check | Успех | Провал |
|---|---:|---:|
| `GET /venues` | 5 548 | 0 |
| `GET /venues/{slug}/menu` | 5 543 | 0 |
| `POST /cart/items` | 1 958 | 0 |
| `GET /cart` | 1 296 | 0 |
| `POST /orders` | 662 | 0 |
| `GET /partner/v1/orders` | 2 398 | 0 |
| **Итого** | **17 405** | **0** |

Максимум 474 ms на одном запросе есть, но это хвост: медиана 2.86 ms, p95 — 11.6 ms. До порога 250 ms остаётся почти порядок.

## S7 — спайк

Пятьдесят секунд, линейный разгон 20 → 300 RPS на дефолтном лимите 100. Смысл сценария это убедиться, что лишние запросы получают 429, а не 5xx, и что «пропущенные» 200 не начинают тормозить.

Фрагмент из [spike-k6-summary.txt](loadtest/spike-k6-summary.txt):

```
  THRESHOLDS
    http_req_duration{status:200}  ✓ p(95)=4.9ms
    http_req_duration{status:429}  ✓ p(95)=804µs
    http_req_failed                ✗ rate=56.71%   ← k6 считает 429 «failed»

  TOTAL RESULTS
    checks_total.......: 19098   382/s
    checks_succeeded...: 100.00%

    http_reqs..........: 9549   191/s
      200 OK  → 4133 (43.3%)
      429     → 5416 (56.7%)
      5xx     →    0 (0.0%)
```

k6 завершился с кодом **99**: порог `http_req_failed rate<0.01` формально нарушен, потому что для k6 статус 429 — ошибка запроса. Наши checks `no 5xx` и `200 or 429` при этом 100%. Это ожидаемое поведение ограничителя, когда цель 300 RPS, а лимит 100.

Ответы 429 дешёвые (p95 меньше миллисекунды): лимитер не ходит в бизнес-логику. Успешные 200 не раздуваются — p95 4.9 ms, даже быстрее mixed, потому что в спайке нет тяжёлого checkout.

## Как воспроизвести

```bash
# S6 — mixed, артефакты пишутся в docs/loadtest/
make loadtest

# S7 — spike на дефолтных лимитах (api без loadtest-оверлея)
docker compose up -d api --force-recreate --wait
make loadtest-spike-default
```

Если меняли `migrations/seed/seed.sql`, перед прогоном обновите данные:

```bash
docker compose up -d --build migrator
docker compose run --rm migrator seed
```

## Выводы

На текущем сиде смешанный поток около 145 RPS платформа держит нормально: p95 11.6 ms против порога 250 ms. Упираться в меню из двенадцати позиций мы не начали; когда каталог вырастет, первый рычаг уже есть — ETag и 304 на меню, дальше имеет смысл кэш готового JSON.

Спайк показывает другую вещь: при превышении лимита API не падает, а отдает 429 без 5xx. Это важнее, чем просто максимальный RPS. Если понадобится увеличить лимит — это делается через конфиг, а не срочным патчем хендлера.
