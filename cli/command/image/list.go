package image

import (
	"context"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/distribution/distribution/v3/reference"
	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/formatter"
	"github.com/docker/cli/opts"
	"github.com/docker/docker/api/types"
	"github.com/docker/image-notifications/client"
	"github.com/docker/image-notifications/resolver"
	"github.com/spf13/cobra"
	"strings"
)

type imagesOptions struct {
	matchName string

	quiet       bool
	all         bool
	noTrunc     bool
	showDigests bool
	format      string
	filter      opts.FilterOpt
}

// NewImagesCommand creates a new `docker images` command
func NewImagesCommand(dockerCli command.Cli) *cobra.Command {
	options := imagesOptions{filter: opts.NewFilterOpt()}

	cmd := &cobra.Command{
		Use:   "images [OPTIONS] [REPOSITORY[:TAG]]",
		Short: "List images",
		Args:  cli.RequiresMaxArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				options.matchName = args[0]
			}
			return runImages(dockerCli, options)
		},
	}

	flags := cmd.Flags()

	flags.BoolVarP(&options.quiet, "quiet", "q", false, "Only show image IDs")
	flags.BoolVarP(&options.all, "all", "a", false, "Show all images (default hides intermediate images)")
	flags.BoolVar(&options.noTrunc, "no-trunc", false, "Don't truncate output")
	flags.BoolVar(&options.showDigests, "digests", false, "Show digests")
	flags.StringVar(&options.format, "format", "", "Pretty-print images using a Go template")
	flags.VarP(&options.filter, "filter", "f", "Filter output based on conditions provided")

	return cmd
}

func newListCommand(dockerCli command.Cli) *cobra.Command {
	cmd := *NewImagesCommand(dockerCli)
	cmd.Aliases = []string{"list"}
	cmd.Use = "ls [OPTIONS] [REPOSITORY[:TAG]]"
	return &cmd
}

func runImages(dockerCli command.Cli, options imagesOptions) error {
	ctx := context.Background()

	filters := options.filter.Value()
	if options.matchName != "" {
		filters.Add("reference", options.matchName)
	}

	listOptions := types.ImageListOptions{
		All:     options.all,
		Filters: filters,
	}

	images, err := dockerCli.Client().ImageList(ctx, listOptions)
	if err != nil {
		return err
	}

	format := options.format
	if len(format) == 0 {
		if len(dockerCli.ConfigFile().ImagesFormat) > 0 && !options.quiet {
			format = dockerCli.ConfigFile().ImagesFormat
		} else {
			format = formatter.TableFormatKey
		}
	}

	imageCtx := formatter.ImageContext{
		Context: formatter.Context{
			Output: dockerCli.Out(),
			Format: formatter.NewImageFormat(format, options.quiet, options.showDigests),
			Trunc:  !options.noTrunc,
		},
		Digest: options.showDigests,
	}
	var is []formatter.ImageSummary

	r := resolver.New(docker.NewAuthorizer(nil, func(hostName string) (string, string, error) {
		if hostName == "docker.io" {
			hostName = "https://index.docker.io/v1/"
		}
		a, err := dockerCli.ConfigFile().GetAuthConfig(hostName)
		if err != nil {
			return "", "", err
		}
		if a.IdentityToken != "" {
			return "", a.IdentityToken, nil
		}
		return a.Username, a.Password, nil
	}))

	c, err := client.New()
	if err != nil {
		return err
	}

	for _, i := range images {
		newTag := newestTag(ctx,dockerCli, r, c, i)

		is = append(is, formatter.ImageSummary{
			ImageSummary: i,
			Newest:       newTag,
		})
	}
	return formatter.ImageWrite(imageCtx, is)
}

func newestTag(ctx context.Context,dockerCli command.Cli, r resolver.Resolver, c client.Client, image types.ImageSummary) string {
	imageInspect, _, err := dockerCli.Client().ImageInspectWithRaw(ctx, image.ID)
	if err != nil {
		return ""
	}
	if len(imageInspect.RepoDigests) == 0 {
		return ""
	}

	ref, err := parseRef(imageInspect.RepoTags[0])
	if err != nil {
		return ""
	}

	if strings.Contains(reference.FamiliarName(ref), "/") {
		return ""
	}

	digest, err := r.GetDigest(ctx, ref, imageInspect.Architecture)
	if err != nil {
		return ""
	}

	imageInfoResponse, err := c.GetImageInfo(ctx, digest.String())
	if err != nil{
		return ""
	}
	imageInfo := *imageInfoResponse.ImageInfos[0]
	return imageInfo.Tags[0]
}

func parseRef(s string) (reference.Named, error) {
	ref, err := reference.ParseNormalizedNamed(s)
	if err != nil {
		return nil, err
	}

	ref = reference.TagNameOnly(ref)

	return ref, nil
}
