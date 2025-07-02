package godmenu

import "fmt"

// Configuration defines how go-DMenu interacts with DMenu. You can either specify these arguments
type Configuration struct {
	Path            string
	BackgroundColor string
	Foreground      string
	Font            string
	Prompt          string
	CaseSensitive   bool
	Sorted          bool
	Bottom          bool
	Lines           int
	Monitor         Optional[int]
	WindowID        Optional[int]
}

func (op *Operation) applyOptions(opts []Option) *Operation {
	for _, opt := range opts {
		opt(op)
	}
	return op
}

type Option func(*Operation)

func MakeDefaultConfiguration() *Configuration  { c := &Configuration{}; c.fillDefault(); return c }
func MakeOperation(s ...string) *Operation      { return &Operation{Selections: s} }
func WithConfiguration(n *Configuration) Option { return func(c *Operation) { *c.DMenu = *n } }
func WithOperation(op *Operation) Option        { return func(c *Operation) { *c = *op } }
func SetSelections(s []string) Option           { return func(c *Operation) { c.Selections = s } }
func AppendSelections(s []string) Option        { return func(c *Operation) { c.extendSelections(s) } }
func UnsetSelections() Option                   { return func(c *Operation) { c.Selections = []string{} } }
func UseStrictMode() Option                     { return func(c *Operation) { c.Strict = true } }
func SetStrict(v bool) Option                   { return func(c *Operation) { c.Strict = v } }
func DMenuPath(p string) Option                 { return func(c *Operation) { c.DMenu.Path = p } }
func DMenuBackground(cl string) Option          { return func(c *Operation) { c.DMenu.BackgroundColor = cl } }
func DMenuTextColor(cl string) Option           { return func(c *Operation) { c.DMenu.Foreground = cl } }
func DMenuCaseSensitive() Option                { return func(c *Operation) { c.DMenu.CaseSensitive = true } }
func DMenuCaseInsensitive() Option              { return func(c *Operation) { c.DMenu.CaseSensitive = false } }
func DMenuPrompt(p string) Option               { return func(c *Operation) { c.DMenu.Prompt = p } }
func DMenuBottom() Option                       { return func(c *Operation) { c.DMenu.Bottom = true } }
func DMenuTop() Option                          { return func(c *Operation) { c.DMenu.Bottom = false } }
func DMenuSorted() Option                       { return func(c *Operation) { c.DMenu.Sorted = true } }
func DMenuUnsorted() Option                     { return func(c *Operation) { c.DMenu.Sorted = false } }
func DMenuLines(n int) Option                   { return func(c *Operation) { c.DMenu.Lines = n } }
func DMenuMonitor(n int) Option                 { return func(c *Operation) { c.DMenu.Monitor.Set(n) } }
func DMenuMonitorUnset() Option                 { return func(c *Operation) { c.DMenu.Monitor.Reset() } }
func DMenuWindowID(n int) Option                { return func(c *Operation) { c.DMenu.WindowID.Set(n) } }
func DMenuWindowIDUnset() Option                { return func(c *Operation) { c.DMenu.WindowID.Reset() } }

const (
	DefaultBackgroundColor = "#000000"
	DefaultForegroundColor = "#ffffff"
	DefaultFount           = "Source Code Pro-12"
	DefaultDmenuPath       = "dmenu"
)

var defaultDmenuConfig Configuration

func init() { defaultDmenuConfig.fillDefault() }

// fillDefault sets any unset fields in the DMenuConfiguration that
// have the zero value. All of the default values are defined in
// package constants.
func (conf *Configuration) fillDefault() {
	if conf.Path == "" {
		conf.Path = DefaultDmenuPath
	}

	if conf.BackgroundColor == "" {
		conf.BackgroundColor = DefaultBackgroundColor
	}

	if conf.Foreground == "" {
		conf.Foreground = DefaultForegroundColor
	}

	if conf.Font == "" {
		conf.Font = DefaultFount
	}
}

func (conf Configuration) resolveArgs() []string {
	args := []string{
		"-nb", conf.BackgroundColor,
		"-nf", conf.Foreground,
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
