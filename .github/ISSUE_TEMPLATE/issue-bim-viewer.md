---
title: "[APP] BIM Viewer — Three.js + IFC.js viewer"
labels: enhancement, app, bim, viewer
assignees: ""
---

## Описание BIM Viewer

Веб-приложение для просмотра IFC-моделей с использованием Three.js и web-ifc-viewer.

### Веб-приложение
- ✅ `apps/bim-viewer/index.html`:
  - Загрузка IFC-файла через file input
  - Просмотр 3D модели (Three.js + web-ifc-viewer через CDN)
  - Дерево элементов с группировкой по типу
  - Выбор элемента → отображение свойств
  - Орбитальный контрол (поворот, масштабирование, панорамирование)
  - Кнопки: центрировать, каркасный режим
  - Fallback: если IFC.js не загрузился — демо Three.js с кубом
  - Тёмная тема (0f172a)

### Docker
- ✅ `apps/bim-viewer/Dockerfile` — nginx:alpine, healthcheck
- ✅ Порт: 8200

### Интеграция
- ✅ Обновлён `docker-compose.dev.yml` — добавлен сервис bim-viewer