# Антиплагиат — Микросервисная система

Система для приема контрольных работ, хранения файлов и автоматического обнаружения плагиата с использованием алгоритмов сравнения текстов.

## 📋 Описание проекта

Проект представляет собой микросервисную архитектуру, состоящую из трех основных сервисов:

- **API Gateway** — единая точка входа для клиентов, маршрутизация запросов
- **File Storing Service** — сервис для загрузки и хранения файлов работ студентов
- **File Analysis Service** — сервис для анализа работ на предмет плагиата

### Основные возможности

- ✅ Загрузка файлов (TXT, DOCX) с валидацией размера и типа
- ✅ Автоматическое вычисление SHA256 хеша для обнаружения идентичных файлов
- ✅ Анализ текстовой схожести с использованием алгоритма Jaccard similarity на шинглах
- ✅ Асинхронная обработка задач анализа
- ✅ RESTful API с JSON форматом
- ✅ Graceful shutdown и healthcheck endpoints
- ✅ Типизированная обработка ошибок
- ✅ Unit-тесты для критичной логики

## 🏗️ Архитектура

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       ▼
┌─────────────────┐
│  API Gateway     │  :8080
│  (маршрутизация) │
└──────┬──────┬────┘
       │      │
       ▼      ▼
┌──────────┐ ┌──────────────┐
│ File     │ │ File         │
│ Store    │ │ Analysis     │
│ :8081    │ │ :8082        │
└──────────┘ └──────────────┘
```

### Компоненты

- **API Gateway** (`/gateway`): REST API, валидация, маршрутизация запросов к store/analysis сервисам
- **File Storing Service** (`/file-store`): прием файлов (TXT/DOCX), сохранение, вычисление хеша, метаданные работ
- **File Analysis Service** (`/file-analysis`): постановка задач анализа в очередь, worker для обработки, расчет схожести/плагиата

### Поток работы

1. **Загрузка работы**: `POST /works` (Gateway → Store)
   - Клиент отправляет multipart форму с файлом, `student_id`, `assignment_id`
   - Gateway валидирует и проксирует запрос в Store
   - Store сохраняет файл, вычисляет SHA256 хеш, создает метаданные
   - Gateway автоматически запускает анализ работы

2. **Анализ плагиата**: `POST /analysis/jobs` (Gateway → Analysis)
   - Gateway отправляет задачу анализа в Analysis сервис
   - Worker обрабатывает задачу асинхронно:
     - Проверяет точное совпадение хеша с другими работами
     - Извлекает текст и вычисляет схожесть через Jaccard similarity
     - Сохраняет отчет с результатами

3. **Получение результатов**: `GET /works/{work_id}/reports` (Gateway → Analysis)
   - Клиент запрашивает отчеты по работе
   - Возвращается список отчетов со статусами, флагом плагиата и оценкой схожести

### Модель данных

**Work** (работа студента):
- `id` — уникальный идентификатор
- `student_id` — идентификатор студента
- `assignment_id` — идентификатор задания
- `filename` — имя файла
- `storage_path` — путь к файлу в хранилище
- `mime` — MIME тип файла
- `size` — размер файла в байтах
- `hash` — SHA256 хеш файла
- `created_at` — время создания

**Report** (отчет анализа):
- `id` — уникальный идентификатор
- `work_id` — идентификатор работы
- `status` — статус анализа (`pending`, `running`, `done`, `failed`)
- `plagiarism_flag` — флаг обнаружения плагиата
- `score` — оценка схожести (0.0 до 1.0)
- `details_path` — путь к детальному отчету (JSON)
- `created_at`, `updated_at` — временные метки

### Алгоритм обнаружения плагиата

1. **Проверка хеша (SHA256)**:
   - Если найден идентичный хеш для другой работы того же задания → `plagiarism_flag=true`, `score=1.0`

2. **Текстовая схожесть** (для неидентичных файлов):
   - Извлечение текста из файла (TXT поддерживается, DOCX — в разработке)
   - Построение 5-символьных шинглов (n-грамм)
   - Вычисление Jaccard similarity между текстами
   - Если `similarity >= 0.7` → `plagiarism_flag=true`, `score=similarity`

## 🚀 Быстрый старт

### Требования

- Go 1.22 или выше
- Docker и docker-compose
- Make (опционально, для удобства)

### Установка и запуск

#### Вариант 1: Запуск через Docker Compose (рекомендуется)

```bash
# Клонировать репозиторий
git clone <repository-url>
cd File-storing-service

# Собрать и запустить все сервисы
make up

# Или вручную:
cd deploy && docker-compose up -d
```

Сервисы будут доступны:
- **Gateway**: http://localhost:8080
- **File Store**: http://localhost:8081
- **File Analysis**: http://localhost:8082

#### Вариант 2: Локальный запуск (для разработки)

```bash
# Собрать все сервисы
make build

# Запустить сервисы в отдельных терминалах:

# Терминал 1: Gateway
cd gateway && HTTP_ADDR=:8080 go run ./cmd/server

# Терминал 2: File Store
cd file-store && HTTP_ADDR=:8081 go run ./cmd/server

# Терминал 3: File Analysis
cd file-analysis && HTTP_ADDR=:8082 PLAGIARISM_THRESHOLD=0.7 go run ./cmd/server
```

### Переменные окружения

#### Gateway
- `HTTP_ADDR` — адрес для прослушивания (по умолчанию `:8080`)
- `HTTP_MAX_BODY_BYTES` — максимальный размер тела запроса (по умолчанию `20971520` = 20MB)
- `STORE_SERVICE_URL` — URL сервиса file-store (по умолчанию `http://file-store:8081`)
- `ANALYSIS_SERVICE_URL` — URL сервиса file-analysis (по умолчанию `http://file-analysis:8082`)

#### File Store
- `HTTP_ADDR` — адрес для прослушивания (по умолчанию `:8080`)
- `HTTP_MAX_BODY_BYTES` — максимальный размер тела запроса (по умолчанию `20971520`)

#### File Analysis
- `HTTP_ADDR` — адрес для прослушивания (по умолчанию `:8080`)
- `HTTP_MAX_BODY_BYTES` — максимальный размер тела запроса (по умолчанию `20971520`)
- `PLAGIARISM_THRESHOLD` — порог схожести для обнаружения плагиата (по умолчанию `0.7`)

## 🧪 Тестирование

> 📖 **Подробное руководство по тестированию**: см. [TESTING.md](TESTING.md) для детальных примеров и сценариев

### Автоматическое тестирование

Запуск всех unit-тестов:

```bash
make test
```

Или для конкретного сервиса:

```bash
# Тесты file-store
cd file-store && go test ./...

# Тесты file-analysis
cd file-analysis && go test ./...

# Тесты gateway (пока нет тестов)
cd gateway && go test ./...
```

Покрытие тестами:
- ✅ `file-store/internal/service` — тесты сервисного слоя
- ✅ `file-store/internal/validator` — тесты валидации
- ✅ `file-analysis/internal/analyzer` — тесты алгоритма плагиата

### Ручное тестирование

#### 1. Проверка healthcheck

```bash
# Gateway
curl http://localhost:8080/healthz

# File Store
curl http://localhost:8081/healthz

# File Analysis
curl http://localhost:8082/healthz
```

Ожидаемый ответ:
```json
{"status":"ok"}
```

#### 2. Загрузка работы через Gateway

```bash
# Создайте тестовый файл
echo "This is my assignment work. It contains original content." > test.txt

# Загрузите работу
curl -X POST http://localhost:8080/works \
  -F "file=@test.txt" \
  -F "student_id=student1" \
  -F "assignment_id=assignment1"
```

Ожидаемый ответ:
```json
{
  "id": 1,
  "student_id": "student1",
  "assignment_id": "assignment1",
  "filename": "test.txt",
  "hash": "a1b2c3d4e5f6...",
  "size": 58,
  "created_at": "2024-01-15T10:30:00Z",
  "storage_path": "assignment1/test.txt"
}
```

#### 3. Получение метаданных работы

```bash
# Замените {id} на ID из предыдущего ответа
curl http://localhost:8080/works/1
```

#### 4. Загрузка второй работы (для проверки плагиата)

```bash
# Создайте похожий файл
echo "This is my assignment work. It contains original content." > test2.txt

# Загрузите от другого студента
curl -X POST http://localhost:8080/works \
  -F "file=@test2.txt" \
  -F "student_id=student2" \
  -F "assignment_id=assignment1"
```

#### 5. Проверка отчетов анализа

```bash
# Подождите несколько секунд для обработки анализа, затем:
curl http://localhost:8080/works/1/reports
```

Ожидаемый ответ:
```json
[
  {
    "id": 1,
    "status": "done",
    "plagiarism_flag": true,
    "score": 1.0,
    "created_at": "2024-01-15T10:30:05Z",
    "updated_at": "2024-01-15T10:30:10Z"
  }
]
```

#### 6. Прямой доступ к File Store

```bash
# Загрузка напрямую в File Store
curl -X POST http://localhost:8081/works \
  -F "file=@test.txt" \
  -F "student_id=student3" \
  -F "assignment_id=assignment1"

# Получение метаданных
curl http://localhost:8081/works/1

# Скачивание файла
curl http://localhost:8081/works/1/file -o downloaded.txt
```

#### 7. Прямой доступ к File Analysis

```bash
# Создание задачи анализа
curl -X POST http://localhost:8082/analysis/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "work_id": 1,
    "assignment_id": "assignment1",
    "student_id": "student1",
    "file_hash": "a1b2c3d4e5f6...",
    "storage_path": "assignment1/test.txt"
  }'

# Получение отчетов
curl http://localhost:8082/works/1/reports
```

### Примеры тестовых сценариев

#### Сценарий 1: Идентичные файлы (100% плагиат)

```bash
# Создайте файл
echo "Identical content" > work1.txt

# Загрузите от студента 1
curl -X POST http://localhost:8080/works \
  -F "file=@work1.txt" \
  -F "student_id=student1" \
  -F "assignment_id=assignment1"

# Скопируйте файл и загрузите от студента 2
cp work1.txt work2.txt
curl -X POST http://localhost:8080/works \
  -F "file=@work2.txt" \
  -F "student_id=student2" \
  -F "assignment_id=assignment1"

# Проверьте отчеты - должен быть plagiarism_flag=true, score=1.0
sleep 3
curl http://localhost:8080/works/2/reports
```

#### Сценарий 2: Высокая схожесть (>70%)

```bash
# Создайте файл с длинным текстом
cat > work1.txt << EOF
This is a long document about computer science.
It discusses algorithms and data structures.
The content is original and well-researched.
EOF

# Создайте похожий файл
cat > work2.txt << EOF
This is a long document about computer science.
It discusses algorithms and data structures.
The content is original and well-researched with minor changes.
EOF

# Загрузите оба файла и проверьте схожесть
```

#### Сценарий 3: Низкая схожесть (<70%)

```bash
# Создайте два разных файла
echo "Completely different content about mathematics" > work1.txt
echo "Totally unrelated text about cooking recipes" > work2.txt

# Загрузите и проверьте - plagiarism_flag должен быть false
```

## 📚 API Документация

### Gateway API (порт 8080)

#### `POST /works`
Загрузка работы студента.

**Request:**
- Content-Type: `multipart/form-data`
- Поля:
  - `file` (обязательно) — файл работы
  - `student_id` (обязательно) — идентификатор студента
  - `assignment_id` (обязательно) — идентификатор задания

**Response:** `201 Created`
```json
{
  "id": 1,
  "student_id": "student1",
  "assignment_id": "assignment1",
  "filename": "work.txt",
  "hash": "abc123...",
  "size": 1024,
  "created_at": "2024-01-15T10:30:00Z",
  "storage_path": "assignment1/work.txt"
}
```

#### `GET /works/{id}`
Получение метаданных работы.

**Response:** `200 OK`
```json
{
  "id": 1,
  "student_id": "student1",
  "assignment_id": "assignment1",
  "filename": "work.txt",
  "hash": "abc123...",
  "size": 1024,
  "mime": "text/plain",
  "created_at": "2024-01-15T10:30:00Z"
}
```

#### `GET /works/{work_id}/reports`
Получение отчетов анализа по работе.

**Response:** `200 OK`
```json
[
  {
    "id": 1,
    "status": "done",
    "plagiarism_flag": true,
    "score": 0.85,
    "created_at": "2024-01-15T10:30:05Z",
    "updated_at": "2024-01-15T10:30:10Z"
  }
]
```

#### `GET /healthz`
Healthcheck endpoint.

**Response:** `200 OK`
```json
{"status":"ok"}
```

### File Store API (порт 8081)

#### `POST /works`
Загрузка файла (аналогично Gateway).

#### `GET /works/{id}`
Получение метаданных работы.

#### `GET /works/{id}/file`
Скачивание файла.

**Response:** `200 OK`
- Content-Type: MIME тип файла
- Body: содержимое файла

#### `GET /healthz`
Healthcheck endpoint.

### File Analysis API (порт 8082)

#### `POST /analysis/jobs`
Создание задачи анализа.

**Request:**
```json
{
  "work_id": 1,
  "assignment_id": "assignment1",
  "student_id": "student1",
  "file_hash": "abc123...",
  "storage_path": "assignment1/work.txt"
}
```

**Response:** `202 Accepted`
```json
{"status":"accepted"}
```

#### `GET /works/{work_id}/reports`
Получение отчетов анализа.

#### `GET /healthz`
Healthcheck endpoint.

## 🛠️ Разработка

### Структура проекта

```
File-storing-service/
├── gateway/              # API Gateway сервис
│   ├── cmd/server/       # Точка входа
│   └── internal/
│       ├── client/       # HTTP клиенты к другим сервисам
│       └── httpapi/     # HTTP handlers и роутинг
├── file-store/           # File Storing Service
│   ├── cmd/server/
│   └── internal/
│       ├── domain/      # Доменные модели и ошибки
│       ├── repo/        # Репозитории (in-memory)
│       ├── service/     # Бизнес-логика
│       ├── storage/     # Хранилище файлов (in-memory)
│       ├── validator/   # Валидация входных данных
│       └── httpapi/     # HTTP handlers
├── file-analysis/       # File Analysis Service
│   ├── cmd/server/
│   └── internal/
│       ├── domain/      # Доменные модели
│       ├── analyzer/    # Алгоритм обнаружения плагиата
│       ├── repo/        # Репозитории
│       ├── service/     # Бизнес-логика
│       ├── storage/     # Хранилище отчетов
│       ├── worker/      # Worker для обработки задач
│       └── httpapi/    # HTTP handlers
├── pkg/                 # Общие пакеты
│   ├── config/         # Конфигурация
│   ├── httpx/          # HTTP утилиты
│   └── logger/         # Логирование
└── deploy/             # Docker конфигурация
    ├── docker-compose.yml
    └── *.Dockerfile
```

### Команды Make

```bash
make build    # Собрать все сервисы
make test     # Запустить все тесты
make tidy     # Обновить зависимости
make up       # Запустить через docker-compose
make down     # Остановить сервисы
make logs     # Просмотр логов
make clean    # Удалить артефакты сборки
```

### Добавление новых тестов

```bash
# Создайте файл *_test.go в нужном пакете
# Запустите тесты
go test ./...

# С покрытием
go test -cover ./...

# С verbose выводом
go test -v ./...
```

## 🔧 Устранение неполадок

### Сервисы не запускаются

1. Проверьте, что порты 8080, 8081, 8082 свободны:
```bash
lsof -i :8080
lsof -i :8081
lsof -i :8082
```

2. Проверьте логи:
```bash
make logs
# или
docker-compose -f deploy/docker-compose.yml logs
```

### Ошибки при загрузке файлов

- Убедитесь, что файл не превышает 20MB
- Проверьте, что MIME тип поддерживается (text/plain или docx)
- Проверьте, что все обязательные поля переданы (file, student_id, assignment_id)

### Анализ не выполняется

- Проверьте, что File Analysis сервис запущен
- Проверьте логи worker'а
- Убедитесь, что очередь задач не переполнена

## 📝 TODO (для production)

- [ ] Реализовать Postgres репозитории (миграции, sqlc/pgx)
- [ ] Реализовать MinIO/S3 storage драйвер
- [ ] Добавить парсинг DOCX для извлечения текста
- [ ] Добавить метрики (Prometheus)
- [ ] Добавить трейсинг (OpenTelemetry)
- [ ] Добавить интеграционные тесты
- [ ] Добавить graceful shutdown для worker
- [ ] Улучшить обработку ошибок и retry логику
- [ ] Добавить аутентификацию и авторизацию
- [ ] Добавить rate limiting

## 📄 Лицензия

[Укажите лицензию проекта]

## 👥 Авторы

[Укажите авторов проекта]
