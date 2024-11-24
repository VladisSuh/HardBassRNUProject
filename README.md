# HardBassRNUProject

## Описание
Проект HardBassRNUProject включает серверное приложение на Go и KeyDB для хранения данных. Серверное приложение запускается в контейнере Docker и может быть настроено с помощью переменных окружения.

## Требования
- Docker
- Docker Compose

## Установка и запуск

1. Клонируйте репозиторий:
    ```sh
    git clone https://github.com/VladisSuh/HardBassRNUProject.git
    cd HardBassRNUProject
    ```

2. Запустите контейнеры с помощью Docker Compose:
    ```sh
    PORT=9090 STORAGE=./data docker-compose up --build -d
    ```

    Переменные окружения:
    - `PORT`: Порт, на котором будет запущен сервер (по умолчанию 6382).
    - `STORAGE`: Путь к директории для хранения данных (по умолчанию [`./data`]('./data')).
3. Остановка контейнеров:
    ```sh
    docker-compose down
    ```

4. Остановка контейнеров и удаление данных:
    ```sh
    docker-compose down -v
    ```

5. Перезапуск контейнеров:
    ```sh
    docker-compose restart
    ```

6. Запуск для клиента:
   ```sh
    go run cmd/client/main.go -port=6372 -file=example.txt
    ```