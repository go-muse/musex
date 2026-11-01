- Статус: черновик
- Версия: 1
- Запуск: 1
- Входы: software-architect/step-7/architecture v1 (черновик, открыт Q-SA1), task-brief v2 (принят), business-analyst/step-5/domain-model v2, business-analyst/step-5/requirements v3, business-analyst/step-1/readiness v1, business-analyst/questions.md (Q-1…Q-8, Q-A1…Q-A6), software-architect/questions.md (Q-SA1 — открыт), docs/rules/main.md, docs/rules/dev.md, docs/rules/dev-test.md, .golangci.yml, .golangci-lint-version (v2.15.0), .github/workflows/ci.yml, CONTRIBUTING.md, go.mod (go 1.27.3), README.md, ROADMAP.md

# План реализации: пакет `letter` (шаг 1 roadmap)

Статус «черновик»: зависит от ответа на Q-SA1. При ответе (а) план переходит в «готов» без изменений; при (б) — новая версия вслед за architecture.

Порядок процесса: review-design (architecture + этот план) → точки человека → полный набор тестов (plan-1…plan-3) → review-tests → реализация (plan-4) → документация README (plan-5) → review-code → QA → приёмка человеком → plan-6 (ROADMAP.md). Реализация не начинается, пока весь набор тестов plan-1…plan-3 не принят на review-tests.

## Исполнители и разрешённые пути
| Участник | Шаги | Меняет | Не читает и не меняет |
|---|---|---|---|
| go-test-developer | plan-1, plan-2, plan-3 | `letter/*_test.go` (пакет `letter_test`), при необходимости `letter/testdata/`; `go.mod`, `go.sum` (Testify) | `letter/*.go` без `_test`, README.md, ROADMAP.md; их историю и диффы |
| go-code-developer | plan-4, plan-5, plan-6 | `letter/*.go` без `_test` (`letter/doc.go`, `letter/letter.go`, при желании ещё файлы без `_test`); `go.mod`, `go.sum` (только проверка `go mod tidy`, Testify сохранить); README.md (plan-5); ROADMAP.md (только plan-6) | `letter/*_test.go`, `letter/testdata/`, их историю и диффы |

- Общие входы обоих: task-brief v2, domain-model v2, requirements v3, architecture v1 (после перевода в «готов»), этот план, docs/rules/, docs/domain-research.
- Не меняются никем: `.golangci.yml`, `.github/`, `scripts/`, `CONTRIBUTING.md`, `docs/`.
- Оба запускают сборку, тесты, `go vet`, golangci-lint и команды Go для зависимостей самостоятельно.

## Подготовка окружения до разработки
- **П-0. Ветка.** До первого изменяющего шага создать локальную ветку `feat/1-add-letter-package` от актуального `main` (docs/rules/dev.md «Именование веток»; номер 1 — из roadmap). Создаёт участник первого изменяющего шага (go-test-developer в начале plan-1) в пределах права агентов создавать локальные ветки (docs/rules/main.md «Git и CI») либо человек — по решению оркестратора. Коммиты и push — только человек. Проверка: `git branch --show-current` → `feat/1-add-letter-package`.
- **П-0a. Инструменты.** `go version` → go1.27.3; `golangci-lint version` → v2.15.0 (CONTRIBUTING.md: можно в `.cache/bin/`). Недоступность — сообщить оркестратору; установка инструментов этим планом не разрешена.
- Отдельного шага подготовки зависимостей нет: Testify добавляет автор тестов в plan-1; go.mod уже есть. DevOps-шагов нет.

## Зависимости
- `github.com/stretchr/testify` **v1.11.1** — только тесты (`require`, `assert`). Добавляет go-test-developer в plan-1: `go get github.com/stretchr/testify@v1.11.1`. Транзитивные (`davecgh/go-spew`, `pmezard/go-difflib`, `gopkg.in/yaml.v3`) — как разрешит Go. Если версия недоступна в источнике — вопрос архитектору, другую не выбирать.
- До реализации `go mod tidy` может завершиться ошибкой только из-за отсутствия продуктовых файлов пакета `letter`: тогда `go mod tidy -e`, результат и диагностика — в change-summary. В plan-4 go-code-developer выполняет `go mod tidy` без `-e`; изменений go.mod/go.sum быть не должно, кроме приведения к аккуратному виду; Testify сохраняется.
- Продуктовый код — только стандартная библиотека.
- `go mod vendor` запрещён.

## Шаги
| Шаг | Исполнитель | Содержание | Файлы | Зависит от | Параллельно |
|---|---|---|---|---|---|
| plan-1 | go-test-developer | Testify в go.mod/go.sum; тесты основного поведения К-1…К-6, К-8, К-9, К-14 | `go.mod`, `go.sum`, `letter/letter_test.go` | П-0, принятые architecture и план (review-design, точки человека, ответ Q-SA1) | нет |
| plan-2 | go-test-developer | Исполнимые примеры (Q-A5, К-16) | `letter/example_test.go` | plan-1 (go.mod с Testify не нужен примерам, но порядок изменений — по одному) | нет: общий исполнитель и рабочая копия |
| plan-3 | go-test-developer | Недопустимые значения, владение срезами, конкурентность, импорты: К-7, К-10, К-12, К-13 | `letter/invalid_test.go`, `letter/concurrency_test.go`, `letter/imports_test.go` (имена файлов — по усмотрению автора, суффикс `_test.go` обязателен) | plan-1 | нет |
| gate | code-reviewer | review-tests всего набора plan-1…plan-3 | — | plan-1…plan-3 | — |
| plan-4 | go-code-developer | Весь публичный API по architecture; doc-комментарии на английском; `doc.go` с формулировками НФ-1 п. 2 | `letter/doc.go`, `letter/letter.go` | принятый review-tests | нет |
| plan-5 | go-code-developer | README.md по НФ-6, К-16 | `README.md` | plan-4 (имена и поведение API) | нет |
| gate | code-reviewer, qa-engineer | review-code (plan-4, plan-5), затем QA всего результата | — | plan-4, plan-5 | — |
| plan-6 | go-code-developer | После приёмки человеком: ROADMAP.md — отметить `[x]` п. 1 и исправить вводную фразу «All packages are planned; none is implemented yet.» (она станет неверной) | `ROADMAP.md` | приёмка человеком | нет |

- Параллельных шагов нет: все изменяющие шаги одной рабочей копии, plan-1…plan-3 и plan-4…plan-6 — у одного исполнителя каждый, и тесты целиком предшествуют реализации. plan-1…plan-3 оркестратор может выдать одним поручением, назвав части (task-state).

## Как проверить каждый шаг

### Подход к тестам (общий для plan-1…plan-3)
- **Внешний тестовый пакет** `letter_test`; только публичный API из architecture. Внутренний пакет `letter` в тестах не используется.
- **Testify**: `require` — для предусловий и проверок, после провала которых продолжать бессмысленно; `assert` — для независимых сравнений в таблицах. Ошибки — `require.ErrorIs(err, letter.ErrInvalid)`; текст ошибки не проверяется (не входит в контракт).
- **Табличные тесты** с именованными случаями (`name`) для однотипных проверок одной операции (docs/rules/dev-test.md); `t.Run` на случай.
- **Изоляция и параллельность**: `t.Parallel()` в каждом тесте верхнего уровня и в каждом подтесте; общих изменяемых данных между тестами нет; таблицы — локальные переменные теста; в Go 1.27 переменная цикла своя на итерацию — копировать не нужно. Время, случайность, файлы, сеть не нужны.
- **Оракул, независимый от реализации**: локальный в тестах литерал из семи констант в каноническом порядке `C D E F G A B`; порядковое место = индекс в литерале; ожидаемое вычисляется формулами требований (П-3, П-5) с математическим остатком, например `((x % 7) + 7) % 7`. Перечень пакета (`All`) как оракул для других операций не использовать.
- **Полный перебор**: 7 букв; 49 пар; сдвиги n ∈ [−15, 15] для каждой буквы; все 256 значений `uint8` для недопустимых значений.
- **Ожидаемые падения до реализации** (исключение «тесты до реализации» стека Go): `go test ./letter/...` и `go vet ./letter/...` не собираются с диагностикой об отсутствии не-тестовых Go-файлов в `github.com/go-muse/musex/letter` (или отсутствии символов API). Других пакетов в модуле нет. golangci-lint для пакета невыполним по той же причине — указать. Ошибки синтаксиса, неверные импорты, ошибки зависимостей — недопустимы: проверить `gofmt -l` на своих файлах (пустой вывод) и что единственная причина сборочных ошибок — отсутствующий пакет или его символы.

### plan-1 — основное поведение
Проверяемые случаи (все из requirements v3):
- **К-1**: `All()` равен литералу оракула; длина = `letter.Count` = 7; все различны. `Count` — константа, пригодная как длина массива (объявление массива `[letter.Count]…` компилируется).
- **К-2**: `Cycle` от C, A, B — фиксированные последовательности; для каждой из 7 начальных: длина 7, все различны, элемент i = `start.Shift(i)` и = оракул[(ord(start) + i) mod 7]; `C.Cycle()` равен `All()`.
- **К-3**: таблица примеров (C→E 2, B→D 2, E→F 1, G→C 3, B→C 1, C→B 6, C→C 0); 49 пар: 0…6 и формула.
- **К-4**: таблица (C→E 5, E→C 2, D→B 2, C→B 1, C→C 0); 49 пар: `DistanceDown(a, b) == DistanceUp(b, a)`, сумма 7 при a ≠ b, оба 0 при a = b.
- **К-5**: таблица (C+0, C+1, B+1, C+9, G+4, C+7, C+70, C−1, C−7, C−8, E−2); 7 × n ∈ [−15, 15]: формула и возврат `Shift(n).Shift(-n)`; 49 пар: `a.Shift(a.DistanceUp(b)) == b`, `a.Shift(-a.DistanceDown(b)) == b`.
- **К-6**: `C.Shift(9) == E`, `B.Shift(1) == C`, `A.Cycle()[2] == All()[0]`; 7 букв попарно различны; буквы, полученные разными путями, — один ключ `map`.
- **К-8**: `Shift(math.MaxInt)`, `Shift(math.MinInt)`, `Shift(math.MaxInt - k)`, `Shift(math.MinInt + k)` для малых k у всех 7 букв; ожидаемое — формулой без переполнения (остаток считать до сложения); дополнительно при `strconv.IntSize == 64`: `C + MaxInt = C`, `C + MinInt = B`; при 32: D и A. Возврат `Shift(n).Shift(-n)` не проверять для `n = math.MinInt` (в Go `-math.MinInt == math.MinInt`).
- **К-9**: входит в таблицы К-3, К-4.
- **К-14**: перебор К-3…К-5 под `require.NotPanics`; операции ошибок не возвращают (по сигнатурам).
- Q-3: `Ordinal` 0…6 по оракулу; `FromOrdinal(i)` для 0…6 — буква оракула и `nil`; `FromOrdinal(l.Ordinal()) == l`; `MustFromOrdinal(i)` для 0…6 без паники. `String` для 7 букв — `"C"…"B"`.
- Проверка шага: change-summary перечисляет требования и случаи; `gofmt -l letter/letter_test.go` пусто; `go.mod` содержит `github.com/stretchr/testify v1.11.1`; `go.sum` заполнен; ожидаемая диагностика отсутствующего пакета записана.

### plan-2 — исполнимые примеры
- Функции и содержание — таблица «Исполняемые примеры Go» в architecture: обязательны `Example`, `ExampleAll`, `ExampleLetter_Cycle`, `ExampleLetter_DistanceUp`, `ExampleLetter_DistanceDown`, `ExampleLetter_Shift`, `ExampleFromOrdinal`, `ExampleLetter_IsValid`. Каждый с `// Output:`; значения — из К-2…К-5. Testify в примерах не используется (вывод проверяет `go test`).
- Проверка шага: имена совпадают с таблицей (их проверит `go vet` после реализации: анализатор tests сверяет имена примеров с идентификаторами); `gofmt -l` пусто.

### plan-3 — недопустимые значения, владение, конкурентность, импорты
- **К-7** (таблица «Поведение при недопустимом значении» в architecture): по всем 256 значениям `letter.Letter(v)` ровно 7 допустимы (`IsValid`), и это семь констант. Для каждого недопустимого v, включая нулевое `var l letter.Letter`: `IsValid` false; `Ordinal` −1; `DistanceUp`/`DistanceDown` −1, когда недопустим получатель и когда недопустим аргумент (с каждой из 7 букв), и когда оба; `Shift(n)` для n ∈ {−8, −1, 0, 1, 7, math.MinInt, math.MaxInt} — результат недопустим и равен нулевому `Letter`; `Cycle()` — длина 0 (nil); `String()` = `fmt.Sprintf("Letter(%d)", v)` и не равна ни одному из "C"…"B"; `!=` каждой из семи констант. `FromOrdinal` для {−1, 7, 8, 100, math.MinInt, math.MaxInt}: `ErrorIs(ErrInvalid)`, возвращённая буква недопустима; `MustFromOrdinal` для тех же — паника, восстановленное значение — `error`, `errors.Is(…, ErrInvalid)` (через `recover` в помощнике теста; `assert.PanicsWithError` не подходит — он сравнивает текст ошибки, а текст не входит в контракт).
- **К-10**: изменить все элементы среза `All()` и среза `Cycle()` (для каждой начальной буквы), затем повторный вызов возвращает прежний результат; дополнительно `append` к полученному срезу не влияет на следующий вызов.
- **К-13**: N горутин (`sync.WaitGroup.Go`, N порядка `runtime.GOMAXPROCS(0) * 4`, без `time.Sleep`) одновременно вызывают все операции для всех букв и пар, сверяют результаты с оракулом и меняют свои полученные срезы; результаты собираются через канал или срез по индексу горутины, проверка `assert` — после `wg.Wait()` в горутине теста (вызов `require` из рабочих горутин запрещён). Гонки ловит `go test -race`; тест обязан выполняться под `-race` (так в CI).
- **К-12**: тест со стандартным `go/build` (`build.ImportDir(".", 0)` из каталога пакета): список не-тестовых импортов пакета не содержит путей с префиксом `github.com/go-muse/musex` и путей сторонних модулей (только стандартная библиотека — признак: в первом элементе пути нет точки). Это автоматическая проверка списка импортов, а не чтение реализации автором тестов. Дублируется командой QA (ниже).
- Проверка шага: change-summary со связью «критерий → тест»; `gofmt -l` пусто; ожидаемая диагностика.

### review-tests (code-reviewer)
- Полнота: каждый критерий К-1…К-10, К-12…К-14 и Q-3 покрыт; примеры — по таблице architecture (К-16 в части примеров); К-11, К-15, К-16 (README, ROADMAP) — не тестами, а ревью и QA.
- Testify применён правильно; `t.Parallel()`; табличный стиль; оракул независим от реализации; нет доступа к продуктовой области.

### plan-4 — реализация
- Весь API из architecture: `Letter`, `C…B`, `Count`, `ErrInvalid`, `All`, `FromOrdinal`, `MustFromOrdinal`, методы `IsValid`, `Ordinal`, `DistanceUp`, `DistanceDown`, `Shift`, `Cycle`, `String`; получатели по значению (при ответе (а) на Q-SA1). Ограничение на `Shift`: приведение `steps` по модулю `Count` до сложения (К-8). `MustFromOrdinal` — обёртка над `FromOrdinal`. Комментарии в коде — американский английский. Пакетный doc — в `letter/doc.go`, с формулировками НФ-1 п. 2.
- Проверка (автор сам, результаты в change-summary):
  1. `gofmt -l letter/doc.go letter/letter.go` — пусто; `golangci-lint fmt ./...` не меняет файлы;
  2. `go build ./...`;
  3. `go vet ./...` — без замечаний;
  4. `go test -race -count=1 ./...` — все тесты и примеры проходят;
  5. `golangci-lint run ./...` — без замечаний (`.golangci.yml`);
  6. `go mod tidy` — go.mod/go.sum без лишних изменений, Testify на месте;
  7. `go doc -all ./letter` — каждый экспортируемый идентификатор документирован.
- Не правит тесты; несогласие с тестом или пробел — вопрос автору тестов через questions.md по порядку «Запрос расширения или исправления тестов».

### plan-5 — README.md
- Разделы Features, Concept, Examples описывают только реализованное в `letter` и помечены unreleased; Examples — в формате «вопрос — краткое объяснение — ссылка» на примеры по именам из architecture (например, `https://pkg.go.dev/github.com/go-muse/musex/letter#example-Letter.Shift`) или на документацию API.
- Баннер «Early development»: убрать утверждение «Musical functionality is not implemented yet», сохранить предупреждение о ранней стадии и неготовности к промышленному использованию (Д-5 требований).
- Use Cases и значок Roadmap не менять (Д-6, Д-7 требований). Комментарии-правила README сохранить.
- Проверка: дифф README.md затрагивает только названные места; ссылки ведут на существующие имена примеров и идентификаторы; язык — английский.

### review-code и QA (сведения для проверяющих)
- review-code: соответствие контракту architecture, согласованность с принятыми тестами, README по К-16.
- QA (полный объём, К-1…К-16): `golangci-lint config verify`; `golangci-lint run ./...`; `go vet ./...`; `go test -race -count=1 -covermode=atomic -coverprofile=coverage.out ./...`; `go tool cover -func=coverage.out` (порога нет); `govulncheck ./...` (v1.9.0, как в CI; при недоступности инструмента — указать ограничение); `go list -f '{{join .Imports "\n"}}' ./letter` — только стандартная библиотека (К-12); `go doc -all ./letter` — документация и формулировки НФ-1 п. 2 (К-11 б), отсутствие способа создать иной набор (К-11 д); README по К-16; `git diff main -- ROADMAP.md` пуст (К-16). Редакции файлов фиксировать хешами (`git hash-object`).
- К-11 а), в), г) проверяет code-reviewer на review-design по architecture (раздел «НФ-1»).

### plan-6 — ROADMAP.md после приёмки
- Только после явной приёмки человеком; отдельное поручение оркестратора.
- Изменения: `- [ ] **1. \`letter\`**` → `- [x] …`; вводная фраза «All packages are **planned**; none is implemented yet.» заменяется на верную (например, что пакеты отмечаются по мере реализации), без изменения остальных пунктов.
- Проверка: `git diff -- ROADMAP.md` содержит только эти две правки; проверяет оркестратор, коммит делает человек.
