package cmd

import (
	"fmt"
	"strings"

	"github.com/danielmrdev/dtasks-cli/internal/output"
	"github.com/danielmrdev/dtasks-cli/internal/repo"
	"github.com/spf13/cobra"
)

var recurCmd = &cobra.Command{
	Use:   "recur",
	Short: "Manage task recurrence",
}

var (
	recurEvery  int
	recurTime   string
	recurDay    string // for weekly: mon,tue,wed,thu,fri,sat,sun
	recurDayNum int    // for monthly: 1-31
	recurStarts string
	recurEnds   string // YYYY-MM-DD or "never"
	recurAfter  int
)

var recurDailyCmd = &cobra.Command{
	Use:   "daily <task-id>",
	Short: "Set daily recurrence (e.g. every 3 days)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseID(args[0])
		if err != nil {
			return err
		}

		r := repo.RecurInput{
			Type:     "daily",
			Interval: recurEvery,
			EndsType: resolveEndsType(cmd),
		}
		if cmd.Flags().Changed("time") {
			r.Time = &recurTime
		}
		if cmd.Flags().Changed("starts") {
			r.Starts = &recurStarts
		}
		applyEnds(cmd, &r)

		if err := repo.TaskSetRecur(DB, id, r); err != nil {
			return err
		}
		output.PrintSuccess(fmt.Sprintf("Task #%d set to recur daily every %d day(s)", id, recurEvery))
		return nil
	},
}

var recurWeeklyCmd = &cobra.Command{
	Use:   "weekly <task-id>",
	Short: "Set weekly recurrence (e.g. every 2 weeks on Thursday)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseID(args[0])
		if err != nil {
			return err
		}
		if !cmd.Flags().Changed("day") {
			return fmt.Errorf("--day is required for weekly recurrence (mon,tue,wed,thu,fri,sat,sun)")
		}
		dow, err := parseDayOfWeek(recurDay)
		if err != nil {
			return err
		}

		r := repo.RecurInput{
			Type:      "weekly",
			Interval:  recurEvery,
			DayOfWeek: &dow,
			EndsType:  resolveEndsType(cmd),
		}
		if cmd.Flags().Changed("time") {
			r.Time = &recurTime
		}
		if cmd.Flags().Changed("starts") {
			r.Starts = &recurStarts
		}
		applyEnds(cmd, &r)

		if err := repo.TaskSetRecur(DB, id, r); err != nil {
			return err
		}
		output.PrintSuccess(fmt.Sprintf("Task #%d set to recur weekly every %d week(s) on day %d", id, recurEvery, dow))
		return nil
	},
}

var recurMonthlyCmd = &cobra.Command{
	Use:   "monthly <task-id>",
	Short: "Set monthly recurrence (e.g. every 3 months on day 1)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseID(args[0])
		if err != nil {
			return err
		}
		if !cmd.Flags().Changed("day") {
			return fmt.Errorf("--day is required for monthly recurrence (1-31)")
		}
		if recurDayNum < 1 || recurDayNum > 31 {
			return fmt.Errorf("--day must be between 1 and 31")
		}

		r := repo.RecurInput{
			Type:       "monthly",
			Interval:   recurEvery,
			DayOfMonth: &recurDayNum,
			EndsType:   resolveEndsType(cmd),
		}
		if cmd.Flags().Changed("time") {
			r.Time = &recurTime
		}
		if cmd.Flags().Changed("starts") {
			r.Starts = &recurStarts
		}
		applyEnds(cmd, &r)

		if err := repo.TaskSetRecur(DB, id, r); err != nil {
			return err
		}
		output.PrintSuccess(fmt.Sprintf("Task #%d set to recur monthly every %d month(s) on day %d", id, recurEvery, recurDayNum))
		return nil
	},
}

var recurRmCmd = &cobra.Command{
	Use:   "rm <task-id>",
	Short: "Remove recurrence from a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseID(args[0])
		if err != nil {
			return err
		}
		if err := repo.TaskRemoveRecur(DB, id); err != nil {
			return err
		}
		output.PrintSuccess(fmt.Sprintf("Recurrence removed from task #%d", id))
		return nil
	},
}

func init() {
	for _, c := range []*cobra.Command{recurDailyCmd, recurWeeklyCmd, recurMonthlyCmd} {
		c.Flags().IntVar(&recurEvery, "every", 1, "Interval (e.g. every 3 days/weeks/months)")
		c.Flags().StringVar(&recurTime, "time", "", "Time of occurrence (HH:MM)")
		c.Flags().StringVar(&recurStarts, "starts", "", "Start date (YYYY-MM-DD)")
		c.Flags().StringVar(&recurEnds, "ends", "", "End date (YYYY-MM-DD) or \"never\"")
		c.Flags().IntVar(&recurAfter, "ends-after", 0, "End after N repetitions")
	}
	recurWeeklyCmd.Flags().StringVar(&recurDay, "day", "", "Day of week (mon,tue,wed,thu,fri,sat,sun)")
	recurMonthlyCmd.Flags().IntVar(&recurDayNum, "day", 0, "Day of month (1-31)")

	recurCmd.AddCommand(recurDailyCmd)
	recurCmd.AddCommand(recurWeeklyCmd)
	recurCmd.AddCommand(recurMonthlyCmd)
	recurCmd.AddCommand(recurRmCmd)
}

// --- helpers ---

func resolveEndsType(cmd *cobra.Command) string {
	if cmd.Flags().Changed("ends-after") {
		return "after_n"
	}
	if cmd.Flags().Changed("ends") {
		if recurEnds == "never" {
			return "never"
		}
		return "on_date"
	}
	return "never"
}

func applyEnds(cmd *cobra.Command, r *repo.RecurInput) {
	if r.EndsType == "on_date" && cmd.Flags().Changed("ends") {
		r.EndsDate = &recurEnds
	}
	if r.EndsType == "after_n" {
		r.EndsAfter = &recurAfter
	}
}

func parseDayOfWeek(s string) (int, error) {
	days := map[string]int{
		"sun": 0, "sunday": 0,
		"mon": 1, "monday": 1,
		"tue": 2, "tuesday": 2,
		"wed": 3, "wednesday": 3,
		"thu": 4, "thursday": 4,
		"fri": 5, "friday": 5,
		"sat": 6, "saturday": 6,
	}
	d, ok := days[strings.ToLower(strings.TrimSpace(s))]
	if !ok {
		return 0, fmt.Errorf("invalid day %q — use mon,tue,wed,thu,fri,sat,sun", s)
	}
	return d, nil
}
