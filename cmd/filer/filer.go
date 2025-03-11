package filer

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/GBA-BI/tes-filer/cmd/filer/options"
	"github.com/GBA-BI/tes-filer/internal/application"
	"github.com/GBA-BI/tes-filer/internal/infra/repo"
	"github.com/GBA-BI/tes-filer/pkg/log"
	"github.com/GBA-BI/tes-filer/pkg/version"
)

func newFilerCommand(ctx context.Context, opts *options.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "filer input/output/all",
		Short: "filer",
		Long: `vetes filer
`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.AppFiler.Mode = args[0]
			if err := opts.Validate(); err != nil {
				return err
			}

			logger, err := log.GetLogger(opts.Log)
			if err != nil {
				return err
			}
			defer log.Close()

			version.PrintVersionOrContinue()

			if err := run(ctx, opts, logger); err != nil {
				logger.Errorf("run error: %v", err)
				return err
			}
			return nil
		},
		Args: cobra.ExactArgs(1),
	}
}

func run(ctx context.Context, opts *options.Options, logger log.Logger) error {
	filerRepo, err := repo.NewFilerRepo(opts.RepoConfig, logger)
	if err != nil {
		return err
	}
	cmd, err := application.NewTransputCmd(opts.AppFiler, filerRepo)
	if err != nil {
		return err
	}
	return cmd.Transput(ctx)
}

func NewFilerCommand(ctx context.Context) *cobra.Command {
	opts := options.NewFromENV()

	cmd := newFilerCommand(ctx, opts)
	opts.AddFlags(cmd.Flags())
	version.AddFlags(cmd.Flags())

	return cmd
}
