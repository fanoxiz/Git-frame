# Git Frame

Утилита для подсчёта статистики по авторам или коммиттерам в Git-репозитории

## Требования

- Go 1.25+
- GNU Make

## Сборка и запуск

Быстрый вариант:

```bash
go install github.com/fanoxiz/Git-frame/cmd/gitframe@latest
export PATH="$(go env GOPATH)/bin:$PATH"
```

Сборка с исходным кодом:

```bash
git clone https://github.com/fanoxiz/Git-frame.git
cd Git-frame
make build
./gitframe [флаги]
```

Запуск без установки:

```bash
make run ARGS="[флаги]"
```

## Аргументы CLI

| Флаг | Значение по умолчанию | Что означает |
|---|---|---|
| `--repository` | `"."` | Путь к Git-репозиторию |
| `--revision` | `"HEAD"` | Ревизия/коммит/ветка, по которой считается статистика |
| `--order-by` | `lines` | Поле сортировки: `lines` / `commits` / `files` |
| `--use-committer` | `false` | Использовать `committer` вместо `author` |
| `--format` | `tabular` | Формат вывода: `tabular` / `csv` / `json` / `json-lines` |
| `--extensions` | без ограничений | Включать только файлы с указанными расширениями |
| `--languages` | без ограничений | Включать только файлы языков из `configs/language_extensions.json` |
| `--exclude` | без ограничений | Исключать файлы, подходящие под glob-паттерны |
| `--restrict-to` | без ограничений | Включать только файлы, подходящие под glob-паттерны |

## Примеры

Только .go файлы, сортировка по коммитам:

```bash
gitframe --extensions .go --order-by commits
```

Исключить тесты и ограничить папкой `cmd`:

```bash
gitframe --exclude "*_test.go" --restrict-to "cmd/*"
```

Использовать committer:

```bash
gitframe --use-committer
```

## Конфиг языков

При использовании `--languages` утилита ищет файл `configs/language_extensions.json` по текущей директории и её родителям

Формат:

```json
[
  { "name": "golang", "extensions": [".go"] },
  { "name": "python", "extensions": [".py"] },
  { "name": "C++", "extensions": [".cpp", ".hpp", ".h"]}
]
```
