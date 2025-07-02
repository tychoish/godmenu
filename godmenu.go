// Package godmenu is a wrapper around the dmenu program to easily
// interact with dmenu from Go applications.
package godmenu

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

var (
	ErrSelectionMissing     = errors.New("missing selection")
	ErrSelectionUnknown     = errors.New("unknown selection")
	ErrDmenuFailure         = errors.New("dmenu failed")
	ErrConfigurationInvalid = errors.New("invalid configuration")
)

// Run cals Do but takes its configuration as Option arguments.
func Run(ctx context.Context, opts ...Option) (string, error) {
	return Do(ctx, newop().resolve(opts).ref())
}

// Do shells out to dmenu with the given options and returns the
// selected result. If there was a problem with the command, the error
// is returned.
func Do(ctx context.Context, opts Configuration) (string, error) {
	opts.flags()

	selections := opts.selections()

	if err := selections.validate(); err != nil {
		return "", err
	}

	cmd := exec.CommandContext(ctx, opts.Flags.Path, opts.Flags.args()...)

	cmd.Stdin = bytes.NewBuffer(selections.rendered(opts.Sorted))

	return selections.processOutput(cmd.CombinedOutput())
}

// Configuration defines an DMenu operation.
type Configuration struct {
	// Selections are the options presented to dmenu.
	Selections []string
	// Flags, which may be nil--resulting in the defaults defined
	// in the godmenu package--describe the commandline options
	// passed to DMenu.
	Flags *Flags
	// Sorted, when true, causes godmenu to sort the Selections
	// before they're passed to DMenu.
	Sorted bool
}

func newop() *Configuration                           { return &Configuration{} }
func (op *Configuration) extendSelections(s []string) { op.Selections = append(op.Selections, s...) }
func (op *Configuration) selections() *set            { return newset(op.Selections) }
func (op Configuration) ref() Configuration           { return op }

func (op *Configuration) flags() {
	if op.Flags == nil {
		conf := defaultDmenuConfig
		op.Flags = &conf

	}
	op.Flags.fillDefault()
}

func (op *Configuration) resolve(opts []Option) *Configuration {
	for _, opt := range opts {
		opt(op)
	}

	return op
}

type set struct {
	set   map[string]int
	items []string
}

func newset(in []string) *set { s := &set{}; return s.init(in) }
func (s *set) Len() int       { return len(s.set) }

func (s *set) init(in []string) *set {
	s.set = make(map[string]int, len(in))
	s.items = in
	for idx := range in {
		k := strings.TrimSpace(s.items[idx])
		if k == "" {
			continue
		}
		s.set[k] = idx
		s.items[idx] = k
	}
	return s
}

func (s *set) validate() error {
	if len(s.items) == 0 {
		return fmt.Errorf("must define selections: %w", ErrConfigurationInvalid)
	}
	if diff := len(s.items) - len(s.set); diff != 0 {
		return fmt.Errorf("found %d duplicate selections: %w", diff, ErrConfigurationInvalid)
	}
	return nil
}

func (s *set) check(key string) bool {
	if s == nil || s.set == nil || key == "" {
		return false
	}
	_, ok := s.set[key]
	return ok
}

func (s *set) rendered(shouldSort bool) []byte {
	out := s.selections()
	if shouldSort {
		sort.Strings(out)
	}
	return []byte(strings.Join(out, "\n"))
}

func (s set) selections() []string {
	out := make([]string, 0, len(s.set))
	for idx := range s.items {
		out = append(out, s.items[idx])
	}
	return out
}

func (s set) processOutput(data []byte, err error) (string, error) {
	if err != nil {
		return "", fmt.Errorf("dmenu failed [%s]: %w", string(data), err)
	}

	if len(data) == 0 {
		return "", ErrSelectionMissing
	}

	out := string(bytes.TrimSpace(data))

	if out == "" {
		return "", ErrSelectionMissing
	}

	if !s.check(out) {
		return "", fmt.Errorf("value %q was not provided: %w", out, ErrSelectionUnknown)
	}

	return out, nil
}
