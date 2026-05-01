package main

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
)

// Task represents a unit of work with optional dependencies.
type Task struct {
	ID   string
	Deps []string
	Run  func(ctx context.Context, depResults map[string]any) (any, error)
}

type TaskResult struct {
	ID     string
	Result any
	Err    error
}

var ErrCycle = errors.New("cycle detected in task graph")

func getDeps(ctx context.Context, resultChan <-chan TaskResult, task Task) (map[string]any, error) {

	if len(task.Deps) == 0 {
		return nil, nil
	}

	deps := make(map[string]any, len(task.Deps))
	for _, depName := range task.Deps {
		deps[depName] = nil
	}

major:
	for {
		select {
		case result, ok := <-resultChan:
			// Check is it dependency result
			if _, ok = deps[result.ID]; !ok {
				continue
			}

			// If got a depended task responded with error,
			// Here is no reason to wait longer.
			if result.Err != nil {
				return nil, fmt.Errorf("task failed : depended task got fail : %w", result.Err)
			}

			deps[result.ID] = result.Result

			// Check if all dependencies' results are collected
			for _, depValue := range deps {
				if depValue == nil {
					continue major
				}
			}
			break major
		case <-ctx.Done():
			return nil, ctx.Err()
		}

	}
	return deps, nil
}

type ResultStorage struct {
	totalResultMutex sync.Mutex
	totalResult      map[string]any
	totalResultChan  chan map[string]any
	subscribers      map[string][]chan TaskResult
}

func NewResultStorage() *ResultStorage {

	storage := &ResultStorage{
		totalResult:     map[string]any{},
		totalResultChan: make(chan map[string]any, 1),
		subscribers:     map[string][]chan TaskResult{},
	}

	return storage
}

func (s *ResultStorage) Close()

func (s *ResultStorage) Subscribe(task Task) <-chan TaskResult {
	s.totalResult[task.ID] = nil

	if len(task.Deps) == 0 {
		return nil
	}

	dependencyTaskResultChan := make(chan TaskResult)
	for _, dep := range task.Deps {
		subGroup, ok := s.subscribers[dep]
		if !ok {
			subGroup = []chan TaskResult{}
		}
		subGroup = append(subGroup, dependencyTaskResultChan)
		s.subscribers[dep] = subGroup
	}
	return dependencyTaskResultChan
}

func (s *ResultStorage) saveResult(result TaskResult) {
	s.totalResultMutex.Lock()
	defer s.totalResultMutex.Unlock()

	if result.Err != nil {
		s.totalResult[result.ID] = result.Err
	} else {
		s.totalResult[result.ID] = result.Result
	}

	for _, value := range s.totalResult {
		if value == nil {
			return
		}
	}

	filteredResult := map[string]any{}
	for key, value := range s.totalResult {
		switch value.(type) {
		case error:
			continue
		default:
			filteredResult[key] = value
		}
	}

	s.totalResultChan <- filteredResult
}

func (s *ResultStorage) Broadcast(result TaskResult) {
	defer s.saveResult(result)

	subGroup, ok := s.subscribers[result.ID]
	if !ok {
		return
	}
	for _, subChannel := range subGroup {
		subChannel <- result
	}
}

func (s *ResultStorage) Finish() chan map[string]any {
	return s.totalResultChan
}

type TaskProc struct {
	Task       Task
	DepResults map[string]any
}

func cycledTasksCheck(tasks []Task, pipeline []string) error {

	// fmt.Println("pipeline: ", pipeline)

	filteredTasks := []Task{}
	for _, task := range tasks {
		if len(pipeline) == 0 {
			filteredTasks = append(filteredTasks, task)
			continue
		}

		lastDependency := pipeline[len(pipeline)-1]
		if slices.Contains(task.Deps, lastDependency) {
			filteredTasks = append(filteredTasks, task)
		}
	}

	// fmt.Println("deps:", filteredTasks)

	if len(filteredTasks) == 0 {
		return nil
	}

	for _, task := range filteredTasks {

		if slices.Contains(pipeline, task.ID) {
			return ErrCycle
		}

		newPipeline := []string{}
		newPipeline = append(newPipeline, pipeline...)
		newPipeline = append(newPipeline, task.ID)

		// fmt.Println("new pipeline:", newPipeline)

		if err := cycledTasksCheck(tasks, newPipeline); err != nil {
			return err
		}
	}

	return nil
}

func validate(tasks []Task) error {
	return cycledTasksCheck(tasks, nil)
}

// RunDAG executes tasks in dependency order, up to maxWorkers concurrently.
// Returns a map of taskID → result for all completed tasks.
// If a task fails, its dependents are skipped; independent tasks keep running.
func RunDAG(ctx context.Context, tasks []Task, maxWorkers int) (map[string]any, error) {

	if len(tasks) == 0 {
		return nil, nil
	}

	if err := validate(tasks); err != nil {
		return nil, err
	}

	storage := NewResultStorage()

	taskProcChan := make(chan TaskProc)
	defer close(taskProcChan)

	// Result Storage
	// 1. Subscription for tasks looking for results
	// 2. Fan out

	for _, task := range tasks {
		resultChan := storage.Subscribe(task)

		go func() {
			deps, err := getDeps(ctx, resultChan, task)
			if err != nil {
				storage.Broadcast(TaskResult{
					ID:  task.ID,
					Err: err,
				})
				return
			}
			taskProcChan <- TaskProc{
				Task:       task,
				DepResults: deps,
			}

		}()
	}

	// Task Processor
	go func(maxWorkers int) {

		business := make(chan struct{}, maxWorkers)
		for taskProc := range taskProcChan {

			business <- struct{}{}

			go func() {
				result, err := taskProc.Task.Run(ctx, taskProc.DepResults)
				storage.Broadcast(TaskResult{
					ID:     taskProc.Task.ID,
					Result: result,
					Err:    err,
				})
				<-business
			}()

		}
		close(business)
	}(maxWorkers)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case results := <-storage.Finish():
		return results, nil
	}
}

// ── demo ─────────────────────────────────────────────────────────────────────

func main() {
	tasks := []Task{
		// {
		// 	ID:   "A",
		// 	Deps: []string{"C"},
		// 	Run: func(ctx context.Context, depResults map[string]any) (any, error) {
		// 		return "A", nil
		// 	},
		// },
		// {
		// 	ID:   "B",
		// 	Deps: []string{"A"},
		// 	Run: func(ctx context.Context, depResults map[string]any) (any, error) {
		// 		return "B", nil
		// 	},
		// },
		// {
		// 	ID:   "C",
		// 	Deps: []string{"B"},
		// 	Run: func(ctx context.Context, depResults map[string]any) (any, error) {
		// 		return "C", nil
		// 	},
		// },
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
		{
			ID:   "lint",
			Deps: []string{"generate"},
			Run: func(ctx context.Context, deps map[string]any) (any, error) {
				files := deps["generate"].([]string)
				return fmt.Sprintf("linted %d files", len(files)), nil
			},
		},
	}

	results, err := RunDAG(context.Background(), tasks, 2)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	for id, res := range results {
		fmt.Printf("[%s] → %v\n", id, res)
	}
}
