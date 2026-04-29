FROM python:3.10-slim
WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg build-essential && \
    rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
# Ставим torch для CPU и остальные либы
RUN pip install --no-cache-dir torch --index-url https://download.pytorch.org/whl/cpu && \
    pip install --no-cache-dir -r requirements.txt

COPY . .
ENV PYTHONPATH=/app
ENV PYTHONUNBUFFERED=1

EXPOSE 8080
CMD ["python", "-m", "src.main"]