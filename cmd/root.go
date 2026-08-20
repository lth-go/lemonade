package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/lemonade-command/lemonade/lemon"
)

func NewRootCmd(cfg *lemon.Config) *cobra.Command {
	root := &cobra.Command{
		Use:           "lemonade [options]... SUB_COMMAND [arg]",
		Short:         "Remote clipboard utility over HTTP.",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	pf := root.PersistentFlags()
	pf.StringVar(&cfg.Host, "host", "", "Destination host name. [Client only]")
	pf.IntVar(&cfg.Port, "port", 2489, "TCP port number")
	pf.StringVar(&cfg.Allow, "allow", "0.0.0.0/0,::/0", "Allow IP range. [Server only]")
	pf.StringVar(&cfg.LineEnding, "line-ending", "", "Convert Line Endings (CR/CRLF)")
	pf.StringVar(&cfg.LogLevel, "log-level", "info", "Log level (debug, info, warn, error, critical)")

	// PersistentPreRunE runs before each subcommand. We merge the config
	// file and environment fallbacks here, honoring which flags the user
	// explicitly set on the command line.
	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		changed := changedFlags(cmd)
		conf, err := lemon.LoadConfig()
		if err != nil {
			return err
		}
		cfg.Merge(conf, changed)
		return nil
	}

	root.AddCommand(newCopyCmd(cfg))
	root.AddCommand(newPasteCmd(cfg))
	root.AddCommand(newServerCmd(cfg))
	return root
}

// changedFlags returns the set of flag names the user explicitly passed on
// the command line, across the whole command tree (root + sub).
func changedFlags(cmd *cobra.Command) map[string]bool {
	out := map[string]bool{}
	add := func(fs *pflag.FlagSet) {
		fs.Visit(func(f *pflag.Flag) {
			out[f.Name] = true
		})
	}
	add(cmd.Flags())
	add(cmd.PersistentFlags())
	for p := cmd; p != nil; p = p.Parent() {
		add(p.PersistentFlags())
	}
	return out
}
