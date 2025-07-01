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

// Operation defines an DMenu operation.
type Operation struct {
	// Selections are the options presented to dmenu.
	Selections []string
	// DMenu, which may be nil--resulting in the default
	// option--which are passed to dmenu.
	DMenu *Configuration
	// Strict, which must be explicitly enabled, causes RunDMenu
	// to fail if the selection returned from dmenu is not in the
	// selections provided in the option struct.
	Strict bool
}

func Run(ctx context.Context, opts ...Option) (string, error) {
	conf := Operation{}
	conf.DMenu.applyOptions(opts)
	return RunDMenu(ctx, conf)
}

type set map[string]struct{}

func (s set) add(key string) { s[key] = struct{}{} }

func (s set) extend(keys []string) {
	for _, k := range keys {
		s.add(k)
	}
}

func (s set) check(key string) bool {
	if s == nil {
		return false
	}
	_, ok := s[key]
	return ok
}

func renderCommandOutput(data []byte, err error) (string, error) {
	return string(bytes.TrimSpace(data)), err
}

func renderSelections(shouldSort bool, in []string) string {
	if shouldSort {
		sort.Strings(in)
	}
	for idx := range in {
		in[idx] = strings.TrimSpace(in[idx])
	}
	return strings.Join(in, "\n")
}

func (s set) processOutput(out string, err error) (string, error) {
	switch {
	case s != nil && out != "" && !s.check(out):
		return "", errors.Join(fmt.Errorf("selection %q", out), ErrSelectionUnknown)
	case out == "":
		return "", errors.Join(err, ErrSelectionMissing)
	case err == nil:
		return out, nil
	default:
		return "", fmt.Errorf("dmenu failed [%s]: %w", out, err)
	}
}

// RunDMenu shells out to dmenu with the given options and returns the
// selected result. If there was a problem with the command, the error
// is returned.
func RunDMenu(ctx context.Context, opts Operation) (string, error) {
	conf := defaultDmenuConfig
	if opts.DMenu != nil {
		conf = *opts.DMenu
	}
	if opts.Strict && len(opts.Selections) == 0 {
		return "", fmt.Errorf("must specify selections in strict mode: %w", ErrConfigurationInvalid)
	}

	var selections set
	if opts.Strict {
		selections = make(set, len(opts.Selections))
		selections.extend(opts.Selections)
	}

	cmd := exec.CommandContext(ctx, conf.Path, conf.resolveArgs()...)

	cmd.Stdin = bytes.NewBuffer([]byte(renderSelections(conf.Sorted, opts.Selections)))

	return selections.processOutput(renderCommandOutput(cmd.CombinedOutput()))
}
