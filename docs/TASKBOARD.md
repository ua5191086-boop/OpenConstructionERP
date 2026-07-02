# TASKBOARD — очередь работ для всех сессий и разработчиков

**Правило захвата:** перед началом работы сессия/разработчик ПЕРВЫМ коммитом меняет
статус задачи на `in_progress (владелец, дата)` и пушит. Только потом пишет код.
Кто первым запушил захват — тот владеет задачей. Конфликт захвата = git сам скажет.
Закончил — статус `done (коммит)`. Новые задачи добавлять в конец соответствующего блока.

**Перед любой работой:** `git pull --rebase`. Миграции — только через реестр в WORKSTREAMS.md.

## Готово (проверено e2e)
| Блок | SAD | Статус |
|------|-----|--------|
| Онтологическое ядро (types/objects/links/graph) | §3 | done |
| BOQ: импорт RU/EN, свод, коэффициенты, экспорт | B-04 | done |
| Конвертер сравнительных ведомостей (реальный TTZ $431M внутри) | B-04 | done |
| Тоннель: проходки, кольца, темп, прогноз сбойки | L-01/03/04 | done |
| Стоимость: транзакции, план-факт, версии бюджета | F-01/02 | done |
| RFI + суточные рапорты + физпрогресс (EV) | D-03, C-05 | done |
| Executive report (.xlsx одним вызовом) | P-01 | done |
| CDE: нумерация, ревизии, ISO 19650 статусы, transmittals | D-01 | done (verified 02.07) |
| Тендерный модуль: схема+генератор+дашборд | B-02 | done (schema fixed) |
| Схемы V007–V012 (contracts/HR/finance/procurement/BIM/AI) | G,J,F,H,O,P | schema only — API нет |
| React frontend каркас, compose single-node, Go scaffold | — | in progress (infra lane) |

## Очередь MVP (брать сверху)
| # | Задача | SAD | Статус |
|---|--------|-----|--------|
| 1 | NCR: реестр, workflow, связь с BOQ/кольцами | M-03 | done (V028 + ncr_hse router) |
| 2 | Permit to Work: выдача/активация/закрытие, борд активных | N-01 | done (V028 + ncr_hse router) |
| 3 | Variation Orders (V034): workflow, включение в бюджет+Commitment | G-02 | done (money router) |
| 4 | IPC из earned value: retention/advance recovery, оплата→Actual | F-05 | done (money router) |
| 5 | Импорт XER (P6): TASK/TASKPRED/PROJWBS → V014, summary, critical filter | E-01 | done (schedule router) |
| 6 | API поверх V010 Procurement (заявка→PO→приёмка→склад) | H-03, I-01 | todo |
| 7 | Мобильный контур: offline-форма суточного рапорта (PWA до Flutter) | §2 Mobile | todo |
| 8 | Инструментальный мониторинг: КИА, пороги, алерты | L-12/13 | todo |
| 9 | Сравнение вариантов BOQ внутри системы (дельты по секциям) | B-06 | todo |
| 10 | Auth на API: Keycloak OIDC (JWKS) + static HS256 для dev/CI, 401 повсюду без токена | A-02 | done (auth middleware, CI проверяет 401/200) |
| 11 | React frontend: подключить к API вертикалей (BOQ/tunnel/CDE) | §2 Frontend | todo |
| 12 | Том 2 SAD: спецификации модулей L и F доменов | док | todo |

## Правила против затирания (кратко, полные — WORKSTREAMS.md)
1. Один блок = одна полоса = один владелец. 2. Захват — первым коммитом.
3. Номер миграции — из реестра WORKSTREAMS в том же коммите.
4. main всегда зелёный: CI ставит всю цепочку миграций с нуля.
