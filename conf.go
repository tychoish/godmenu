package godmenu

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

const (
	DefaultDMenuPath               = "dmenu"
	DefaultFont                    = "Source Code Pro-13"
	DefaultBackgroundColor         = "#000000"
	DefaultTextColor               = "#ffffff"
	DefaultSelectedBackgroundColor = "#005577"
	DefaultSelectedTextColor       = "#ffffff"
)

var defaultDmenuConfig Flags

func init() { defaultDmenuConfig.fillDefault() }

// Options defines an DMenu operation.
type Options struct {
	// Selections are the options presented to dmenu.
	Selections []string
	// Flags, which may be nil--resulting in the defaults defined
	// in the godmenu package--describe the commandline options
	// passed to DMenu.
	Flags *Flags
	// Sorted, when true, causes godmenu to sort the Selections
	// before they're passed to DMenu.
	Sorted bool
	// AllowDuplicates
	AllowDuplicates bool
	// RequireMatch, when true requires that output of the dmenu
	// operation is in the selections operation.
	RequireMatch bool
	// Transform takes the selection returned by the user, and modifies it before. This is
	// useful for annotating messages or truncating longer messages.
	Transform func(string) string
	// ConfirmSubstitution instructs the application to display a second menu to confirm or
	// modify the _actual_ selection.
	ConfirmSubstitution bool
}

// Arg is a type for functional arguments.
type Arg func(*Options)

func newop() *Options                     { return &Options{} }
func (op Options) ref() Options           { return op }
func (op *Options) with(opt Arg) *Options { opt(op); return op }

func (op *Options) selections() *set {
	return newset(op.Selections).
		withRequireMatch(op.RequireMatch).
		withTransform(op.Transform).
		withAllowDuplicates(op.AllowDuplicates)
}

func (op *Options) flags() *Options {
	if op.Flags == nil {
		conf := defaultDmenuConfig
		op.Flags = &conf
	}

	return op
}

func (op *Options) apply(opts []Arg) *Options {
	op = op.flags()
	for _, opt := range opts {
		op = op.with(opt)
	}

	return op
}

func (op *Options) extendSelections(s []string) *Options {
	op.Selections = append(op.Selections, s...)
	return op
}

func (op *Options) validate() (*set, error) {
	op.flags()
	op.Flags.fillDefault()

	selections := op.selections()

	errs := []error{
		op.Flags.validate(),
		selections.validate(),
	}
	if op.RequireMatch && op.Transform != nil {
		errs = append(errs, errors.New("the combination of the requireMatch option and transform function is ambiguous."))
	}

	if op.Transform == nil && op.ConfirmSubstitution {
		errs = append(errs, errors.New("the confirmSubstitution option without the transform function is ambiguous."))
	}

	if err := errors.Join(errs...); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrConfigurationInvalid, err)
	}

	return selections, nil
}

// Flags defines how go-DMenu interacts with DMenu. You can either
// specify these either using the Option argument to godmenu.Run() or
// as part of the Configuration structure.
type Flags struct {
	Path              string
	BackgroundColor   string
	TextColor         string
	SelectedBgColor   string
	SelectedTextColor string
	Font              string
	Prompt            string
	CaseSensitive     bool
	Bottom            bool
	Lines             int
	Monitor           int
	WindowID          int
}

func (f *Flags) validate() error {
	var errs []error
	if _, err := exec.LookPath(f.Path); err != nil {
		errs = append(errs, fmt.Errorf("could not find path %q to dmenu: %w", f.Path, err))
	}

	if !possiblyValidColor(f.BackgroundColor) {
		errs = append(errs, fmt.Errorf("invalid background color %s", f.BackgroundColor))
	}
	if !possiblyValidColor(f.SelectedBgColor) {
		errs = append(errs, fmt.Errorf("invalid selected background color %s", f.SelectedBgColor))
	}
	if !possiblyValidColor(f.TextColor) {
		errs = append(errs, fmt.Errorf("invalid text color %s", f.SelectedBgColor))
	}
	if !possiblyValidColor(f.SelectedTextColor) {
		errs = append(errs, fmt.Errorf("invalid selected text color %s", f.SelectedTextColor))
	}

	// only 0 or larger values are valid. we don't pass to dmenu
	// if they're -1 (which is the "unset") value
	if f.Monitor < -1 {
		errs = append(errs, fmt.Errorf("invalid X11 montior id '%d'", f.Monitor))
	}
	if f.WindowID < -1 {
		errs = append(errs, fmt.Errorf("invalid X11 window id '%d'", f.WindowID))
	}

	return errors.Join(errs...)
}

func possiblyValidColor(color string) bool {
	if !strings.HasPrefix(color, "#") {
		// X11 color names are supported, and there isn't a
		// stdlib list.
		return true
	}

	// either '#RGB' or '#RRGGBB'
	l := len(color)
	if l != 4 && l != 7 {
		return false
	}

	_, err := strconv.ParseInt(color[1:], 16, 64)

	return err == nil
}

// fillDefault sets any unset fields in the DMenuConfiguration that
// have the zero value. All of the default values are defined in
// package constants.
func (conf *Flags) fillDefault() {
	conf.Path = loadDefault(conf.Path, DefaultDMenuPath)
	conf.BackgroundColor = loadDefault(conf.BackgroundColor, DefaultBackgroundColor)
	conf.TextColor = loadDefault(conf.TextColor, DefaultTextColor)
	conf.SelectedBgColor = loadDefault(conf.SelectedBgColor, DefaultSelectedBackgroundColor)
	conf.SelectedTextColor = loadDefault(conf.SelectedTextColor, DefaultSelectedTextColor)
	conf.Font = loadDefault(conf.Font, DefaultFont)
	conf.WindowID = -1
	conf.Monitor = -1
}

func loadDefault(currentValue, defaultValue string) string {
	if currentValue != "" {
		return currentValue
	}
	return defaultValue
}

func (conf Flags) args() []string {
	args := make([]string, 0, 20)

	if !conf.CaseSensitive {
		args = append(args, "-i")
	}

	if conf.Bottom {
		args = append(args, "-b")
	}

	if conf.Lines > 0 {
		args = append(args, "-l", fmt.Sprint(conf.Lines))
	}

	if conf.Font != "" {
		args = append(args, "-fn", conf.Font)
	}

	if conf.Prompt != "" {
		args = append(args, "-p", conf.Prompt)
	}

	if conf.BackgroundColor != "" {
		args = append(args, "-nb", conf.BackgroundColor)
	}

	if conf.SelectedBgColor != "" {
		args = append(args, "-sb", conf.SelectedBgColor)
	}

	if conf.TextColor != "" {
		args = append(args, "-nf", conf.TextColor)
	}

	if conf.SelectedTextColor != "" {
		args = append(args, "-sf", conf.SelectedTextColor)
	}

	if conf.Monitor > -1 {
		args = append(args, "-m", fmt.Sprint(conf.Monitor))
	}

	if conf.WindowID > -1 {
		args = append(args, "-w", fmt.Sprint(conf.WindowID))
	}

	return args
}
