package anbuGenerics

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	u "github.com/tanq16/anbu/internal/utils"
)

type TaskStatus string

const (
	TaskPending TaskStatus = "pending"
	TaskDone    TaskStatus = "done"
)

type TaskEntry struct {
	ID        int        `json:"id"`
	Task      string     `json:"task"`
	Status    TaskStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
}

type TaskStore struct {
	NextID int         `json:"next_id"`
	Tasks  []TaskEntry `json:"tasks"`
}

func getTasksFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	anbuDir := filepath.Join(homeDir, ".config", "anbu")
	if err := os.MkdirAll(anbuDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(anbuDir, "tasks.json"), nil
}

func loadTaskStore() (*TaskStore, error) {
	tasksPath, err := getTasksFilePath()
	if err != nil {
		return nil, err
	}
	store := &TaskStore{
		NextID: 1,
		Tasks:  []TaskEntry{},
	}
	data, err := os.ReadFile(tasksPath)
	if err != nil {
		if os.IsNotExist(err) {
			return store, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, store); err != nil {
		return nil, err
	}
	return store, nil
}

func saveTaskStore(store *TaskStore) error {
	tasksPath, err := getTasksFilePath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(tasksPath, data, 0644)
}

func findTaskByID(store *TaskStore, id int) *TaskEntry {
	for i := range store.Tasks {
		if store.Tasks[i].ID == id {
			return &store.Tasks[i]
		}
	}
	return nil
}

func removeTaskByID(store *TaskStore, id int) bool {
	for i, task := range store.Tasks {
		if task.ID == id {
			store.Tasks = append(store.Tasks[:i], store.Tasks[i+1:]...)
			return true
		}
	}
	return false
}

func TasksList(showDone bool, filter string) error {
	store, err := loadTaskStore()
	if err != nil {
		return err
	}
	if len(store.Tasks) == 0 {
		u.PrintInfo("No tasks found")
		return nil
	}
	sort.Slice(store.Tasks, func(i, j int) bool {
		return store.Tasks[i].ID > store.Tasks[j].ID
	})
	var filterRe *regexp.Regexp
	if filter != "" {
		filterRe, err = regexp.Compile(filter)
		if err != nil {
			return fmt.Errorf("invalid filter regex: %w", err)
		}
	}
	var headers []string
	if showDone {
		headers = []string{"ID", "Task", "Status", "Added"}
	} else {
		headers = []string{"ID", "Task", "Added"}
	}
	table := u.NewTable(headers)
	for _, task := range store.Tasks {
		if !showDone && task.Status == TaskDone {
			continue
		}
		if filterRe != nil && !filterRe.MatchString(task.Task) {
			continue
		}
		timeAgo := formatTimeAgo(time.Since(task.CreatedAt))
		if showDone {
			table.Rows = append(table.Rows, []string{
				fmt.Sprintf("%d", task.ID),
				task.Task,
				string(task.Status),
				timeAgo,
			})
		} else {
			table.Rows = append(table.Rows, []string{
				fmt.Sprintf("%d", task.ID),
				task.Task,
				timeAgo,
			})
		}
	}
	if len(table.Rows) == 0 {
		u.PrintInfo("No matching tasks")
		return nil
	}
	table.PrintTable(false)
	return nil
}

func TasksAdd() error {
	task := u.GetInput("Enter task:", "What needs to be done?")
	if task == "" {
		return fmt.Errorf("no task provided")
	}
	store, err := loadTaskStore()
	if err != nil {
		return err
	}
	entry := TaskEntry{
		ID:        store.NextID,
		Task:      task,
		Status:    TaskPending,
		CreatedAt: time.Now(),
	}
	store.Tasks = append(store.Tasks, entry)
	store.NextID++
	if err := saveTaskStore(store); err != nil {
		return err
	}
	u.PrintGeneric(fmt.Sprintf("%s %s %s", u.FDebug(task), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess(fmt.Sprintf("Added (ID: %d)", entry.ID))))
	return nil
}

func TasksDone(id int) error {
	store, err := loadTaskStore()
	if err != nil {
		return err
	}
	task := findTaskByID(store, id)
	if task == nil {
		return fmt.Errorf("task ID not found")
	}
	task.Status = TaskDone
	if err := saveTaskStore(store); err != nil {
		return err
	}
	u.PrintGeneric(fmt.Sprintf("%s %s %s", u.FDebug(fmt.Sprintf("%d", id)), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess(fmt.Sprintf("Marked done (ID: %d)", id))))
	return nil
}

func TasksDelete(id int) error {
	store, err := loadTaskStore()
	if err != nil {
		return err
	}
	if findTaskByID(store, id) == nil {
		return fmt.Errorf("task ID not found")
	}
	if !removeTaskByID(store, id) {
		return fmt.Errorf("failed to remove task")
	}
	if len(store.Tasks) == 0 {
		store.NextID = 1
	}
	if err := saveTaskStore(store); err != nil {
		return err
	}
	u.PrintGeneric(fmt.Sprintf("%s %s %s", u.FDebug(fmt.Sprintf("%d", id)), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess(fmt.Sprintf("Deleted (ID: %d)", id))))
	return nil
}
