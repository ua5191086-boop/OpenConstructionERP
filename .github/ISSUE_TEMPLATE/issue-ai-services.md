---
title: "[SERVICE] AI Services — FastAPI microservice"
labels: enhancement, service, ai, fastapi
assignees: ""
---

## Описание AI Services

Python FastAPI микросервис для AI/ML операций: классификация, прогнозирование, извлечение данных и рекомендации.

### FastAPI приложение
- ✅ `services/ai-svc/main.py`:
  - `GET /api/v1/ai/health` — health check
  - `POST /api/v1/ai/classify` — классификация запроса агентами
  - `POST /api/v1/ai/predict` — прогноз стоимости/сроков/рисков
  - `POST /api/v1/ai/extract` — извлечение данных из текста (суммы, даты, проценты, коды)
  - `POST /api/v1/ai/recommend` — рекомендации по контексту
  - Pydantic models, CORS middleware, stub AI (keyword-based classifier, formula-based predictor)

### Docker
- ✅ `services/ai-svc/Dockerfile` — python:3.12-slim, uvicorn, healthcheck
- ✅ Порт: 8100

### Интеграция
- ✅ Обновлён `docker-compose.dev.yml` — добавлен сервис ai-svc