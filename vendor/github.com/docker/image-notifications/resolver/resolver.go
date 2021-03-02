package resolver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"

	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/distribution/distribution/v3/reference"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// Resolver fetches the descriptor of an oci index.
type Resolver interface {
	GetDigest(ctx context.Context, ref reference.Named, arch string) (digest.Digest, error)
}

type resolver struct {
	remotesResolver remotes.Resolver
}

// Opt is for configuring the resolver
type Opt func(r *resolver) error

// New creates a new resolver
func New(auth docker.Authorizer) Resolver {
	remotesResolver := docker.NewResolver(docker.ResolverOptions{
		Authorizer: auth,
	})

	r := &resolver{
		remotesResolver: remotesResolver,
	}

	return r
}

func (r *resolver) GetDigest(ctx context.Context, ref reference.Named, arch string) (digest.Digest, error) {
	name, descriptor, err := r.remotesResolver.Resolve(ctx, ref.String())
	if err != nil {
		return "", err
	}

	fetcher, err := r.remotesResolver.Fetcher(ctx, name)
	if err != nil {
		return "", err
	}

	rc, err := fetcher.Fetch(ctx, descriptor)
	if err != nil {
		return "", err
	}

	buffer := &bytes.Buffer{}
	_, err = io.Copy(buffer, rc)
	rc.Close()
	if err != nil {
		return "", err
	}

	switch descriptor.MediaType {
	case images.MediaTypeDockerSchema2Manifest:
		var manifest ocispec.Manifest
		if err := json.Unmarshal(buffer.Bytes(), &manifest); err != nil {
			return "", err
		}
		return manifest.Config.Digest, nil
	case images.MediaTypeDockerSchema2ManifestList, ocispec.MediaTypeImageIndex:
		var idx ocispec.Index
		if err := json.Unmarshal(buffer.Bytes(), &idx); err != nil {
			return "", err
		}

		for _, manifest := range idx.Manifests {
			if manifest.Platform.Architecture == arch {
				return manifest.Digest, nil
			}
		}

		return "", errors.New("not found")
	}

	return "", nil
}
