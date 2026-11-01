- Статус: принят
- Версия: 1
- Запуск: 1
- Входы: software-architect/step-8/implementation-plan v4 (раздел plan-6), orchestrator/questions.md (О-15: приёмка человеком 2026-11-02), business-analyst/step-5/requirements v3 (К-16, «Ограничения»), ROADMAP.md (редакция `c8ead20`, до правки), docs/rules/main.md («Языки»)

# Описание изменений: ROADMAP.md после приёмки (plan-6)

## Что сделано
- Пункт 1 отмечен как реализованный: `- [ ] **1. \`letter\`**` → `- [x] **1. \`letter\`**`.
- Вводная фраза «All packages are **planned**; none is implemented yet.», ставшая неверной, заменена на «Checked packages are implemented; the others are **planned**.» (английский). Соседние предложения абзаца не изменены; выделение **planned** сохранено.
- Остальные пункты и текст не менялись.

## Изменённые компоненты и файлы
| Файл | Изменение | `git hash-object` |
|---|---|---|
| `ROADMAP.md` | две правки выше | `a9b0a1dc0218654209b1c13f7c2ff79409bf4c94` (до правки — `c8ead20`) |

## Дифф
```diff
-`musex` is planned as a set of focused Go packages across three areas. This roadmap shows the intended package order and scope. All packages are **planned**; none is implemented yet. Check a package off when its agreed API and behavior are implemented, reviewed, and accepted.
+`musex` is planned as a set of focused Go packages across three areas. This roadmap shows the intended package order and scope. Checked packages are implemented; the others are **planned**. Check a package off when its agreed API and behavior are implemented, reviewed, and accepted.
@@
-- [ ] **1. `letter`** — the seven note letters and their cyclic order.
+- [x] **1. `letter`** — the seven note letters and their cyclic order.
```

## Как проверить
- `git diff -- ROADMAP.md` — ровно две изменённые строки (`git diff --stat`: 2 insertions, 2 deletions): вводный абзац и пункт 1.
- Затем plan-7: review-code только по этому диффу.

## Исходное состояние тестов
- Не применимо: изменён только Markdown-файл; Go-код, тесты и конфигурация не затронуты, проверки Go не запускались.

## Изменения существующих тестов
- Нет.

## Отклонения от плана
- Нет.

## Учтённые отчёты
- Нет (отчётов к plan-6 ещё не было).

## Известные ограничения
- Нет.

## Допущения
- Новая фраза выбрана так, чтобы оставаться верной по мере отметки следующих пунктов, без правки при каждом шаге roadmap (пример в плане: «пакеты отмечаются по мере реализации»).

## Затронутая документация
- ROADMAP.md — изменение само является обновлением документации.
