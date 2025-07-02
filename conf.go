package godmenu

import "fmt"

// Flags defines how go-DMenu interacts with DMenu. You can either specify these arguments
type Flags struct {
	Path               string
	BackgroundColor    string
	TextColor          string
	SelectedBackground string
	SelectedTextColor  string
	Font               string
	Prompt             string
	CaseSensitive      bool
	Bottom             bool
	Lines              int
	Monitor            int
	WindowID           int
}

type Option func(*Configuration)

func MakeDefaultConfiguration() *Flags         { c := &Flags{}; c.fillDefault(); return c }
func MakeOperation(s ...string) *Configuration { return &Configuration{Selections: s} }
func WithConfiguration(n *Flags) Option        { return func(o *Configuration) { *o.Flags = *n } }
func WithOperation(op *Configuration) Option   { return func(o *Configuration) { *o = *op } }
func SetSelections(s []string) Option          { return func(o *Configuration) { o.Selections = s } }
func AppendSelections(s []string) Option       { return func(o *Configuration) { o.extendSelections(s) } }
func UnsetSelections() Option                  { return func(o *Configuration) { o.Selections = []string{} } }
func TextColor(c string) Option                { return func(o *Configuration) { o.Flags.TextColor = c } }
func BackgroundColor(c string) Option          { return func(o *Configuration) { o.Flags.BackgroundColor = c } }
func SelectedText(c string) Option             { return func(o *Configuration) { o.Flags.SelectedTextColor = c } }
func SelectedBackground(c string) Option {
	return func(o *Configuration) { o.Flags.SelectedBackground = c }
}
func CaseSensitive() Option       { return func(o *Configuration) { o.Flags.CaseSensitive = true } }
func CaseInsensitive() Option     { return func(o *Configuration) { o.Flags.CaseSensitive = false } }
func DMenuPath(p string) Option   { return func(o *Configuration) { o.Flags.Path = p } }
func DMenuPrompt(p string) Option { return func(o *Configuration) { o.Flags.Prompt = p } }
func DMenuBottom() Option         { return func(o *Configuration) { o.Flags.Bottom = true } }
func DMenuTop() Option            { return func(o *Configuration) { o.Flags.Bottom = false } }
func Sorted() Option              { return func(o *Configuration) { o.Sorted = true } }
func Unsorted() Option            { return func(o *Configuration) { o.Sorted = false } }
func DMenuLines(n int) Option     { return func(o *Configuration) { o.Flags.Lines = n } }
func DMenuMonitor(n int) Option   { return func(o *Configuration) { o.Flags.Monitor = n } }
func DMenuMonitorUnset() Option   { return func(o *Configuration) { o.Flags.Monitor = -1 } }
func DMenuWindowID(n int) Option  { return func(o *Configuration) { o.Flags.WindowID = n } }
func DMenuWindowIDUnset() Option  { return func(o *Configuration) { o.Flags.WindowID = -1 } }

const (
	DefaultDMenuPath               = "dmenu"
	DefaultFont                    = "Source Code Pro-12"
	DefaultBackgroundColor         = "#000000"
	DefaultTextColor               = "#ffffff"
	DefaultSelectedBackgroundColor = "#005577"
	DefaultSelectedTextColor       = "#ffffff"
)

var defaultDmenuConfig Flags

func init() { defaultDmenuConfig.fillDefault() }

// fillDefault sets any unset fields in the DMenuConfiguration that
// have the zero value. All of the default values are defined in
// package constants.
func (conf *Flags) fillDefault() {
	conf.Path = loadDefault(conf.Path, DefaultDMenuPath)
	conf.BackgroundColor = loadDefault(conf.BackgroundColor, DefaultBackgroundColor)
	conf.TextColor = loadDefault(conf.TextColor, DefaultTextColor)
	conf.SelectedBackground = loadDefault(conf.SelectedBackground, DefaultSelectedBackgroundColor)
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
	args := []string{
		"-nb", conf.BackgroundColor,
		"-sb", conf.SelectedBackground,
		"-nf", conf.TextColor,
		"-sf", conf.SelectedTextColor,
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

	if conf.Monitor > -1 {
		args = append(args, "-m", fmt.Sprint(conf.Monitor))
	}

	if conf.WindowID > -1 {
		args = append(args, "-w", fmt.Sprint(conf.WindowID))
	}

	return args
}
