- Статус: готов
- Версия: 1
- Запуск: 1
- Входы: orchestrator/task-brief v2, business-analyst/step-5/domain-model v2, business-analyst/step-5/requirements v3, software-architect/step-7/architecture v5, software-architect/step-8/implementation-plan v4 (разделы «Исполнители и разрешённые пути», plan-4, plan-5), software-architect/questions.md (Q-SA1: «(а) uint8 + Equal»), business-analyst/questions.md (Q-8 (а)), orchestrator/questions.md (О-7, О-10, О-11, О-12), docs/rules/main.md, docs/rules/dev.md, .golangci.yml, .github/workflows/ci.yml, README.md, ROADMAP.md (только чтение), go.mod (go 1.27.3)

# Описание изменений: реализация пакета `letter` и README (plan-4, plan-5)

## Что сделано
- **plan-4.** Реализован весь публичный API пакета `letter` по контракту architecture v5:
  - `type Letter uint8`; константы `C = 1 … B = 7` (`1 + iota`); `const Count = 7` (нетипизированная); `var ErrInvalid = errors.New("letter: invalid letter")`;
  - функции `All`, `FromOrdinal`, `MustFromOrdinal`; методы `IsValid`, `Equal`, `Ordinal`, `DistanceUp`, `DistanceDown`, `Shift`, `Cycle`, `String`;
  - получатели и результаты по значению — исключение из правила dev.md по решению человека Q-SA1;
  - недопустимые значения — по таблице контракта: `IsValid`/`Equal` → `false` (в том числе `z.Equal(z)`), `Ordinal`/`DistanceUp`/`DistanceDown` → `-1`, `Shift` → нулевое `Letter`, `Cycle` → `nil`, `String` → `"Letter(<n>)"`; `FromOrdinal` вне 0…6 → нулевое `Letter` и ошибка, обёртывающая `ErrInvalid` (`%w`, текст с переданным значением); `MustFromOrdinal` — обёртка над `FromOrdinal`, паникует значением этой ошибки;
  - `Shift` без переполнения: `steps % Count` вычисляется до сложения, промежуточная сумма в диапазоне −6…12, затем приведение в 0…6;
  - `All` и `Cycle` возвращают новый срез длины и ёмкости `Count` на каждый вызов; `Cycle()[i] == l.Shift(i)`;
  - глобального изменяемого состояния нет: приватная таблица символов `[Count]string` только читается;
  - импорты — только стандартная библиотека: `errors`, `fmt`, `strconv`;
  - doc-комментарии на американском английском у пакета и у каждого экспортируемого идентификатора; диапазон 0…6, период 7 и семь элементов названы свойствами seven-letter set (НФ-1 п. 2, К-11 б); тип `Letter` — «a letter of the seven-letter set»; описаны стабильные числа констант, непроверяемое `Letter(n)`, арифметика ≠ сдвиг, владение срезами, конкурентность, текст ошибок вне API.
- **plan-5.** README.md (НФ-6, К-16):
  - баннер «Early development»: убрана фраза «Musical functionality is not implemented yet»; предупреждение о неготовности к промышленному использованию и неустановленном API сохранено; добавлено, что реализован только пакет `letter` и он не выпущен;
  - Features, Concept, Examples — подраздел «Letters (unreleased)» в каждом; только реализованное в `letter`;
  - Examples — «вопрос — краткое объяснение — ссылка» на файл `letter/example_test.go` с именами примеров из таблицы architecture (`ExampleAll`, `ExampleLetter_Cycle`, `ExampleLetter_DistanceUp`, `ExampleLetter_DistanceDown`, `ExampleLetter_Shift`, `ExampleLetter_Equal`, `ExampleFromOrdinal`, `ExampleMustFromOrdinal`, `ExampleLetter_IsValid`); имена подтверждены по выводу `go test -v` (тесты не читались);
  - комментарии-правила README сохранены без изменений; Use Cases и значок Roadmap не тронуты.

## Изменённые компоненты и файлы
| Файл | Изменение | `git hash-object` |
|---|---|---|
| `letter/letter.go` | новый: тип, константы, функции, методы, приватные `symbols`, `wrap` | `b88c88f31f51bf11369f46a17bfa76c4dec89c3c` |
| `letter/doc.go` | новый: документация пакета | `fda44b066090ab6e5ca594a2c6697081d293136c` |
| `README.md` | баннер, Features, Concept, Examples | `870ee31350a86bca865b5e1ef8ba6d1f8a22329c` |
| `tasks/2026-11-01-letter/go-code-developer/questions.md` | новый: Q-CD1 | — |

- `go.mod`, `go.sum` — не изменены этим поручением (`go mod tidy` без изменений, сверено с копией до начала работы); Testify v1.11.1 на месте.
- Тесты, ROADMAP.md, docs/, .golangci.yml, .github/, scripts/, CONTRIBUTING.md — не менялись.

## Исходное состояние тестов
- Окружение: macOS (darwin/arm64), `go version go1.27.3 darwin/arm64`, `./.cache/bin/golangci-lint` v2.15.0; ветка `feat/1-add-letter-package`.
- До начала работы (продуктовых файлов `letter` не было):
  - `go build ./...` — код выхода 0;
  - `go vet ./...` — `github.com/go-muse/musex/letter: no non-test Go files in …/musex/letter`;
  - `go test -race -count=1 ./...` — `FAIL github.com/go-muse/musex/letter [build failed]` с той же диагностикой.
- Других пакетов в модуле нет; иных падений нет. Причина единственная — ожидаемое отсутствие реализации.

## Как проверить
Команды из корня репозитория; фактические результаты после последней правки:

| # | Команда | Результат |
|---|---|---|
| 1 | `./.cache/bin/golangci-lint fmt --diff letter/doc.go letter/letter.go` | пустой вывод, код 0 (также `gofmt -l` — пусто) |
| 2 | `go build ./...` | код 0 |
| 3 | `go vet ./...` | без замечаний, код 0 |
| 4 | `go test -race -count=1 ./...` | `ok github.com/go-muse/musex/letter`; 44 верхнеуровневых теста и примера PASS (32 теста, 12 примеров), всего 455 строк PASS с подтестами, FAIL нет |
| 5 | `./.cache/bin/golangci-lint run ./...` | **3 замечания staticcheck QF1010, все в тестовых файлах**: `letter/example_test.go:23:14`, `letter/example_test.go:32:14`, `letter/letter_test.go:502:49`; в `letter/letter.go`, `letter/doc.go` — 0 замечаний. Вопрос Q-CD1 автору тестов |
| 6 | `go mod tidy` | код 0; go.mod и go.sum побайтно совпадают с состоянием до начала работы |
| 7 | `go doc -all ./letter` | код 0; документированы пакет и все экспортируемые идентификаторы: `Count`, `ErrInvalid`, `Letter`, `C…B`, `All`, `FromOrdinal`, `MustFromOrdinal`, `Cycle`, `DistanceDown`, `DistanceUp`, `Equal`, `IsValid`, `Ordinal`, `Shift`, `String` |
| — | `go list -f '{{join .Imports " "}}' ./letter` | `errors fmt strconv` (К-12) |

- README: `git diff -- README.md` — изменены только строка баннера и содержимое под комментариями Features, Concept, Examples.

## Изменения существующих тестов
- Нет: тесты не изменялись и не читались.

## Отклонения от плана
- Нет. Добавлены приватные помощники `symbols` (таблица только для чтения) и `wrap` (приведение в 0…6) — внутренняя реализация, публичный контракт не затронут.

## Учтённые отчёты
- Отчёты review-code / test-report к этому поручению ещё не выдавались.

## Известные ограничения
- **Проверка 5 plan-4 (К-15) не чистая** из-за QF1010 в тестовых файлах — Q-CD1 в `go-code-developer/questions.md`. Причина — базовый тип `uint8` у `Letter` по контракту (Q-SA1): линтер принимает `[]Letter` за срез байтов; это ложное срабатывание (элементы реализуют `fmt.Stringer`, примеры с `// Output:` проходят). Со стороны реализации не устраняется без изменения контракта. То же замечание получат будущие потребители, печатающие `[]letter.Letter` через `fmt`, — для сведения архитектора.
- Сборка проверена только на darwin/arm64 (64 бита); 32-битные платформы не поддерживаются (О-7).
- `govulncheck` не запускался (проверка QA по плану).

## Допущения
- В README ссылки на примеры ведут на файл `letter/example_test.go` (допустимо до выпуска по architecture Q-A5, Н-8), а не на pkg.go.dev.
- Формулировка интервала в Concept README («from C up to E the letter distance is 2, while the interval is a third») — по глоссарию domain-model v2 («ступеневая величина = расстояние + 1»).

## Затронутая документация
- README.md — обновлён в этом поручении (plan-5).
- ROADMAP.md — после приёмки человеком: отметка п. 1 и вводная фраза (plan-6, отдельное поручение); сейчас не изменён.
