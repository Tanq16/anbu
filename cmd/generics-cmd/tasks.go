package genericsCmd

import (
	"strconv"

	"github.com/spf13/cobra"
	anbuGenerics "github.com/tanq16/anbu/internal/generics"
	u "github.com/tanq16/anbu/internal/utils"
)

var tasksFlags struct {
	done   bool
	filter string
	pipe   bool
}

var TasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "Lightweight personal task tracking with pending/done status",
}

var tasksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks (pending by default, use --done to include completed)",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if err := anbuGenerics.TasksList(tasksFlags.done, tasksFlags.filter); err != nil {
			u.PrintFatal("failed to list tasks", err)
		}
	},
}

var tasksAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new task interactively",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if err := anbuGenerics.TasksAdd(tasksFlags.pipe); err != nil {
			u.PrintFatal("failed to add task", err)
		}
	},
}

var tasksDoneCmd = &cobra.Command{
	Use:   "done <id>",
	Short: "Mark a task as done by its ID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			u.PrintFatal("invalid task ID", nil)
		}
		if err := anbuGenerics.TasksDone(id); err != nil {
			u.PrintFatal("failed to mark task done", err)
		}
	},
}

var tasksDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a task by its ID regardless of status",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			u.PrintFatal("invalid task ID", nil)
		}
		if err := anbuGenerics.TasksDelete(id); err != nil {
			u.PrintFatal("failed to delete task", err)
		}
	},
}

func init() {
	tasksListCmd.Flags().BoolVar(&tasksFlags.done, "done", false, "Show completed tasks alongside pending")
	tasksListCmd.Flags().StringVar(&tasksFlags.filter, "filter", "", "Filter tasks by regex pattern")
	tasksAddCmd.Flags().BoolVar(&tasksFlags.pipe, "pipe", false, "Read task from piped stdin instead of interactive input")
	TasksCmd.AddCommand(tasksListCmd)
	TasksCmd.AddCommand(tasksAddCmd)
	TasksCmd.AddCommand(tasksDoneCmd)
	TasksCmd.AddCommand(tasksDeleteCmd)
}
