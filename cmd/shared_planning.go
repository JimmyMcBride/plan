package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	brainplanning "github.com/JimmyMcBride/brain/planning"
	brainapp "github.com/JimmyMcBride/brain/planning/application"
	brainlocal "github.com/JimmyMcBride/brain/planning/local"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

const sharedPlanningModuleID = "standalone.plan"

var (
	migrationWarningOnce sync.Once
	isInteractiveWriter  = defaultInteractiveWriter
)

type compatibilityAuthorizer struct{}

func (compatibilityAuthorizer) Require(context.Context, string) error { return nil }

type compatibilityEventSink struct{}

func (compatibilityEventSink) Publish(context.Context, brainapp.Event) error { return nil }

func withSharedPlanning(cmd *cobra.Command, equivalent string, machineOutput bool, run func(*brainapp.Service) error) (bool, error) {
	service, shared, err := sharedPlanningService(cmd.Context())
	if err != nil || !shared {
		return false, err
	}
	err = run(service)
	if sharedCompatibilityFallback(err) {
		return false, nil
	}
	warnSharedPlanning(cmd, equivalent, machineOutput)
	return true, err
}

func sharedPlanningService(ctx context.Context) (*brainapp.Service, bool, error) {
	service := brainapp.New(brainlocal.New(projectDir), brainapp.Options{
		ModuleID:    sharedPlanningModuleID,
		ProjectRoot: projectDir,
	})
	status, err := service.Status(ctx)
	if err != nil {
		return nil, false, err
	}
	shared := status.State == brainapp.WorkspaceCompatible && status.Ownership == brainplanning.OwnershipLocal
	return service, shared, nil
}

func warnSharedPlanning(cmd *cobra.Command, equivalent string, machineOutput bool) {
	if machineOutput || !isInteractiveWriter(cmd.ErrOrStderr()) {
		return
	}
	migrationWarningOnce.Do(func() {
		fmt.Fprintf(cmd.ErrOrStderr(), "Warning: this local command now uses Brain Planning; prefer `brain plan %s`.\n", equivalent)
	})
}

func defaultInteractiveWriter(writer io.Writer) bool {
	file, ok := writer.(*os.File)
	if !ok {
		return false
	}
	fd := file.Fd()
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}

func sharedCompatibilityFallback(err error) bool {
	return err != nil && strings.Contains(err.Error(), "invalid spec .plan/specs/")
}

func sharedAuthorizer() brainapp.Authorizer { return compatibilityAuthorizer{} }

func sharedEvents() brainapp.EventSink { return compatibilityEventSink{} }
