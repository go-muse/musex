# Вопросы go-code-developer — задача 2026-11-01-letter

Журнал по artifacts/questions. Записи не удаляются и не переписываются; ответы дописываются в поле «Ответ».
Записано: поручение `code` (plan-4, plan-5), запуск 1, 2026-11-02.

---

### Q-CD1. `golangci-lint run ./...` выдаёт staticcheck QF1010 в тестовых файлах
- Адресат: go-test-developer (через оркестратора); для сведения — software-architect (см. «Зачем нужен ответ», п. 3)
- Требование: К-15 / НФ-5 — `golangci-lint` по `.golangci.yml` проходит без замечаний; plan-4, проверка 5 — `golangci-lint run ./...` без замечаний.
- Наблюдение (диагностика собственного запуска `./.cache/bin/golangci-lint run ./...`, v2.15.0; исходники тестов не открывались):
  - `letter/example_test.go:23:14: QF1010: could convert argument to string (staticcheck)` — строка `fmt.Println(all)`;
  - `letter/example_test.go:32:14: QF1010: could convert argument to string (staticcheck)` — строка `fmt.Println(letter.A.Cycle())`;
  - `letter/letter_test.go:502:49: QF1010: could convert argument to string (staticcheck)` — в вызове `fmt.Sprint(letter.All())`.
  - В продуктовых файлах `letter/letter.go`, `letter/doc.go` замечаний нет. `go vet ./...` и `go test -race -count=1 ./...` проходят.
- Сценарий и ожидаемый результат: вывод `[]letter.Letter` через `fmt` печатает `[C D E F G A B]` (architecture v5, раздел `String`: «`fmt.Println(letter.All())` печатает `[C D E F G A B]`»); при этом `golangci-lint run ./...` — без замечаний.
- Предполагаемая причина: QF1010 срабатывает на срез, у элементов которого базовый тип `byte`; `Letter` по контракту — `type Letter uint8` (Q-SA1). Для `[]Letter` это ложное срабатывание: `fmt` печатает элементы через `String` (примеры с `// Output:` проходят), а предлагаемое линтером преобразование в `string` дало бы байты `\x01…\x07` — неверный вывод. Со стороны реализации не устраняется без изменения контракта (базового типа `Letter`), что запрещено.
- Вопрос: как привести тесты и примеры к прохождению `golangci-lint run ./...` без изменения проверяемого вывода — например, `//nolint:staticcheck // QF1010: false positive, elements implement fmt.Stringer` на этих строках или иной способ печати, не меняющий `// Output:`? Выбор способа — за автором тестов.
- Зачем нужен ответ:
  1. без исправления К-15 и проверка 5 plan-4 не выполнены, QA (полный объём) не получит чистый `golangci-lint run ./...`;
  2. правка тестов — только автора тестов, затем review-tests;
  3. для архитектора: то же замечание получит любой потребитель в musex, печатающий `[]letter.Letter` через `fmt` (spelling, notation и др.). Возможно, стоит упомянуть это в architecture (Риски) как известное свойство `type Letter uint8` при текущем `.golangci.yml`; решение — за архитектором.
- Вариант по умолчанию: реализация не меняется; поручение `code` сдаётся с этим открытым вопросом; переход к QA — после исправления тестов и review-tests.
- Ответ и кто ответил: go-test-developer, 2026-11-02 (поручение `q-cd1`, запуск 1):
  - Запрос верен. QF1010 здесь — ложное срабатывание: `fmt` печатает элементы `[]letter.Letter` через `String`, а предложенное линтером преобразование в `string` дало бы неверный вывод. Реализация и контракт не меняются.
  - Тесты исправлены без `nolint`: в трёх местах вывод теперь идёт через явный глагол `%v` — `fmt.Printf("%v\n", …)` в `ExampleAll` и `ExampleLetter_Cycle`, `fmt.Sprintf("%v", …)` в `TestLetter_String`. По документации `fmt` для одного операнда это тот же формат, что у `Println` и `Sprint`, поэтому `// Output:` и ожидаемое значение `[C D E F G A B]` не изменились. На явный глагол QF1010 не срабатывает.
  - Проверки после правки: `./.cache/bin/golangci-lint run ./...` — 0 issues; `go vet ./...` — без замечаний; `go test -race -count=1 ./...` — ok, все 12 примеров PASS; `golangci-lint fmt --diff letter/*_test.go` — пусто.
  - Подробности: go-test-developer/q-cd1/change-summary.md, версия 1. Дальше — review-tests.
  - Для архитектора (п. 3 вопроса): потребители, печатающие `[]letter.Letter` через `Println` или `Sprint` без глагола, получат то же замечание; обход — явный `%v`. Упоминать ли это в architecture, решает архитектор.
