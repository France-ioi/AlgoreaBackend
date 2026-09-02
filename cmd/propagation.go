package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql" // use to force database/sql to use mysql
	"github.com/spf13/cobra"

	"github.com/France-ioi/AlgoreaBackend/v2/app"
	"github.com/France-ioi/AlgoreaBackend/v2/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
)

const (
	propagationCommandLockName    = "propagation_command"
	propagationCommandLockTimeout = 600 * time.Second

	// Must exceed the worst-case single chunk duration plus the post-unlock permissions budget
	// (app/database.postUnlockPermissionsBudget = 30s), otherwise a budgeted run can still be
	// killed mid-transaction. Worst-case chunk duration is bounded by the lock-wait budget
	// (see plans/1-propagation-bounded-lock-waits.md): unblocked chunks run in ~0.16 s, blocked
	// ones ran 53.7-58.2 s on 2026-07-24. 90s leaves ~60s for one blocked chunk + 30s for
	// post-unlock computeAllAccess.
	propagationShutdownMargin = 90 * time.Second
)

func propagationLockTimeout(remainingBudget time.Duration) time.Duration {
	// Clamp at 0: MySQL treats a negative GET_LOCK timeout as an infinite wait.
	return max(min(propagationCommandLockTimeout, remainingBudget), 0)
}

func validatePropagationMaxDuration(maxDuration time.Duration) error {
	if maxDuration > 0 && maxDuration <= propagationShutdownMargin {
		return fmt.Errorf(
			"--max-duration must be greater than the %v shutdown margin (got %v)",
			propagationShutdownMargin, maxDuration)
	}
	return nil
}

func init() { //nolint:gochecknoinits // cobra suggests using init functions to add commands
	var propagationMaxDuration time.Duration

	propagationCmd := &cobra.Command{
		Use:   "propagation [environment]",
		Short: "apply propagation to the database",
		Long:  `runs items, permissions and results propagation`,
		Args:  cobra.MaximumNArgs(1),
		RunE:  runPropagationCommand(&propagationMaxDuration),
	}

	propagationCmd.Flags().DurationVar(&propagationMaxDuration, "max-duration", 0,
		"soft limit on the total run time; propagation stops between chunks once reached (0 = no limit)")

	rootCmd.AddCommand(propagationCmd)
}

func runPropagationCommand(propagationMaxDuration *time.Duration) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Include app.New() and the lock wait: that is what the Lambda bills.
		startTime := time.Now()

		if len(args) > 0 {
			appenv.SetEnv(args[0])
		}

		if err := validatePropagationMaxDuration(*propagationMaxDuration); err != nil {
			return err
		}

		application, err := app.New()
		defer func() {
			if application != nil && application.Database != nil {
				_ = application.Database.Close()
			}
		}()
		if err != nil {
			return err
		}

		lockTimeout := configurePropagationSoftDeadline(application, startTime, *propagationMaxDuration)
		stopSIGTERM := startPropagationSIGTERMHandler(application)
		defer stopSIGTERM()

		err = database.NewDataStore(application.Database).
			WithNamedLock(propagationCommandLockName, lockTimeout, func(s *database.DataStore) error {
				return s.InTransaction(func(store *database.DataStore) error {
					store.SchedulePermissionsPropagation()
					store.ScheduleResultsPropagation()
					return nil
				})
			})
		if err != nil {
			return handlePropagationLockError(cmd, application, *propagationMaxDuration, err)
		}

		printPropagationCompletion(cmd, application.Database)
		return nil
	}
}

func configurePropagationSoftDeadline(
	application *app.Application, startTime time.Time, maxDuration time.Duration,
) time.Duration {
	if maxDuration <= 0 {
		return propagationCommandLockTimeout
	}
	deadline := startTime.Add(maxDuration - propagationShutdownMargin)
	database.SetPropagationSoftDeadline(application.Database, deadline)
	lockTimeout := propagationLockTimeout(time.Until(deadline))
	logEntry := logging.EntryFromContext(application.Database.GetContext())
	logEntry.Infof("propagation soft deadline %v, named-lock timeout %v", deadline.UTC(), lockTimeout)
	if time.Until(deadline) <= 0 {
		logEntry.Warnf(
			"propagation soft deadline is already in the past (max-duration %v leaves only %v after the %v shutdown margin)",
			maxDuration, maxDuration-propagationShutdownMargin, propagationShutdownMargin)
	}
	return lockTimeout
}

// startPropagationSIGTERMHandler expires the soft deadline on SIGTERM so loops stop after the
// current committed chunk (exit 0 + early-stop warning). Closing the pool here would turn that into
// a non-zero failure on the next DB acquire; the deferred Close handles cleanup on return.
// Note: a plain CLI/Lambda invocation of this subcommand does not register a Lambda extension, so
// SIGTERM may not be delivered on the production propagation path today.
func startPropagationSIGTERMHandler(application *app.Application) (stop func()) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM)
	stopSignals := make(chan struct{})
	go func() {
		select {
		case <-sigCh:
			ctx := application.Database.GetContext()
			logging.EntryFromContext(ctx).Info(
				"Got SIGTERM, expiring soft deadline so propagation stops between chunks")
			database.SetPropagationSoftDeadline(application.Database, time.Now())
		case <-stopSignals:
		}
	}()
	return func() {
		signal.Stop(sigCh)
		close(stopSignals)
	}
}

func handlePropagationLockError(
	cmd *cobra.Command, application *app.Application, maxDuration time.Duration, err error,
) error {
	if maxDuration > 0 && errors.Is(err, database.ErrNamedLockWaitTimeoutExceeded) {
		logging.EntryFromContext(application.Database.GetContext()).Info(
			"propagation skipped: named lock not acquired within remaining time budget")
		cmd.Println("Propagation skipped: could not acquire lock within time budget.")
		return nil
	}
	return fmt.Errorf("error while doing propagation: %w", err)
}

func printPropagationCompletion(cmd *cobra.Command, db *database.DB) {
	if database.SoftDeadlineExceeded(db) {
		cmd.Println("Propagation stopped early: time budget exhausted.")
		return
	}
	cmd.Println("Propagation done.")
}
