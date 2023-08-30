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

// Options defines an DMenu operation.
type Options struct {
	// Selections are the options presented to dmenu.
	Selections []string
	// DMenu, which may be nil--resulting in the default
	// option--which are passed to dmenu.
	DMenu *DMenuConfiguration
	// Strict, which must be explicitly enabled, causes RunDMenu
	// to fail if the selection returned from dmenu is not in the
	// selections provided in the option struct.
	Strict bool
}

// DMenuConfiguration defines how GoDMenu interacts with DMenu.
type DMenuConfiguration struct {
	Path            string
	BackgroundColor string
	ForegroundColor string
	Font            string
	Prompt          string
	CaseSensitive   bool
	Sorted          bool
	Bottom          bool
	Lines           int
	Monitor         Optional[int]
	WindowID        Optional[int]
}

type Optional[T int] struct {
	v       T
	defined bool
}

func NewOptional[T int](in T) Optional[T]  { return Optional[T]{v: in, defined: true} }
func (o Optional[T]) Set(in T) Optional[T] { o.defined = true; o.v = in; return o }
func (o Optional[T]) Reset() Optional[T]   { return Optional[T]{} }
func (o Optional[T]) Resolve() T           { return o.v }
func (o Optional[T]) OK() bool             { return o.defined }

const (
	DefaultBackgroundColor = "#000000"
	DefaultForegroundColor = "#ffffff"
	DefaultFount           = "Source Code Pro-12"
	DefaultDmenuPath       = "dmenu"
)

var defaultDmenuConfig DMenuConfiguration

func init() { defaultDmenuConfig.SetDefaults() }

// SetDefaults sets any unset fields in the DMenuConfiguration that
// have the zero value. All of the default values are defined in
// package constants.
func (conf *DMenuConfiguration) SetDefaults() {
	if conf.Path == "" {
		conf.Path = DefaultDmenuPath
	}

	if conf.BackgroundColor == "" {
		conf.BackgroundColor = DefaultBackgroundColor
	}

	if conf.ForegroundColor == "" {
		conf.ForegroundColor = DefaultForegroundColor
	}

	if conf.Font == "" {
		conf.Font = DefaultFount
	}
}

func (conf DMenuConfiguration) resolveArgs() []string {
	args := []string{
		"-nb", conf.BackgroundColor,
		"-nf", conf.ForegroundColor,
		"-fn", conf.Font,
	}

	if !conf.CaseSensitive {
		args = append(args, "-i")
	}

	if conf.Bottom {
		args = append(args, "-b")
	}

	if conf.Prompt != "" {
		args = append(args, "-p", conf.Prompt)
	}

	if conf.Lines > 0 {
		args = append(args, "-l", fmt.Sprint(conf.Lines))
	}

	if conf.Monitor.OK() {
		args = append(args, "-m", fmt.Sprint(conf.Monitor.Resolve()))
	}

	if conf.WindowID.OK() {
		args = append(args, "-w", fmt.Sprint(conf.WindowID.Resolve()))
	}

	return args
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
func RunDMenu(ctx context.Context, opts Options) (string, error) {
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
