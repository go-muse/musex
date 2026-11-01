- Статус: готов
- Версия: 1
- Запуск: 2
- Входы: task-brief v2, business-analyst/step-5/domain-model v2, business-analyst/step-5/requirements v3, software-architect/step-7/architecture v4, software-architect/step-8/implementation-plan v4, software-architect/questions.md (Q-SA1: «(а) uint8 + Equal»), business-analyst/questions.md (Q-8 (а)), orchestrator/questions.md (О-7: 32-битные платформы не поддерживаются; О-10: golangci-lint v2.15.0 в `./.cache/bin`; О-11: только текущая архитектура; О-12: правила команды — docs/rules/ и .golangci.yml, CONTRIBUTING.md не нормативен), docs/rules/main.md, docs/rules/dev.md, docs/rules/dev-test.md, .golangci.yml, .github/workflows/ci.yml

# Описание изменений: полный набор тестов пакета `letter` (plan-1, plan-2, plan-3)

## Что сделано
- Testify v1.11.1 добавлен в go.mod/go.sum (`go get github.com/stretchr/testify@v1.11.1`, затем `go mod tidy`).
- Полный набор тестов во внешнем пакете `letter_test`, только по публичному API из architecture v4 (запуск 2 подтвердил соответствие; в запуске 1 — по v3). Реализацию не читал и не создавал (её нет).
- Оракул не зависит от реализации: локальный массив `C D E F G A B` (возвращается по значению, общего состояния нет); место буквы — индекс в нём; ожидаемое — формулы П-3, П-5 с остатком `((x % 7) + 7) % 7`. Остаток сдвига берётся до сложения, поэтому переполнения нет.
- `t.Parallel()` в каждом тесте и подтесте; именованные табличные случаи; `require` — для предусловий, `assert` — для независимых сравнений; ошибки проверяются через `ErrorIs`, текст ошибок не проверяется.
- Исполняемые примеры: все обязательные из таблицы architecture и все три необязательных; в `// Output:` только `String`, числа и `bool`.

### Критерий → тест
| Критерий | Тесты (файл) |
|---|---|
| К-1 | `TestAll` (порядок, длина = ёмкость = `Count`, каждая буква один раз), `TestCount` (нетипизированная константа: длина массива, `int64`, `uint8`; `== len(All())`) — letter_test.go |
| К-2 | `TestLetter_Cycle_Fixed` (от C, A, B), `TestLetter_Cycle_AllStarts` (7 стартов: длина 7, все различны, элемент i = оракул и = `Shift(i)`; `C.Cycle() == All()`) — letter_test.go |
| К-3, К-9 | `TestLetter_DistanceUp_Examples` (7 примеров, включая B→C = 1 и C→C = 0), `TestLetter_DistanceUp_AllPairs` (49 пар: 0…6 и формула) — letter_test.go |
| К-4, К-9 | `TestLetter_DistanceDown_Examples` (5 примеров, включая C→B = 1), `TestLetter_DistanceDown_AllPairs` (49 пар: формула, `down(a,b) == up(b,a)`, сумма 7 при a ≠ b, оба 0 при a = b) — letter_test.go |
| К-5 | `TestLetter_Shift_Examples` (11 примеров), `TestLetter_Shift_SmallRange` (7 × n ∈ [−15, 15]: формула, возврат, период 7), `TestLetter_Shift_LargeSteps` (±10^9, ±(10^9 + 3), ±2^31, ±(2^31 − 1), ±2^40, ±7·2^40, ±(2^62 + 5), кратные 7 у краёв `int`), `TestLetter_Shift_ByDistance` (49 пар: `a.Shift(a.DistanceUp(b)) == b`, `a.Shift(-a.DistanceDown(b)) == b`) — letter_test.go |
| К-6 | `TestLetter_Equal_Examples` (C+9 = E, B+1 = C, `A.Cycle()[2]` = `All()[0]`, плюс отрицательный случай; каждая проверка через `Equal` и `==`), `TestLetter_Equal_AllPairs` (49 пар: истинно только для одной буквы, совпадает с `==`, симметрия), `TestLetter_MapKey` (пять путей получения C — один ключ map) — letter_test.go |
| К-7 | `TestLetter_IsValid_AllValues` (из 256 значений `uint8` допустимы ровно 7 констант); `TestInvalidValues` (249 недопустимых значений, включая 0: `IsValid` false; `Ordinal` −1; `Cycle` nil; `String` = `Letter(<v>)` и не совпадает с "C"…"B"; `Equal` false при недопустимом получателе, аргументе, обоих, с самим собой; `DistanceUp`/`DistanceDown` −1 в тех же сочетаниях; `!=` каждой из семи; `Shift(n)` для n ∈ {−8, −1, 0, 1, 7, MinInt, MaxInt} = нулевое `Letter`); `TestZeroValue`; `TestArithmeticIsNotShift` (`B + 1` недопустимо, Р-1); `TestFromOrdinal_Invalid` и `TestMustFromOrdinal_Invalid` для {−1, 7, 8, 100, MinInt, MaxInt}: `ErrorIs(ErrInvalid)`, возвращено нулевое `Letter`; паника, значение — `error`, `ErrorIs(ErrInvalid)` (через `recover` в помощнике; `PanicsWithError` не используется, так как сравнивает текст) — invalid_test.go |
| К-8 | `TestLetter_Shift_IntLimits` (все 7 букв: `MaxInt − k`, `MinInt + k`, k = 0…13; возврат, кроме n = MinInt; C + MaxInt = C, C + MinInt = B), `TestLetter_Shift_LargeSteps` — letter_test.go |
| К-10 | `TestAll_Ownership` (перезапись всех элементов; срез, полученный раньше другим вызывающим; `append` в свободную ёмкость префикса), `TestLetter_Cycle_Ownership` (то же для каждой начальной буквы) — invalid_test.go |
| К-12 | `TestPackageImports` (`go/build.ImportDir`: у пакета есть не-тестовые файлы; импорты без префикса `github.com/go-muse/musex` и без точки в первом элементе пути) — imports_test.go |
| К-13 | `TestConcurrentUse` (`GOMAXPROCS × 4` горутин через `wg.Go`, старт по закрытию канала, по 20 итераций: все операции, включая `Equal`, `FromOrdinal`, `MustFromOrdinal`, для всех букв и пар; сверка с оракулом; изменение полученных срезов; ошибки — в слот своей горутины, `assert` — после `wg.Wait()` в горутине теста). Гонки ловит только запуск с `-race` (так в CI) — concurrency_test.go |
| К-14 | `TestValidLetters_NoPanics` (для каждой буквы все операции, перебор К-3…К-5 и пределы `int` под `require.NotPanics`, `FromOrdinal` без ошибки) — letter_test.go |
| Q-3 | `TestLetter_Ordinal` (0…6, `== C.DistanceUp(l)`), `TestFromOrdinal_Valid` (буква оракула и nil; `FromOrdinal(l.Ordinal()) == l`), `TestMustFromOrdinal_Valid` (без паники, буква оракула) — letter_test.go |
| Q-4 | `TestLetter_String` ("C"…"B", `fmt.Sprint`, `fmt.Sprint(All()) == "[C D E F G A B]"`) — letter_test.go |
| Стабильные числа | `TestConstantValues` (`uint8(C) == 1` … `uint8(B) == 7`) — letter_test.go |
| К-16 (примеры) | example_test.go: `Example`, `ExampleAll`, `ExampleLetter_Cycle`, `ExampleLetter_DistanceUp`, `ExampleLetter_DistanceDown`, `ExampleLetter_Shift`, `ExampleFromOrdinal`, `ExampleLetter_IsValid`, `ExampleLetter_Equal`; по усмотрению — `ExampleMustFromOrdinal`, `ExampleLetter_Ordinal`, `ExampleLetter_String`. Значения — из К-2…К-5 |
| К-11, К-15, остальное К-16 | не тестами: review-design, review-code, QA (по плану) |

## Изменённые компоненты и файлы
| Файл | Содержание | `git hash-object` |
|---|---|---|
| `go.mod` | `require github.com/stretchr/testify v1.11.1`; косвенные `davecgh/go-spew v1.1.1`, `pmezard/go-difflib v1.0.0`, `gopkg.in/yaml.v3 v3.0.1` | db78bedc4fddabb64e223f8d56adfccc65a6d79b |
| `go.sum` (новый) | контрольные суммы Testify и транзитивных | c4c1710c475c1aaf6b41e75cf7003b167ddb3739 |
| `letter/helpers_test.go` | оракул, `mod7`, `expectedShift`, `recoverPanic` | 920fdf54837bc6a5edabf958bacd50138da6c3fe |
| `letter/letter_test.go` | plan-1 | 9989d3f0f9515d45cb3b8b4c34caa42ffc6326ef |
| `letter/example_test.go` | plan-2 | ba5f3959cb6e5c38f52918c5b15ca70d31e38eff |
| `letter/invalid_test.go` | plan-3: К-7, К-10 | ef31914964848d611ea3bd00f33a10668649b97e |
| `letter/concurrency_test.go` | plan-3: К-13 | bdb3281ff817e97aba87ee5317b6efc0fa80ecee |
| `letter/imports_test.go` | plan-3: К-12 | c5efe532624d7fc0e7cc48a04024a42d7ac6fe8e |

Продуктовые файлы, README.md, ROADMAP.md, docs/, .golangci.yml, .github/, scripts/ не менялись. `letter/testdata/` не понадобился. Коммитов нет.

## Как проверить
- `./.cache/bin/golangci-lint fmt --diff letter/*_test.go` — пустой вывод.
- После реализации (plan-4): `go vet ./...`, `go test -race -count=1 ./...`, `./.cache/bin/golangci-lint run ./...` — все тесты и примеры проходят.
- Точечно: `go test -race -count=1 -run '^TestConcurrentUse$' ./letter`.

## Исходное состояние тестов
Окружение: darwin/arm64, go1.27.3, golangci-lint 2.13.2. Ветка `feat/1-add-letter-package`. Go-пакетов не было:
- `go build ./...` — `warning: "./..." matched no packages`, exit 0;
- `go test ./...` — `no packages to test`, exit 1;
- `go vet ./...` — `no packages to vet`, exit 1.

## Выполненные команды и результаты (тесты до реализации)

### Запуск 2: перепроверка входов
- Входы: architecture v4 и implementation-plan v4 отличаются от v3 только удалением 32-битных положений (О-7), положением о 64-битных платформах и ссылками на версию 4. О-11 и О-12 требований к тестам не меняют: порядок импортов — gci из .golangci.yml (стандартная → `github.com/go-muse/musex` → прочие), тот же, что применён. Итог — **подтверждено**: тесты не менялись, хеши файлов прежние (таблица выше).
- Окружение: darwin/arm64, go1.27.3, golangci-lint 2.15.0 (`./.cache/bin/golangci-lint`, О-10).

| Команда | Результат |
|---|---|
| `./.cache/bin/golangci-lint version` | 2.15.0, built with go1.27.3 |
| `gofmt -l letter/*_test.go` | пусто, exit 0 |
| `./.cache/bin/golangci-lint fmt --diff letter/*_test.go` | пусто, exit 0 |
| `./.cache/bin/golangci-lint config verify` | exit 0 |
| `./.cache/bin/golangci-lint run ./...` | **ожидаемое падение**: `typechecking error: no non-test Go files in …/musex/letter`, exit 7 |
| `go vet ./...` | **ожидаемое падение**: `github.com/go-muse/musex/letter: no non-test Go files in …/musex/letter`, exit 1 |
| `go test -race -count=1 ./...` | **ожидаемое падение**: та же диагностика, `[build failed]`, exit 1 |
| `go mod tidy` | exit 0, go.mod/go.sum без изменений |

### Запуск 1
| Команда | Результат |
|---|---|
| `go get github.com/stretchr/testify@v1.11.1` | `go: added github.com/stretchr/testify v1.11.1`, exit 0 |
| `go mod tidy` (после появления импортов) | exit 0, без `-e`: модуль собрал граф зависимостей тестов; повторный запуск изменений не дал |
| `gofmt -l letter/*_test.go` | пусто, exit 0 |
| `golangci-lint fmt --diff letter/*_test.go` (2.13.2) | пусто, exit 0 (gofmt + gci по .golangci.yml: стандартная → musex → сторонние) |
| `golangci-lint config verify` | exit 0 |
| `go list -e -f … ./letter` | `GoFiles=[]`, 6 файлов `XTestGoFiles`, ошибок разбора нет |
| `go build ./...` | exit 0, без вывода (не-тестовых файлов нет, собирать нечего) |
| `go vet ./...` | **ожидаемое падение**: `github.com/go-muse/musex/letter: no non-test Go files in …/musex/letter`, exit 1 |
| `go test -race -count=1 ./...` | **ожидаемое падение**: та же диагностика, `FAIL github.com/go-muse/musex/letter [build failed]`, exit 1 |
| `go test -count=1 -run '^TestPackageImports$' ./letter` | то же: `-run` не обходит компиляцию |
| `golangci-lint run ./...` | **ожидаемое падение**: `typechecking error: no non-test Go files in …/musex/letter`, exit 7 |

### Ожидаемые падения (исключение «тесты до реализации» стека Go)
- Причина — нет продуктовых Go-файлов пакета `github.com/go-muse/musex/letter`. Тесты используют только символы из «Сводки» architecture v4: `Letter`, `C D E F G A B`, `Count`, `ErrInvalid`, `All`, `FromOrdinal`, `MustFromOrdinal`, `IsValid`, `Equal`, `Ordinal`, `DistanceUp`, `DistanceDown`, `Shift`, `Cycle`, `String`.
- Пакет целиком не собирается, поэтому ни один тест и пример не выполнен. Ни одно поведение не проверено. Других пакетов в модуле нет.
- Других причин нет. Синтаксис проверен (`gofmt`, `go list` разбирает все 6 файлов), форматтеры и импорты чистые, зависимости разрешаются (`go mod tidy` без `-e`).

### Невыполненные проверки
- Проверка типов тестового кода (`go vet`), выполнение тестов и примеров, детектор гонок, `golangci-lint run` — невыполнимы до plan-4. Ошибки типов в тестах, если они есть, выявятся только при первой сборке с реализацией. Заглушки API для проверки типов не создавались: это запрещено поручением.
- Проверка имён примеров анализатором `tests` (`go vet`) — после реализации.
- Для К-13 гонки ловятся только при запуске с `-race`. Сам тест этот флаг не проверяет.

## Изменения существующих тестов
Нет: тестов раньше не было.

## Отклонения от плана
- **32-битные платформы (решение человека О-7, orchestrator/questions.md, 2026-11-01).** Тесты написаны только для 64-битного `int`: большие сдвиги — константы `int`, проверяются только C + MaxInt = C и C + MinInt = B. В запуске 1 это расходилось с architecture v3 и планом v3. Запуск 2: **расхождение снято** — architecture v4 и implementation-plan v4 требуют того же (О-7, О-11).
- **Версия golangci-lint.** В запуске 1 проверки форматтеров выполнялись системным 2.13.2. Запуск 2: **ограничение снято** — все проверки повторены на v2.15.0 (`./.cache/bin/golangci-lint`, О-10), результаты прежние.
- Состав файлов: plan-3 разделён на `invalid_test.go` (К-7, К-10), `concurrency_test.go`, `imports_test.go`; общий оракул вынесен в `helpers_test.go`. План разрешает выбирать имена файлов.
- Работа вне репозитория (запуск 1): чтобы убедиться, что `golangci-lint fmt --diff` с `.golangci.yml` действительно проверяет порядок импортов, я один раз запустил его на синтетическом файле (неверный порядок трёх групп, без кода проекта). Файл лежал во временном каталоге сессии вне musex и удалён сразу после проверки. Это выходило за границу docs/rules/main.md «Границы файловой работы». В запуске 2 файлов вне musex и временных файлов не создавалось.

## Учтённые отчёты
Нет (первый запуск).

## Известные ограничения
- До реализации тесты не выполнялись (см. «Невыполненные проверки»).
- `TestPackageImports` читает каталог пакета через `go/build` по текущему GOOS/GOARCH. Файлы, исключённые build-ограничениями, не учитываются. QA дублирует проверку командой `go list` (план).
- В `TestLetter_Cycle_*` ёмкость среза `Cycle` не проверяется: architecture задаёт для `Cycle` новый срез «как у `All`», без явной ёмкости. Для `All` ёмкость `Count` проверяется, как в контракте.

## Допущения
- Для `Cycle` недопустимой буквы проверяется именно `nil` (`assert.Nil`) и длина 0: architecture пишет «`nil` (длина 0)».
- `MustFromOrdinal` проверяется на допустимость и значение паники через `errors.Is`. Равенство ошибки паники ошибке `FromOrdinal` не проверяется: текст не входит в контракт, тождества значения ошибки контракт не задаёт.

## Затронутая документация
Нет. README и doc-комментарии — plan-4 и plan-5 (go-code-developer). Имена примеров совпадают с таблицей architecture, ссылки README на них будут верны.

## Вопросы
Нет открытых. tasks/2026-11-01-letter/go-test-developer/questions.md не создавался.
