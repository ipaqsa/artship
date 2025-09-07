package client

import (
	"context"
	"fmt"

	"artship/internal/tools"

	"gopkg.in/yaml.v3"
)

// ImageMeta contains metadata information about an OCI/Docker image
type ImageMeta struct {
	Author       string            `json:"author" yaml:"Author,omitempty"`
	Digest       string            `json:"digest" yaml:"Digest"`
	MediaType    string            `json:"mediaType" yaml:"MediaType"`
	Architecture string            `json:"architecture" yaml:"Architecture"`
	OS           string            `json:"os" yaml:"OS"`
	Size         string            `json:"size" yaml:"Size"`
	Created      string            `json:"created,omitempty" yaml:"Created,omitempty"`
	Env          []string          `json:"env,omitempty" yaml:"Env,omitempty"`
	Cmd          []string          `json:"cmd,omitempty" yaml:"Cmd,omitempty"`
	Entrypoint   []string          `json:"entrypoint,omitempty" yaml:"Entrypoint,omitempty"`
	WorkingDir   string            `json:"workingDir,omitempty" yaml:"WorkingDir,omitempty"`
	User         string            `json:"user,omitempty" yaml:"User,omitempty"`
	Labels       map[string]string `json:"labels,omitempty" yaml:"Labels,omitempty"`
	Annotations  map[string]string `json:"annotations,omitempty" yaml:"Annotations,omitempty"`
	Layers       []LayerMeta       `json:"layers" yaml:"Layers"`
}

// LayerMeta contains metadata about a single layer
type LayerMeta struct {
	Digest      string            `json:"digest" yaml:"Digest"`
	Size        string            `json:"size" yaml:"Size"`
	MediaType   string            `json:"mediaType" json:"MediaType"`
	Annotations map[string]string `json:"annotations,omitempty" yaml:"Annotations,omitempty"`
}

func (m *ImageMeta) String() string {
	marshalled, _ := yaml.Marshal(m)
	return string(marshalled)
}

// GetImageMeta retrieves image metadata like digest, os, size, etc
func (c *Client) GetImageMeta(ctx context.Context, imageRef string) (*ImageMeta, error) {
	if imageRef == "" {
		return nil, fmt.Errorf("no image ref provided")
	}

	img, err := c.image(ctx, imageRef)
	if err != nil {
		return nil, err
	}

	digest, err := img.Digest()
	if err != nil {
		return nil, fmt.Errorf("get image digest: %w", err)
	}

	manifest, err := img.Manifest()
	if err != nil {
		return nil, fmt.Errorf("get image manifest: %w", err)
	}

	config, err := img.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("get image config: %w", err)
	}

	size, err := img.Size()
	if err != nil {
		return nil, fmt.Errorf("get image size: %w", err)
	}

	layers, err := img.Layers()
	if err != nil {
		return nil, fmt.Errorf("get image layers: %w", err)
	}

	var layerMetas []LayerMeta
	for i, layer := range layers {
		layerDigest, err := layer.Digest()
		if err != nil {
			return nil, fmt.Errorf("get layer digest: %w", err)
		}

		layerSize, err := layer.Size()
		if err != nil {
			return nil, fmt.Errorf("get layer size: %w", err)
		}

		layerMediaType, err := layer.MediaType()
		if err != nil {
			return nil, fmt.Errorf("get layer media type: %w", err)
		}

		layerMeta := LayerMeta{
			Digest:    layerDigest.String(),
			Size:      tools.FormatSize(layerSize),
			MediaType: string(layerMediaType),
		}

		// Add layer annotations from manifest if available
		if i < len(manifest.Layers) && len(manifest.Layers[i].Annotations) > 0 {
			layerMeta.Annotations = manifest.Layers[i].Annotations
		}

		layerMetas = append(layerMetas, layerMeta)
	}

	meta := &ImageMeta{
		Author:       config.Author,
		Digest:       digest.String(),
		MediaType:    string(manifest.MediaType),
		Architecture: config.Architecture,
		OS:           config.OS,
		Size:         tools.FormatSize(size),
		Layers:       layerMetas,
	}

	if !config.Created.IsZero() {
		meta.Created = config.Created.Format("2006-01-02T15:04:05Z")
	}

	if len(config.Config.Env) > 0 {
		meta.Env = config.Config.Env
	}

	if len(config.Config.Cmd) > 0 {
		meta.Cmd = config.Config.Cmd
	}

	if len(config.Config.Entrypoint) > 0 {
		meta.Entrypoint = config.Config.Entrypoint
	}

	if config.Config.WorkingDir != "" {
		meta.WorkingDir = config.Config.WorkingDir
	}

	if config.Config.User != "" {
		meta.User = config.Config.User
	}

	if len(config.Config.Labels) > 0 {
		meta.Labels = config.Config.Labels
	}

	if len(manifest.Annotations) > 0 {
		meta.Annotations = manifest.Annotations
	}

	return meta, nil
}
