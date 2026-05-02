FROM python:3.10-slim

ENV PYTHONPATH=/app
ENV PYTHONUNBUFFERED=1
ENV PYTHONDONTWRITEBYTECODE=1

WORKDIR /app

# Добавляем python3-dev (нужен для компиляции некоторых расширений)
RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg \
    build-essential \
    python3-dev \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .

# 1. Сначала обновляем pip, setuptools и wheel (критично для новых версий protobuf)
# 2. Ставим torch
# 3. Ставим всё остальное
RUN pip install --no-cache-dir --upgrade pip setuptools wheel && \
    pip install --no-cache-dir torch --index-url https://download.pytorch.org/whl/cpu && \
    pip install --no-cache-dir -r requirements.txt

COPY . .

EXPOSE 50051

CMD ["python", "-m", "src.main"]