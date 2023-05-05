// Package godmenu is a wrapper around the dmenu program to easily
// interact with dmenu from Go applications.
package godmenu

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

// Options defines an DMenu operation.
type Options struct {
	// Selections are the options presented to dmenu.
	Selections []string
	// DMenu, which may be nil--resulting in the default
	// option--which are passed to dmenu.
	DMenu *DMenuConfiguration
}

// DMenuConfiguration defines how GoDMenu interacts with DMenu.
type DMenuConfiguration struct {
	Path            string
	BackgroundColor string
	ForegroundColor string
	Font            string
	CaseSensitive   bool
	Sorted          bool
}

const (
	DefaultBackgroundColor = "#000000"
	DefaultForegroundColor = "#ffffff"
	DefaultFount           = "Source Code Pro-11"
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

// RunDMenu shells out to dmenu with the given options and returns the
// selected result. If there was a problem with the command, the error
// is returned.
func RunDMenu(ctx context.Context, opts Options) (string, error) {
	conf := defaultDmenuConfig
	if opts.DMenu != nil {
		conf = *opts.DMenu
	}

	args := []string{
		"-nb", conf.BackgroundColor,
		"-nf", conf.ForegroundColor,
		"-fn", conf.Font,
	}

	if !conf.CaseSensitive {
		args = append(args, "-i")
	}

	selections := opts.Selections
	if conf.Sorted {
		sort.Strings(selections)
	}

	cmd := exec.CommandContext(ctx, conf.Path, args...)
	input := strings.TrimSpace(strings.Join(selections, "\n"))
	cmd.Stdin = bytes.NewBuffer([]byte(input))
	selection, err := cmd.CombinedOutput()
	out := strings.TrimSpace(string(selection))
	if err != nil {
		return "", fmt.Errorf("dmenu failed [%s]: %w", out, err)
	}

	return strings.TrimSpace(string(out)), nil
}
