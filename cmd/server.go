package cmd

import (
	"github.com/spf13/cobra"

	"github.com/lemonade-command/lemonade/lemon"
	"github.com/lemonade-command/lemonade/server"
)

func newServerCmd(cfg *lemon.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "server",
		Short: "Start lemonade server.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := server.New(server.Options{
				Allow:      cfg.Allow,
				Port:       cfg.Port,
				LineEnding: cfg.LineEnding,
				Logger:     lemon.NewLogger(*cfg, cmd.ErrOrStderr()),
			})
			if err != nil {
				return err
			}
			return s.Serve()
		},
	}
}
