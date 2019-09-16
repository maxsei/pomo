package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	cli "github.com/jawher/mow.cli"

	pomo "github.com/kevinschoon/pomo/pkg"
	"github.com/kevinschoon/pomo/pkg/config"
	"github.com/kevinschoon/pomo/pkg/functional"
	psort "github.com/kevinschoon/pomo/pkg/sort"
	"github.com/kevinschoon/pomo/pkg/store"
	"github.com/kevinschoon/pomo/pkg/tree"
)

func get(cfg *config.Config) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "[OPTIONS] [FILTER...]"
		cmd.LongDesc = `
Examples:

# output all tasks across all projects
pomo get
# output all tasks across all projects as a tree
pomo get --tree
        `
		var (
			flatten       = cmd.BoolOpt("f flatten", false, "flatten all projects to one level")
			showPomodoros = cmd.BoolOpt("p pomodoros", true, "show status of each pomodoro")
			recent        = cmd.BoolOpt("r recent", true, "sort by most recently modified tasks")
			ascend        = cmd.BoolOpt("a ascend", false, "sort from oldest to newest")
			filters       = cmd.StringsArg("FILTER", []string{}, "filters")
		)
		cmd.Action = func() {
			root := &pomo.Task{
				ID: int64(0),
			}
			db, err := store.NewSQLiteStore(cfg.DBPath)
			maybe(err)
			defer db.Close()
			maybe(db.With(func(db store.Store) error {
				return db.ReadTask(root)
			}))

			root.Tasks = functional.FindMany(root.Tasks, functional.FiltersFromStrings(*filters)...)

			functional.ForEachMutate(root, func(task *pomo.Task) {
				if *ascend {
					sort.Sort(sort.Reverse(psort.TasksByID(task.Tasks)))
				} else if *recent {
					sort.Sort(sort.Reverse(psort.TasksByStart(task.Tasks)))
				}
			})

			if cfg.JSON {
				maybe(json.NewEncoder(os.Stdout).Encode(root))
				return
			} else if *flatten {
				functional.ForEach(*root, func(task pomo.Task) {
					fmt.Println(task.Info())
				})
			} else {
				tree.Tree{Task: *root, ShowPomodoros: *showPomodoros}.Write(os.Stdout, nil)
			}
		}
	}
}
