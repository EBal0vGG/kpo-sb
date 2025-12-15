# Руководство по тестированию

Этот документ содержит подробные примеры для ручного тестирования всех endpoints системы.

## Предварительные требования

1. Убедитесь, что все сервисы запущены:
```bash
make up
# или локально
```

2. Проверьте healthcheck всех сервисов:
```bash
curl http://localhost:8080/healthz  # Gateway
curl http://localhost:8081/healthz  # File Store
curl http://localhost:8082/healthz  # File Analysis
```

## Сценарий 1: Полный цикл работы

### Шаг 1: Загрузка первой работы

```bash
# Создайте тестовый файл
cat > work1.txt << 'EOF'
This is my original assignment work.
It contains unique content written by me.
The topic is about distributed systems.
EOF

# Загрузите через Gateway
curl -v -X POST http://localhost:8080/works \
  -F "file=@work1.txt" \
  -F "student_id=student001" \
  -F "assignment_id=hw1" \
  | jq .
```

**Ожидаемый результат:**
- Статус: `201 Created`
- В ответе будет `id`, `hash`, `storage_path`
- Сохраните `id` для следующих шагов (например, `id=1`)

### Шаг 2: Проверка метаданных

```bash
# Замените 1 на реальный ID из предыдущего шага
curl http://localhost:8080/works/1 | jq .
```

### Шаг 3: Загрузка второй работы (идентичной)

```bash
# Скопируйте файл
cp work1.txt work2.txt

# Загрузите от другого студента
curl -v -X POST http://localhost:8080/works \
  -F "file=@work2.txt" \
  -F "student_id=student002" \
  -F "assignment_id=hw1" \
  | jq .
```

**Ожидаемый результат:**
- Статус: `201 Created`
- Новый `id` (например, `id=2`)

### Шаг 4: Проверка отчетов анализа

```bash
# Подождите 3-5 секунд для обработки, затем:
curl http://localhost:8080/works/2/reports | jq .
```

**Ожидаемый результат:**
```json
[
  {
    "id": 1,
    "status": "done",
    "plagiarism_flag": true,
    "score": 1.0,
    "created_at": "...",
    "updated_at": "..."
  }
]
```

**Объяснение:** Так как файлы идентичны, хеш совпадает → `plagiarism_flag=true`, `score=1.0`

## Сценарий 2: Высокая схожесть текста

### Шаг 1: Загрузка оригинальной работы

```bash
cat > original.txt << 'EOF'
Distributed systems are collections of independent computers
that appear to users as a single coherent system. They provide
fault tolerance, scalability, and resource sharing capabilities.
Modern examples include cloud computing platforms and blockchain networks.
EOF

curl -X POST http://localhost:8080/works \
  -F "file=@original.txt" \
  -F "student_id=alice" \
  -F "assignment_id=cs101" \
  | jq .
```

### Шаг 2: Загрузка похожей работы

```bash
cat > similar.txt << 'EOF'
Distributed systems are groups of independent computers
that appear to users as a single coherent system. They provide
fault tolerance, scalability, and resource sharing capabilities.
Modern examples include cloud computing platforms and blockchain networks.
EOF

curl -X POST http://localhost:8080/works \
  -F "file=@similar.txt" \
  -F "student_id=bob" \
  -F "assignment_id=cs101" \
  | jq .
```

### Шаг 3: Проверка результатов

```bash
# Подождите обработки
sleep 5
curl http://localhost:8080/works/2/reports | jq .
```

**Ожидаемый результат:**
- `plagiarism_flag`: `true` (если схожесть >= 0.7)
- `score`: значение от 0.7 до 1.0

## Сценарий 3: Низкая схожесть (нет плагиата)

### Шаг 1: Загрузка разных работ

```bash
cat > work_math.txt << 'EOF'
Linear algebra is fundamental to machine learning.
Matrices and vectors are used to represent data.
Eigenvalues and eigenvectors have important applications.
EOF

cat > work_cooking.txt << 'EOF'
Italian cuisine is known for its simplicity and fresh ingredients.
Pasta dishes are popular worldwide.
Traditional recipes have been passed down for generations.
EOF

# Загрузите обе работы
curl -X POST http://localhost:8080/works \
  -F "file=@work_math.txt" \
  -F "student_id=math_student" \
  -F "assignment_id=essay1" \
  | jq .

curl -X POST http://localhost:8080/works \
  -F "file=@work_cooking.txt" \
  -F "student_id=cooking_student" \
  -F "assignment_id=essay1" \
  | jq .
```

### Шаг 2: Проверка результатов

```bash
sleep 5
curl http://localhost:8080/works/2/reports | jq .
```

**Ожидаемый результат:**
- `plagiarism_flag`: `false`
- `score`: значение < 0.7

## Сценарий 4: Прямая работа с File Store

### Загрузка напрямую в File Store

```bash
curl -X POST http://localhost:8081/works \
  -F "file=@work1.txt" \
  -F "student_id=direct_student" \
  -F "assignment_id=direct_assignment" \
  | jq .
```

### Получение метаданных

```bash
curl http://localhost:8081/works/1 | jq .
```

### Скачивание файла

```bash
curl http://localhost:8081/works/1/file -o downloaded.txt
cat downloaded.txt
```

## Сценарий 5: Прямая работа с File Analysis

### Создание задачи анализа вручную

```bash
curl -X POST http://localhost:8082/analysis/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "work_id": 1,
    "assignment_id": "hw1",
    "student_id": "student001",
    "file_hash": "abc123def456...",
    "storage_path": "hw1/work1.txt"
  }' | jq .
```

**Примечание:** Замените `file_hash` и `storage_path` на реальные значения из метаданных работы.

### Проверка статуса обработки

```bash
# Проверяйте несколько раз, статус будет меняться: pending -> running -> done
curl http://localhost:8082/works/1/reports | jq .
```

## Сценарий 6: Обработка ошибок

### Попытка загрузить слишком большой файл

```bash
# Создайте файл > 20MB (если нужно)
dd if=/dev/zero of=large.txt bs=1M count=21

curl -v -X POST http://localhost:8080/works \
  -F "file=@large.txt" \
  -F "student_id=student1" \
  -F "assignment_id=hw1"
```

**Ожидаемый результат:**
- Статус: `413 Request Entity Too Large`
- Сообщение об ошибке

### Попытка загрузить неподдерживаемый тип файла

```bash
# Создайте PDF файл (не поддерживается)
echo "fake pdf content" > test.pdf

curl -v -X POST http://localhost:8080/works \
  -F "file=@test.pdf" \
  -F "student_id=student1" \
  -F "assignment_id=hw1"
```

**Ожидаемый результат:**
- Статус: `415 Unsupported Media Type`

### Запрос несуществующей работы

```bash
curl -v http://localhost:8080/works/99999
```

**Ожидаемый результат:**
- Статус: `404 Not Found` или `502 Bad Gateway`

### Отсутствие обязательных полей

```bash
curl -v -X POST http://localhost:8080/works \
  -F "file=@work1.txt"
  # student_id и assignment_id отсутствуют
```

**Ожидаемый результат:**
- Статус: `400 Bad Request`
- Сообщение об ошибке

## Сценарий 7: Массовая загрузка

### Загрузка нескольких работ

```bash
for i in {1..5}; do
  echo "Work number $i" > work_$i.txt
  curl -X POST http://localhost:8080/works \
    -F "file=@work_$i.txt" \
    -F "student_id=student$i" \
    -F "assignment_id=hw1" \
    | jq -r '.id'
done
```

### Проверка всех отчетов

```bash
# Для каждой работы проверьте отчеты
for work_id in 1 2 3 4 5; do
  echo "=== Work $work_id ==="
  curl -s http://localhost:8080/works/$work_id/reports | jq .
  echo
done
```

## Полезные команды для отладки

### Просмотр логов

```bash
# Docker Compose
make logs

# Или конкретный сервис
docker-compose -f deploy/docker-compose.yml logs gateway
docker-compose -f deploy/docker-compose.yml logs file-store
docker-compose -f deploy/docker-compose.yml logs file-analysis
```

### Проверка статусов контейнеров

```bash
docker-compose -f deploy/docker-compose.yml ps
```

### Очистка данных (перезапуск)

```bash
make down
make up
```

## Автоматизация тестирования

### Скрипт для базового smoke test

```bash
#!/bin/bash
set -e

BASE_URL="http://localhost:8080"

echo "1. Healthcheck"
curl -s $BASE_URL/healthz | jq .

echo "2. Upload work"
RESPONSE=$(curl -s -X POST $BASE_URL/works \
  -F "file=@work1.txt" \
  -F "student_id=test_student" \
  -F "assignment_id=test_assignment")

WORK_ID=$(echo $RESPONSE | jq -r '.id')
echo "Work ID: $WORK_ID"

echo "3. Get work metadata"
curl -s $BASE_URL/works/$WORK_ID | jq .

echo "4. Wait for analysis..."
sleep 5

echo "5. Get reports"
curl -s $BASE_URL/works/$WORK_ID/reports | jq .

echo "Test completed!"
```

Сохраните как `test.sh`, сделайте исполняемым и запустите:
```bash
chmod +x test.sh
./test.sh
```

## Примечания

- Все примеры используют `jq` для форматирования JSON. Если его нет, установите:
  ```bash
  # Ubuntu/Debian
  sudo apt-get install jq
  
  # macOS
  brew install jq
  ```

- Для Windows используйте Git Bash или WSL для выполнения bash команд

- Время обработки анализа может варьироваться в зависимости от размера файла и нагрузки

- Для тестирования больших файлов создайте их с помощью:
  ```bash
  dd if=/dev/urandom of=large.txt bs=1M count=10  # 10MB файл
  ```

