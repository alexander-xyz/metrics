# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**

## Профилирование памяти (инкремент 17)

Профили сняты с эндпоинта `/debug/pprof/heap` под одинаковой нагрузкой
(8 воркеров × 1500 запросов `POST /updates` с заголовком `Accept-Encoding: gzip`).

Что показал `profiles/base.pprof`: 96% всех аллокаций приходилось на
`compress/flate.NewWriter` — gzip-писатель создавался заново на каждый запрос
в `GzipMiddleware`, даже если ответ в итоге не сжимался.

Что сделано:

- gzip-писатели в `internal/logger/compress.go` берутся из `sync.Pool`
  и создаются только когда ответ действительно сжимается;
- то же самое для функции `compress` агента в `cmd/agent/sender.go`;
- в `cmd/agent/collector.go` убрана промежуточная мапа на 28 метрик,
  которая создавалась при каждом опросе.

Результат сравнения профилей:

```
$ go tool pprof -top -sample_index=alloc_space -diff_base=profiles/base.pprof profiles/result.pprof
File: server
Type: alloc_space
Showing nodes accounting for -9228.41MB, 95.62% of 9650.84MB total
Dropped 155 nodes (cum <= 48.25MB)
      flat  flat%   sum%        cum   cum%
-7513.32MB 77.85% 77.85% -9101.28MB 94.31%  compress/flate.NewWriter (inline)
-1542.42MB 15.98% 93.83% -1542.42MB 15.98%  compress/flate.(*compressor).initDeflate (inline)
  -69.65MB  0.72% 94.56%   -69.65MB  0.72%  compress/flate.(*huffmanEncoder).generate
  -52.91MB  0.55% 95.10%   -52.91MB  0.55%  io.init.func1
  -48.61MB   0.5% 95.61%   -48.61MB   0.5%  sync.(*Pool).pinSlow
   -1.50MB 0.016% 95.62% -9108.30MB 94.38%  handler.GetRouter.UpdateMetricsJSONHandler.func8
      -1MB  0.01% 95.63% -9338.64MB 96.77%  net/http.(*conn).serve
    0.50MB 0.0052% 95.63% -9179.46MB 95.12%  logger.GzipMiddleware.func1
    0.50MB 0.0052% 95.62% -9181.46MB 95.14%  logger.RequestLogger.func1
```

Сравнение идёт по `alloc_space` — суммарному объёму выделенной памяти.
По умолчанию `pprof` показывает `inuse_space`, где выигрыш не виден:
пул намеренно удерживает переиспользуемые писатели в памяти.

Те же изменения на бенчмарках:

```
BenchmarkUpdateMetricsJSONHandlerGzip   830899 B/op  156 allocs/op  ->  22528 B/op  138 allocs/op
BenchmarkCompress                       814106 B/op   20 allocs/op  ->    250 B/op    3 allocs/op
BenchmarkCollectMetrics                    936 B/op    3 allocs/op  ->      0 B/op    0 allocs/op
```
