- Статус: готов
- Версия: 1
- Запуск: 1
- Входы: letter/letter.go b88c88f31f51bf11369f46a17bfa76c4dec89c3c, letter/doc.go fda44b066090ab6e5ca594a2c6697081d293136c, README.md 870ee31350a86bca865b5e1ef8ba6d1f8a22329c, go-code-developer/code/change-summary v1, software-architect/step-7/architecture v5, business-analyst/step-5/requirements v3, software-architect/step-8/implementation-plan v4, orchestrator/task-brief v2, business-analyst/step-5/domain-model v2; для согласованности — тесты пакета в редакциях example_test.go f865f94, letter_test.go 59b2746, concurrency_test.go bdb3281, helpers_test.go 920fdf5, imports_test.go c5efe53, invalid_test.go ef31914; go.mod db78bed, go.sum c4c1710

# Отчёт ревью: review-code, реализация `letter` и README (plan-4, plan-5)

Адресат: go-code-developer.

## Вердикт
**Принять.** Блокирующих замечаний нет.

## Что проверено
- **Контракт architecture v5.**
  - Сигнатуры и состав экспортируемого API совпадают со сводкой; лишних экспортируемых идентификаторов нет.
  - Тип `Letter uint8`; константы C = 1 … B = 7 (`1 + iota`); `Count` нетипизированная; текст `ErrInvalid` — как в контракте.
  - Таблица К-7 выполнена полностью:
    - `IsValid` и `Equal` → false, в том числе `z.Equal(z)`;
    - `Ordinal`, `DistanceUp`, `DistanceDown` → −1 для недопустимого получателя и аргумента;
    - `Shift` → нулевое `Letter`; `Cycle` → `nil`;
    - `String` → `"Letter(n)"` с десятичным `uint8`;
    - `FromOrdinal` → нулевое `Letter` и `%w`-обёртка `ErrInvalid`;
    - `MustFromOrdinal` — обёртка над `FromOrdinal`, паникует значением её ошибки (правило dev.md «Must»).
  - `Equal` совпадает с `==` на допустимых буквах и отличается только на недопустимых.
  - `Shift`: `steps % Count` вычисляется до сложения, сумма в −6…12, `wrap` покрывает −7…13. Ручная проверка: `C.Shift(math.MaxInt)` = C, `C.Shift(math.MinInt)` = B (MinInt % 7 = −1). Переполнения нет на всём `int`.
  - `All` и `Cycle` возвращают новый срез длины и ёмкости 7 на каждый вызов.
  - Изменяемого общего состояния нет: `symbols` — неэкспортируемый массив, только чтение; горутин нет. Конкурентная безопасность обеспечена построением.
  - Импорты: `errors fmt strconv` (К-12).
- **Правила docs/rules.**
  - Получатели по значению — исключение по решению Q-SA1.
  - Must только как обёртка.
  - `Equal` есть.
  - vendor нет; Go 1.27.3.
- **Документация (`go doc -all ./letter`).**
  - Американский английский; британских написаний не найдено.
  - Каждый экспортируемый идентификатор документирован.
  - К-11 б: диапазон 0…6, период 7 и семь элементов названы свойствами seven-letter set — в doc.go и в комментариях `Count`, `All`, `Cycle`, `DistanceUp`, `DistanceDown`, `Shift`, `Ordinal`, `FromOrdinal`. `Letter` — «a letter of the seven-letter set».
  - К-11 д: конструктора набора, параметра числа букв и изменяемой таблицы нет.
  - Описаны стабильные числа констант, непроверяемое `Letter(n)` и то, что арифметика не является сдвигом. Текст ошибки вне API.
- **README (НФ-6, К-16, plan-5).**
  - Дифф с `main` затрагивает только строку баннера и подразделы «Letters (unreleased)» в Features, Concept и Examples (строки 33, 50, 69).
  - В баннере (строка 21) нет «not implemented yet»; предупреждение о ранней стадии и неготовности к промышленному использованию сохранено.
  - Use Cases, значок Roadmap и комментарии-правила не изменены.
  - Examples оформлены как «вопрос — объяснение — ссылка». Все 9 упомянутых примеров существуют и проходят (`go test -v`).
  - Утверждения README сверены с API: `B.DistanceUp(D)` = 2, `C.DistanceDown(E)` = 5, `C.Shift(9)` = E, `C.Shift(-1)` = B, `FromOrdinal(4)` = G, `A.Cycle()` = A…G. Термины Concept согласованы с глоссарием domain-model v2.
- **Границы областей.**
  - Тестовые файлы автор реализации не менял. Изменения example_test.go и letter_test.go после реализации внесены автором тестов в поручении q-cd1, и хеши совпадают с его change-summary.
  - Строки тестов в Q-CD1 взяты из стандартного вывода golangci-lint: это допустимая диагностика, а не чтение тестов.
  - ROADMAP.md, .golangci.yml, .github/, CONTRIBUTING.md, docs/ и scripts/ не изменены.
- **go.mod и go.sum.** Хеши db78bed и c4c1710; `go mod tidy -diff` — пусто. Testify v1.11.1 — единственная прямая зависимость, только для тестов.

## Результаты команд (из корня, go1.27.3 darwin/arm64)
| Команда | Результат |
|---|---|
| `go build ./...` | код 0 |
| `go vet ./...` | без замечаний, код 0 |
| `go test -race -count=1 ./...` | `ok github.com/go-muse/musex/letter`; 44 верхнеуровневых теста и примера PASS (12 примеров), FAIL нет |
| `./.cache/bin/golangci-lint run ./...` (v2.15.0) | `0 issues.` |
| `./.cache/bin/golangci-lint fmt --diff ./...` | пусто, код 0 |
| `gofmt -l letter` | пусто |
| `go mod tidy -diff` | пусто, код 0 |
| `go list -f '{{join .Imports " "}}' ./letter` | `errors fmt strconv` |
| `go doc -all ./letter` | все идентификаторы документированы |

- В change-summary v1 (строка 5 таблицы «Как проверить», «Известные ограничения») записаны 3 замечания QF1010 в тестах. Они относились к прежней редакции тестов и устранены автором тестов (Q-CD1). Для текущих редакций линтер чист, К-15 выполнен. Это не замечание к реализации.

## Блокирующие замечания
Нет.

## Неблокирующие замечания
- **Н-1. `letter/doc.go`, строки 37–40.** Фраза «All methods accept invalid values and return results that no valid letter can produce» вместе с «false for [Letter.IsValid] and [Letter.Equal]».
  - Суть: false от `Equal` выдают и допустимые буквы (`C.Equal(D)`). Утверждение верно для −1, нулевого `Letter`, `nil` и `"Letter(n)"`, но не для `Equal`.
  - Почему проблема: документация неточна. architecture v5 в «Поведении при недопустимом значении» перечисляет «не-буквенные» результаты только для −1, нулевого `Letter` и `nil`.
  - Предложение: отделить `Equal` и `IsValid`. Например: «return -1 for numbers, the zero Letter for letters, nil for slices, and "Letter(n)" for String — results that no valid letter can produce; IsValid and Equal report false».
  - Адресат: go-code-developer.
- **Н-2. `letter/letter.go`, строка 82.** `return Letter(ordinal + 1), nil`.
  - Суть: здесь используется числовое представление (C = 1), а в `Shift` то же преобразование записано как `C + Letter(...)`.
  - Почему проблема: только читаемость и единообразие. Поведение верное.
  - Предложение: `C + Letter(ordinal)`.
  - Адресат: go-code-developer.

## Замечания к тестам
- Нет. Тесты согласованы с реализацией и контрактом. Поручение `review-tests-from-code` не требуется.
- Влияние на продолжение: переход к QA возможен после принятия текущих редакций тестов (example_test.go f865f94, letter_test.go 59b2746) на review-tests, запуск 3. Это ревью идёт параллельно.

## На будущее
- **Б-1. architecture, раздел «Риски».** У потребителей musex, печатающих `[]letter.Letter` через `fmt.Println` или `fmt.Sprint` без глагола, staticcheck выдаст ложное QF1010 из-за базового типа `uint8`. Обход — явный `%v`. Решать, упоминать ли это, архитектору. Адресат: software-architect, через оркестратора. Задачу не блокирует.
