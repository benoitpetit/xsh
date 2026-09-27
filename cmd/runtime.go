package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/benoitpetit/xsh/core"
)

const coreExitError = core.ExitError

// OutputMode identifies the selected serialization format.
type OutputMode string

const (
	HumanOutput   OutputMode = "human"
	JSONOutput    OutputMode = "json"
	YAMLOutput    OutputMode = "yaml"
	CompactOutput OutputMode = "compact"
)

// Runtime contains the process-bound resources used by commands.
type Runtime struct {
	Context context.Context
	In      io.Reader
	Out     io.Writer
	Err     io.Writer
	Account string
	Mode    OutputMode
	Failure error
}

// CommandError carries the exit status that used to be applied directly by
// individual handlers. It is converted to a process exit only by Execute.
type CommandError struct {
	Code int
	Err  error
}

func (e *CommandError) Error() string {
	if e == nil || e.Err == nil {
		return "command failed"
	}
	return e.Err.Error()
}

func (e *CommandError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func abortCommand(code int) {
	panic(&CommandError{Code: code})
}

func commandExitCode(err error) int {
	if commandErr, ok := err.(*CommandError); ok && commandErr.Code != 0 {
		return commandErr.Code
	}
	return coreExitError
}

// ExitCode maps a command error to the process-level exit status.
func ExitCode(err error) int { return commandExitCode(err) }

var activeRuntime = &Runtime{
	Context: context.Background(),
	In:      os.Stdin,
	Out:     os.Stdout,
	Err:     os.Stderr,
	Mode:    HumanOutput,
}

func runtimeOutput() io.Writer {
	if activeRuntime == nil || activeRuntime.Out == nil {
		return io.Discard
	}
	return activeRuntime.Out
}

func runtimeError() io.Writer {
	if activeRuntime == nil || activeRuntime.Err == nil {
		return io.Discard
	}
	return activeRuntime.Err
}

func runtimeContext() context.Context {
	if activeRuntime == nil || activeRuntime.Context == nil {
		return context.Background()
	}
	return activeRuntime.Context
}

func recordRuntimeFailure(err error) {
	if err != nil && activeRuntime != nil && activeRuntime.Failure == nil {
		activeRuntime.Failure = err
	}
}

func withRuntime(runtime *Runtime, fn func()) {
	previous := activeRuntime
	activeRuntime = runtime
	defer func() { activeRuntime = previous }()
	fn()
}

func fmtDiagnostic(message string) {
	_, _ = fmt.Fprintln(runtimeError(), message)
}

func outputModeFromFlags() OutputMode {
	switch {
	case compactMode:
		return CompactOutput
	case yamlOutput:
		return YAMLOutput
	case jsonOutput:
		return JSONOutput
	default:
		return HumanOutput
	}
}

// ValidateOutputFlags rejects ambiguous serialization requests.
func ValidateOutputFlags(jsonFlag, yamlFlag, compactFlag bool) error {
	count := 0
	if jsonFlag {
		count++
	}
	if yamlFlag {
		count++
	}
	if compactFlag {
		count++
	}
	if count > 1 {
		return fmt.Errorf("output formats --json, --yaml, and --compact are mutually exclusive")
	}
	return nil
}
