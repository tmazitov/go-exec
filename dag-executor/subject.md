# DAG Executor — Параллельный планировщик задач

Реализуй параллельный исполнитель задач, где каждая задача может зависеть от других.

## Контекст

Представь систему CI/CD или систему сборки (как `make` или `Gradle`):  
задача `"test"` зависит от `"build"`, `"build"` зависит от `"generate"` и `"download-deps"`.  
Задачи без зависимостей между собой должны выполняться **параллельно**.

```
download-deps ──┐
                ├──► build ──► test ──► deploy
generate ───────┘
```

## Что нужно реализовать

### Типы

```go
type Task struct {
    ID   string
    Deps []string // ID задач, которые должны завершиться до старта этой
    Run  func(ctx context.Context, depResults map[string]any) (any, error)
}
```

- `ID` — уникальный идентификатор задачи  
- `Deps` — список зависимостей (ID других задач из того же набора)  
- `Run` — функция задачи; получает результаты своих зависимостей через `depResults`

### Функция

```go
func RunDAG(ctx context.Context, tasks []Task, maxWorkers int) (map[string]any, error)
```

- Возвращает `map[taskID]result` для всех успешно завершённых задач  
- При ошибке в любой задаче — прекращает выполнение зависимых от неё задач (не всех!)  
- `maxWorkers` ограничивает количество одновременно работающих горутин  
- Уважает отмену `ctx`

## Требования

### 1. Параллельность
Задачи без взаимных зависимостей выполняются **параллельно**.

### 2. Порядок
Задача стартует **только** когда все её `Deps` успешно завершены.

### 3. Частичная отмена при ошибке
Если задача `A` упала с ошибкой — все задачи, зависящие от `A` (напрямую или транзитивно), должны быть **пропущены** (не запущены).  
Независимые задачи продолжают работать.

### 4. Ограничение горутин
Не более `maxWorkers` задач одновременно в выполнении.

### 5. Отмена контекста
Если `ctx` отменён — немедленно прекратить запуск новых задач и вернуть `ctx.Err()`.

### 6. Обнаружение цикла
Если в графе есть цикл (`A → B → A`) — вернуть ошибку немедленно, без запуска задач.

### 7. Отсутствие утечек горутин
После возврата из `RunDAG` все горутины должны завершиться.

## Пример использования

```go
tasks := []Task{
    {
        ID:   "download",
        Deps: []string{},
        Run: func(ctx context.Context, _ map[string]any) (any, error) {
            return "v1.2.3", nil
        },
    },
    {
        ID:   "generate",
        Deps: []string{},
        Run: func(ctx context.Context, _ map[string]any) (any, error) {
            return []string{"file_a.go", "file_b.go"}, nil
        },
    },
    {
        ID:   "build",
        Deps: []string{"download", "generate"},
        Run: func(ctx context.Context, deps map[string]any) (any, error) {
            version := deps["download"].(string)
            files := deps["generate"].([]string)
            return fmt.Sprintf("built %d files at %s", len(files), version), nil
        },
    },
    {
        ID:   "test",
        Deps: []string{"build"},
        Run: func(ctx context.Context, deps map[string]any) (any, error) {
            artifact := deps["build"].(string)
            return fmt.Sprintf("tested: %s", artifact), nil
        },
    },
}

results, err := RunDAG(context.Background(), tasks, 2)
// results["build"] == "built 2 files at v1.2.3"
// results["test"]  == "tested: built 2 files at v1.2.3"
```

## Подсказки

- Храни счётчик «сколько зависимостей ещё не выполнено» для каждой задачи — это классическая техника топологической сортировки (Kahn's algorithm).
- Канал `ready` — хорошее место, чтобы задачи «объявляли» о своей готовности к запуску.
- Семафор через `chan struct{}` поможет ограничить `maxWorkers`.
- Будь осторожен с состоянием гонок: результаты задач читают несколько горутин одновременно.

## Tags
`Concurrency` `Channels` `Context` `Semaphore` `Graph` `Hard`
