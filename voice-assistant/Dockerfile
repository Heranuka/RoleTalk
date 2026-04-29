# Использование максимально легкого базового образа
FROM python:3.10-slim as builder

# Настройка окружения
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

WORKDIR /app

# Устанавливаем зависимости системы
# Мы ставим build-essential только для сборки и сразу удалим его
RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Копируем requirements
COPY requirements.txt .

# Устанавливаем ТОЛЬКО CPU версию PyTorch и библиотеки
# Кэш pip удаляем сразу (--no-cache-dir)
RUN pip install --no-cache-dir torch --index-url https://download.pytorch.org/whl/cpu && \
    pip install --no-cache-dir -r requirements.txt

# Удаляем инструменты сборки, чтобы облегчить образ
RUN apt-get purge -y --auto-remove build-essential

# Копируем код
COPY . .

# Финальный этап: создаем чистый слой
FROM python:3.10-slim
WORKDIR /app

# Копируем из билдера только установленные пакеты и код
COPY --from=builder /usr/local/lib/python3.10/site-packages /usr/local/lib/python3.10/site-packages
COPY --from=builder /usr/local/bin /usr/local/bin
COPY --from=builder /app /app

# Устанавливаем только ffmpeg (нужен для работы)
RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg \
    && rm -rf /var/lib/apt/lists/*

EXPOSE 8080

CMD ["python", "scripts/assistant_api.py"]