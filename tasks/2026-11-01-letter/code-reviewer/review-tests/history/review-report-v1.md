- Статус: готов
- Версия: 1
- Запуск: 1
- Входы: go.mod db78bed, go.sum c4c1710, letter/concurrency_test.go bdb3281, letter/example_test.go ba5f395, letter/helpers_test.go 920fdf5, letter/imports_test.go c5efe53, letter/invalid_test.go ef31914, letter/letter_test.go 9989d3f (git hash-object, совпадают с поручением и change-summary), go-test-developer/tests/change-summary v1 (запуск 2), business-analyst/step-5/requirements v3, business-analyst/step-5/domain-model v2, software-architect/step-7/architecture v4, software-architect/step-8/implementation-plan v4, решения человека Q-SA1, Q-8 (а), О-7, О-11, О-12, docs/rules/dev-test.md, .golangci.yml

# Отчёт ревью: review-tests, полный набор тестов пакета `letter` (plan-1…plan-3)

Предмет: тесты до реализации. Реализации нет; проверено по requirements v3 и контракту architecture v4, независимо от реализации. Применены скиллы golang-testing и golang-stretchr-testify.

## Вердикт
**Доработать.** Одно блокирующее замечание: ошибочное ожидание в `ExampleMustFromOrdinal`. Остальной набор полный и корректный.

## Блокирующие замечания

### Б-1. Неверный ожидаемый вывод `ExampleMustFromOrdinal`
- Место: `letter/example_test.go:82–87`, строка 86.
  ```go
  func ExampleMustFromOrdinal() {
  	fmt.Println(letter.MustFromOrdinal(2).Shift(3))
  	// Output:
  	// F
  }
  ```
- Суть: по контракту architecture v4 порядковое место 2 — это E («`Ordinal` … C = 0 … B = 6»; набор сам это проверяет в `ExampleFromOrdinal`: `FromOrdinal(2)` печатает `E`). `E.Shift(3)` = место (2 + 3) mod 7 = 5 = **A** (формула `Shift`, П-5, К-5). Ожидается `A`, а не `F`.
- Почему это проблема: с верной реализацией пример упадёт при `go test`, а К-15 и К-16 требуют, чтобы примеры проходили. Пример противоречит другому примеру того же файла. Чтобы пример прошёл, автору реализации пришлось бы нарушить контракт.
- Откуда ошибка: значение взято из комментария в architecture v4 (стр. 187: `letter.MustFromOrdinal(2).Shift(3) // F`). Этот комментарий противоречит самому контракту, см. Н-1.
- Предложение: исправить ожидаемый вывод на `A`. Можно и сменить цепочку так, чтобы получилась F: `letter.MustFromOrdinal(0).Shift(3)` или `letter.MustFromOrdinal(2).Shift(1)`. Контракт однозначен, поэтому ответ архитектора для исправления не нужен.
- Адресат: go-test-developer.

## Неблокирующие замечания

### Н-1. Противоречие в примерах architecture v4 (пробел спецификации)
- Место: `software-architect/step-7/architecture.md`: стр. 187 (`letter.MustFromOrdinal(2).Shift(3) // F`) и стр. 259 (`letter.MustFromOrdinal(2).Shift(3).Equal(letter.F)` как пример работающей цепочки).
- Суть: по разделам `Ordinal`, `FromOrdinal` и `Shift` того же документа результат — A, не F. На тесты это не влияет: контракт однозначен, исправление Б-1 укладывается в него.
- Почему это проблема: автор реализации в plan-5 (README, Examples) и QA работают по architecture и могут перенести ошибку в README или в свои проверки.
- Предложение: исправить оба примера (`// A` и `.Equal(letter.A)` или другую цепочку) до выдачи plan-4 и plan-5. Решает архитектор, порядок задаёт оркестратор.
- Адресат: software-architect (через оркестратора).

### Н-2. `TestValidLetters_NoPanics` отбрасывает результаты: возможен отказ staticcheck SA4017 на К-15
- Место: `letter/letter_test.go:556–575`: `_ = l.Cycle()`, `_ = l.Ordinal()`, `_ = l.IsValid()`, `_ = l.DistanceUp(other)`, `_ = l.DistanceDown(other)`, `_ = l.Equal(other)`, `_ = l.Shift(n)`, `_ = l.Shift(math.MaxInt)` и др.
- Суть: в `.golangci.yml` включён staticcheck. Его проверка SA4017 («function doesn't have side effects and its return value is ignored») срабатывает на вызов функции, которую анализ чистоты признал чистой, если результат отброшен, в том числе через `_ =`. Исключение есть только для функций с параметром `*testing.B`. Методы `Letter` принимают только значения базовых типов (`Letter` на `uint8`, `int`). Если они будут реализованы чистой арифметикой без чтения пакетных таблиц, а для `IsValid`, `Equal`, `Ordinal`, `DistanceUp` и `Shift` это вероятно, то `golangci-lint run ./...` может выдать замечания на эти строки. Проверить до реализации нельзя: пакет не собирается, заглушки запрещены. Поэтому это риск, а не установленный дефект.
- Почему это проблема: К-15 требует чистого golangci-lint по всему модулю, тесты входят в его объём. Замечание после реализации вызовет лишний цикл «тесты → review-tests → код». Кроме того, тест не проверяет полученные значения: К-14 по сути проверяется другими тестами.
- Предложение: использовать результаты, например накапливать их и проверять после `require.NotPanics`: число допустимых результатов `Shift`, сумму или диапазон расстояний, `len(Cycle())`. Так исчезнут отброшенные вызовы и усилится сам тест. Лучше сделать это в той же доработке, что и Б-1.
- Адресат: go-test-developer.

### Н-3. Громоздкая таблица `TestLetter_Equal_Examples`
- Место: `letter/letter_test.go:371–401`.
- Суть: поля `got`/`want` — функции от `*testing.T`. Это нужно только одному случаю с `elementAt`. Корректно: `require` внутри `elementAt` получает `t` подтеста. Но читается тяжелее, чем нужно.
- Предложение (по желанию): вычислить `letter.A.Cycle()` и `letter.All()` с `require.Len` до таблицы, а в таблицу положить готовые значения `letter.Letter`.
- Адресат: go-test-developer.

## Что проверено и принято
- **Редакции:** хеши всех восьми файлов совпадают с поручением и change-summary. Изменены только `go.mod` и `go.sum`, новые файлы — только `letter/*_test.go`. Продуктовых файлов в `letter/` нет. Доступа к продуктовой области нет: все файлы в пакете `letter_test`, используются только символы из «Сводки» architecture v4.
- **Сигнатуры против architecture v4** (компилятор их пока не проверит; сверено вручную): `All() []Letter`; `FromOrdinal(int) (Letter, error)`; `MustFromOrdinal(int) Letter`; `IsValid() bool`; `Equal(Letter) bool`; `Ordinal() int`; `DistanceUp(Letter) int`; `DistanceDown(Letter) int`; `Shift(int) Letter`; `Cycle() []Letter`; `String() string`; `Count` как нетипизированная константа (`[letter.Count]string`, `var asUint8 uint8 = letter.Count`, `var asInt64 int64 = letter.Count`); `ErrInvalid` через `ErrorIs`; `letter.Letter(v)` для `v` типа `int`; `uint8(l)`; `letter.B + 1` — типизированная константа `Letter(8)`. Расхождений нет.
- **Полнота:**
  - К-1: `TestAll`, `TestCount`.
  - К-2: `TestLetter_Cycle_Fixed`, `TestLetter_Cycle_AllStarts`.
  - К-3 и К-9: `TestLetter_DistanceUp_Examples` и `_AllPairs`.
  - К-4 и К-9: `TestLetter_DistanceDown_Examples` и `_AllPairs`.
  - К-5: `TestLetter_Shift_Examples`, `_SmallRange`, `_LargeSteps`, `_ByDistance`.
  - К-6: `TestLetter_Equal_Examples`, `_AllPairs`, `TestLetter_MapKey`.
  - К-7: `TestLetter_IsValid_AllValues`, `TestInvalidValues` (все 249 недопустимых значений, каждая операция из таблицы architecture, получатель и аргумент), `TestZeroValue`, `TestArithmeticIsNotShift`, `TestFromOrdinal_Invalid`, `TestMustFromOrdinal_Invalid`.
  - К-8: `TestLetter_Shift_IntLimits`, `_LargeSteps`.
  - К-10: `TestAll_Ownership`, `TestLetter_Cycle_Ownership`, включая `append` в свободную ёмкость.
  - К-12: `TestPackageImports`.
  - К-13: `TestConcurrentUse`.
  - К-14: `TestValidLetters_NoPanics`, с оговоркой Н-2.
  - Q-3: `TestLetter_Ordinal`, `TestFromOrdinal_Valid`, `TestMustFromOrdinal_Valid`.
  - Q-4: `TestLetter_String`.
  - Стабильные числа C = 1 … B = 7: `TestConstantValues`.
  - Примеры: все девять обязательных по таблице architecture и три необязательных. К-11, К-15 и К-16 (README, ROADMAP) — не тестами, по плану.
- **Ожидания пересчитаны:**
  - К-3: C→E 2, B→D 2, E→F 1, G→C 3, B→C 1, C→B 6.
  - К-4: C→E 5, E→C 2, D→B 2, C→B 1.
  - К-5: все 11 примеров.
  - К-8: 2^63 − 1 ≡ 0 и −2^63 ≡ 6 (mod 7), то есть C и B. `math.MaxInt − math.MaxInt%7` = MaxInt, `math.MinInt − math.MinInt%7` = MinInt + 1 (усечённый остаток −1): оба ≡ 0, сохраняют букву.
  - Выборка больших сдвигов — константы `int`, на 64-битном `int` переполнения нет (О-7).
  - Выводы примеров, кроме Б-1, верны: E/2; `[C D E F G A B]`/`7 7`; `[A B C D E F G]`; 2/2/6; 5/1; E/B/C; `E true`/`true false`; false/true/false; true/false/false; `0 6`/`sol`; `A`/`Letter(0)`.
- **Оракул** независим от реализации: литерал констант в каноническом порядке, место — индекс, `mod7` без переполнения на `MinInt`, в `expectedShift` остаток берётся до сложения. `All()` не используется как оракул для других операций. Связи `Shift(i)` и `C.DistanceUp(l)` проверяются как свойства из контракта дополнительно к оракулу.
- **Testify:**
  - `require` — для предусловий (`Len`, `NoError`, `True(panicked)`, `True(ok)` перед приведением), `assert` — для независимых проверок.
  - Ошибки — `ErrorIs`, текст не сравнивается. `PanicsWithError` обоснованно не используется. `assert.Nil` для `Cycle` недопустимой буквы соответствует контракту (`nil`).
  - Объект `t` везде из своего подтеста.
  - В рабочих горутинах `TestConcurrentUse` нет `assert` и `require`: несовпадения пишутся в слот своей горутины, проверка идёт после `wg.Wait()` в горутине теста.
  - Порядок expected/actual верный.
- **Параллельность и стиль** (docs/rules/dev-test.md): `t.Parallel()` в каждом тесте верхнего уровня и подтесте (кроме примеров, где он неприменим). Табличные тесты с именованными случаями. Общего изменяемого состояния нет: `oracle()` возвращает массив по значению, `want` в подтестах только читается.
- **Конкурентность (golang-concurrency):** число горутин ограничено (`GOMAXPROCS × 4`), у каждой есть владелец (`wg.Go`, `wg.Wait`). Старт по закрытию канала, который закрывает единственный владелец. `time.Sleep` нет. Каждая горутина пишет только в свой элемент `failures`, чтение — после `Wait`: гонок в самом тесте нет. Изменяются только собственные полученные срезы.
- **`recoverPanic`** (`helpers_test.go:39–49`) корректен: в пути без паники именованные результаты — `nil, false`; при панике `panicked` остаётся `true`, значение берётся через `recover`.
- **Форматтеры и зависимости:**
  - `./.cache/bin/golangci-lint version` — 2.15.0.
  - `./.cache/bin/golangci-lint fmt --diff letter/*_test.go` — пусто, exit 0.
  - `gofmt -l letter/*_test.go` — пусто.
  - `golangci-lint config verify` — exit 0.
  - Порядок групп gci (стандартная → `github.com/go-muse/musex` → прочие) соблюдён.
  - Testify v1.11.1, косвенные зависимости — как разрешил Go.
  - `go mod tidy -diff` — пусто, exit 0. `go mod verify` — all modules verified. Каталога `vendor/` нет.
- **Ожидаемые падения до реализации** (исключение «тесты до реализации» стека Go) подтверждены повторным запуском:
  - `go vet ./...` и `go test -race -count=1 ./...`: `github.com/go-muse/musex/letter: no non-test Go files in …/letter`, `[build failed]`, exit 1.
  - `golangci-lint run ./...`: `typechecking error: no non-test Go files`, exit 7.
  - `go list -e ./letter`: `GoFiles=[]`, 6 `XTestGoFiles`, ошибок разбора нет.

  Причина одна — отсутствие продуктовых файлов, это не дефект тестов. Ни один тест и пример не выполнялся. Типы тестового кода, имена примеров (анализатор `tests`) и SA4017 будут проверены только после plan-4.
- **Вывод примеров** без текста ошибок и значений паники: печатаются только `String`, числа и `bool`.

## На будущее
- `TestPackageImports` (`letter/imports_test.go:18`) учитывает только файлы текущих GOOS/GOARCH. По плану QA дублирует проверку командой `go list`. Если в пакете появятся файлы с build-ограничениями, стоит проверять импорты через `go list -deps` для нужных платформ.
- После plan-4 первая сборка проверит типы тестов. Если появятся ошибки типов или замечания `go vet`/staticcheck на тестовых файлах, это возврат go-test-developer по обычному циклу.
