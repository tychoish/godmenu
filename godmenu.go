// Package godmenu is a wrapper around the dmenu program to easily
// interact with dmenu from Go applications.
package godmenu

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
)

var (
	ErrSelectionMissing     = errors.New("missing selection")
	ErrSelectionUnknown     = errors.New("unknown selection")
	ErrDmenuFailure         = errors.New("dmenu failed")
	ErrConfigurationInvalid = errors.New("invalid configuration")
	ErrSelectionRejected    = errors.New("selection rejected")
	ErrConfirmation         = errors.New("selection confirmation")
)

// Do shells out to dmenu with the given options and returns the
// selected result. If there was a problem with the command, the error
// is returned.
func Do(ctx context.Context, opts Options) (string, error) {
	selections, err := opts.validate()
	if err != nil {
		return "", err
	}

	cmd := exec.CommandContext(ctx, opts.Flags.Path, opts.Flags.args()...)

	cmd.Stdin = bytes.NewBuffer(selections.rendered(opts.Sorted))

	out, err := selections.processOutput(cmd.CombinedOutput())
	if err != nil {
		return "", err
	}

	if opts.ConfirmSubstitution && !selections.check(out) {
		confirm := opts

		confirm.with(Items(out, "accept", "reject"))

		confirmOut, err := Do(ctx, confirm)
		switch {
		case err != nil:
			return "", fmt.Errorf("%w: during %w", err, ErrConfirmation)
		case confirmOut == out:
			return out, nil
		case confirmOut == "accept":
			return out, nil
		case confirmOut == "reject":
			return "", fmt.Errorf("during %w, %q was rejected: %w", ErrConfirmation, out, ErrSelectionRejected)
		case opts.RequireMatch:
			return out, fmt.Errorf("modified original selection %q, but selections must be an exact match (%q): %w during  %w", out, confirmOut, ErrSelectionUnknown, ErrConfirmation)
		case confirmOut == "":
			return "", fmt.Errorf("during %w: %w", ErrConfirmation, ErrSelectionMissing)
		default:
			return confirmOut, nil
		}
	}

	return out, nil
}

// Run calls Do but takes its configuration as Args arguments.
func Run(ctx context.Context, args ...Arg) (string, error) { return Do(ctx, newop().apply(args).ref()) }
func MakeOptions(s ...string) *Options                     { return newop().extendSelections(s).flags() }
func ResolveOptions(arg ...Arg) *Options                   { return newop().apply(arg) }
func DefaultFlags() *Flags                                 { c := defaultDmenuConfig; return &c }
func WithFlags(n *Flags) Arg                               { return func(o *Options) { o.Flags = n } }
func WithOptions(override *Options) Arg                    { return func(o *Options) { *o = *override } }
func WithSelections(s ...string) Arg                       { return ExtendSelections(s) }
func Items(s ...string) Arg                                { return ExtendSelections(s) }
func SetMatchRequirement(state bool) Arg                   { return func(o *Options) { o.RequireMatch = state } }
func RequireMatch() Arg                                    { return SetMatchRequirement(true) }
func AllowMatch() Arg                                      { return SetMatchRequirement(false) }
func SetConfirmSubstituion(state bool) Arg                 { return func(o *Options) { o.ConfirmSubstitution = state } }
func ConfirmSubstituion() Arg                              { return SetConfirmSubstituion(true) }
func SkipConfirmSubstitution() Arg                         { return SetConfirmSubstituion(false) }
func SetUniqueSelectionPolicy(state bool) Arg              { return func(o *Options) { o.AllowDuplicates = state } }
func AllowDuplicateSelections() Arg                        { return SetUniqueSelectionPolicy(true) }
func RequireUniqueSelections() Arg                         { return SetUniqueSelectionPolicy(false) }
func Selections(s ...string) Arg                           { return ExtendSelections(s) }
func Prompt(p string) Arg                                  { return MenuPrompt(p) }
func Sorted() Arg                                          { return func(o *Options) { o.Sorted = true } }
func ExtendSelections(s []string) Arg                      { return func(o *Options) { o.extendSelections(s) } }
func SetSelections(s []string) Arg                         { return func(o *Options) { o.Selections = s } }
func ResetSelections() Arg                                 { return func(o *Options) { o.Selections = []string{} } }
func Unsorted() Arg                                        { return func(o *Options) { o.Sorted = false } }
func TextColor(c string) Arg                               { return func(o *Options) { o.Flags.TextColor = c } }
func BackgroundColor(c string) Arg                         { return func(o *Options) { o.Flags.BackgroundColor = c } }
func SelectedText(c string) Arg                            { return func(o *Options) { o.Flags.SelectedTextColor = c } }
func SelectedBgColor(c string) Arg                         { return func(o *Options) { o.Flags.SelectedBgColor = c } }
func CaseSensitive() Arg                                   { return func(o *Options) { o.Flags.CaseSensitive = true } }
func CaseInsensitive() Arg                                 { return func(o *Options) { o.Flags.CaseSensitive = false } }
func DMenuPath(p string) Arg                               { return func(o *Options) { o.Flags.Path = p } }
func MenuPrompt(p string) Arg                              { return func(o *Options) { o.Flags.Prompt = p } }
func MenuBottom() Arg                                      { return func(o *Options) { o.Flags.Bottom = true } }
func MenuTop() Arg                                         { return func(o *Options) { o.Flags.Bottom = false } }
func MenuLines(n int) Arg                                  { return func(o *Options) { o.Flags.Lines = n } }
func MenuMonitor(n int) Arg                                { return func(o *Options) { o.Flags.Monitor = n } }
func MenuMonitorUnset() Arg                                { return func(o *Options) { o.Flags.Monitor = -1 } }
func MenuWindowID(n int) Arg                               { return func(o *Options) { o.Flags.WindowID = n } }
func MenuWindowIDUnset() Arg                               { return func(o *Options) { o.Flags.WindowID = -1 } }
