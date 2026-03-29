# Git Frame

Утилита для подсчёта статистики по авторам или коммиттерам в Git-репозитории

## Требования

- Go 1.21+

## Сборка и запуск

Сборка:

```bash
git clone https://github.com/fanoxiz/git-frame.git
cd git-frame
go mod download
go build -o gitframe ./src
```

Запуск (Windows):

```bash
.\gitframe.exe [flags]
```

Запуск (Linux / MacOS):

```bash
./gitframe [flags]
```

Запуск без сборки:

```bash
go run ./src [флаги]
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
./gitframe --extensions .go --order-by commits
```

По языкам + JSON:

```bash
./gitframe --languages golang,python --format json
```

Исключить тесты и ограничить папкой `cmd`:

```bash
./gitframe --exclude "*_test.go" --restrict-to "cmd/*"
```

Использовать committer:

```bash
./gitframe --use-committer
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
