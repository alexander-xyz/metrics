// Multichecker — статический анализатор проекта сбора метрик.
//
// # Запуск
//
// Собрать и запустить по всем пакетам проекта:
//
//	go build -o cmd/staticlint/staticlint ./cmd/staticlint
//	./cmd/staticlint/staticlint ./...
//
// Проверить отдельный пакет или файл:
//
//	./cmd/staticlint/staticlint ./internal/handler
//	./cmd/staticlint/staticlint internal/handler/handler.go
//
// Запустить только выбранные анализаторы — по имени анализатора как флагу:
//
//	./cmd/staticlint/staticlint -osexit ./...
//	./cmd/staticlint/staticlint -printf -shadow ./...
//
// Полный список анализаторов с описанием флагов:
//
//	./cmd/staticlint/staticlint help
//
// # Состав
//
// Анализатор собран из четырёх групп.
//
// Стандартные анализаторы golang.org/x/tools/go/analysis/passes:
//
//   - appends — вызов append без значений, результат которого никуда не идёт;
//   - asmdecl — расхождение объявлений ассемблерных функций с их Go-сигнатурами;
//   - assign — бессмысленные самоприсваивания вида x = x;
//   - atomic — неверное использование sync/atomic: результат операции потерян;
//   - atomicalign — невыровненные аргументы 64-битных atomic-операций;
//   - bools — избыточные или ошибочные операции над булевыми выражениями;
//   - buildssa — построение SSA-представления для других анализаторов;
//   - buildtag — некорректные директивы сборки //go:build;
//   - cgocall — прямые вызовы указателей на C-функции;
//   - composite — составные литералы без имён полей;
//   - copylock — копирование значений, содержащих sync.Locker;
//   - ctrlflow — построение графа потока управления для других анализаторов;
//   - deepequalerrors — вызовы reflect.DeepEqual для значений типа error;
//   - defers — распространённые ошибки в defer, например defer time.Since(t);
//   - directive — некорректные директивы вида //go:debug;
//   - errorsas — второй аргумент errors.As не указатель на error;
//   - fieldalignment — порядок полей структуры, увеличивающий её размер;
//   - findcall — служебный анализатор поиска вызовов по имени, для отладки;
//   - framepointer — ассемблерный код, затирающий регистр указателя кадра;
//   - httpmux — устаревшие шаблоны маршрутов net/http.ServeMux;
//   - httpresponse — ошибки работы с ответом HTTP, например defer до проверки err;
//   - ifaceassert — невозможные утверждения типа для интерфейсов;
//   - inspect — служебный анализатор обхода AST, используется остальными;
//   - loopclosure — захват переменной цикла вложенной функцией;
//   - lostcancel — потерянная функция отмены контекста, ведёт к утечке;
//   - nilfunc — бессмысленные сравнения функций с nil;
//   - nilness — разыменование заведомо нулевых указателей;
//   - pkgfact — служебный механизм передачи фактов между пакетами;
//   - printf — несоответствие спецификаторов формата типам аргументов;
//   - reflectvaluecompare — сравнение reflect.Value оператором == вместо метода;
//   - shadow — затенение переменной новой декларацией с тем же именем;
//   - shift — сдвиги, превышающие разрядность значения;
//   - sigchanyzer — signal.Notify на небуферизованном канале;
//   - slog — неверные пары ключ-значение в вызовах log/slog;
//   - sortslice — вызов sort.Slice для аргумента, не являющегося срезом;
//   - stdmethods — методы с именами стандартных интерфейсов и неверной сигнатурой;
//   - stdversion — использование символов новее объявленной версии языка;
//   - stringintconv — конверсии вида string(int), дающие неожиданный результат;
//   - structtag — ошибки в тегах полей структур и дублирующиеся имена;
//   - testinggoroutine — вызов t.Fatal из горутины, порождённой тестом;
//   - tests — ошибки в сигнатурах тестов, бенчмарков и примеров;
//   - timeformat — вызовы time.Format с форматом 2006-02-01;
//   - unmarshal — передача не-указателя в json.Unmarshal и подобные функции;
//   - unreachable — недостижимый код;
//   - unsafeptr — недопустимые преобразования uintptr в unsafe.Pointer;
//   - unusedresult — проигнорированный результат вызова, например fmt.Sprintf;
//   - unusedwrite — запись в поле структуры, значение которой не используется;
//   - usesgenerics — служебный анализатор, сообщающий об использовании дженериков;
//   - waitgroup — неверное использование sync.WaitGroup.
//
// Анализаторы honnef.co/go/tools:
//
//   - все проверки класса SA (staticcheck) — ошибки, приводящие к неверной
//     работе программы: недостижимые ветки, утечки, некорректные сравнения;
//   - S1000 (simple) — упрощение конструкций, например select с одним case;
//   - ST1000 (stylecheck) — оформление комментария к пакету;
//   - QF1003 (quickfix) — замена цепочки if-else на switch.
//
// Публичные анализаторы:
//
//   - ineffassign (github.com/gordonklaus/ineffassign) — присваивания,
//     результат которых нигде не читается;
//   - bodyclose (github.com/timakin/bodyclose) — незакрытое тело ответа
//     http.Response, приводящее к утечке соединений.
//
// Собственный анализатор:
//
//   - osexit — запрещает прямой вызов os.Exit в функции main пакета main.
package main

import (
	"strings"

	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpmux"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/slog"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stdversion"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"golang.org/x/tools/go/analysis/passes/waitgroup"
	"honnef.co/go/tools/analysis/lint"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"github.com/alexander-xyz/metrics/cmd/staticlint/osexit"
)

// otherChecks перечисляет анализаторы staticcheck вне класса SA,
// которые включены в multichecker.
var otherChecks = map[string]bool{
	"S1000":  true,
	"ST1000": true,
	"QF1003": true,
}

// standardAnalyzers возвращает стандартные анализаторы пакета passes.
func standardAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		appends.Analyzer,
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		atomicalign.Analyzer,
		bools.Analyzer,
		buildssa.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		ctrlflow.Analyzer,
		deepequalerrors.Analyzer,
		defers.Analyzer,
		directive.Analyzer,
		errorsas.Analyzer,
		fieldalignment.Analyzer,
		findcall.Analyzer,
		framepointer.Analyzer,
		httpmux.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		inspect.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		nilness.Analyzer,
		pkgfact.Analyzer,
		printf.Analyzer,
		reflectvaluecompare.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		slog.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		stdversion.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		testinggoroutine.Analyzer,
		tests.Analyzer,
		timeformat.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
		unusedwrite.Analyzer,
		usesgenerics.Analyzer,
		waitgroup.Analyzer,
	}
}

// honnefAnalyzers возвращает все анализаторы класса SA и выбранные
// анализаторы остальных классов honnef.co/go/tools.
func honnefAnalyzers() []*analysis.Analyzer {
	var analyzers []*analysis.Analyzer

	for _, check := range staticcheck.Analyzers {
		if strings.HasPrefix(check.Analyzer.Name, "SA") {
			analyzers = append(analyzers, check.Analyzer)
		}
	}

	for _, checks := range [][]*lint.Analyzer{simple.Analyzers, stylecheck.Analyzers, quickfix.Analyzers} {
		for _, check := range checks {
			if otherChecks[check.Analyzer.Name] {
				analyzers = append(analyzers, check.Analyzer)
			}
		}
	}

	return analyzers
}

func main() {
	analyzers := standardAnalyzers()
	analyzers = append(analyzers, honnefAnalyzers()...)
	analyzers = append(analyzers,
		ineffassign.Analyzer,
		bodyclose.Analyzer,
		osexit.Analyzer,
	)

	multichecker.Main(analyzers...)
}
