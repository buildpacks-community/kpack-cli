// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/ghodss/yaml"
	"github.com/google/go-containerregistry/pkg/authn"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/pkg/errors"
	buildv1alpha2 "github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
)

const (
	builderOrderLabel = "io.buildpacks.buildpack.order"
)

type Fetcher interface {
	Fetch(keychain authn.Keychain, src string) (v1.Image, error)
}

// cnbOrderEntry represents the structure of a buildpack order entry in the CNB builder image label
type cnbOrderEntry struct {
	Group []cnbBuildpackRef `json:"group"`
}

// cnbBuildpackRef represents a buildpack reference in the CNB builder image label
type cnbBuildpackRef struct {
	ID       string `json:"id"`
	Version  string `json:"version,omitempty"`
	Optional bool   `json:"optional,omitempty"`
}

func ReadOrder(path string) ([]buildv1alpha2.BuilderOrderEntry, error) {
	var (
		file io.ReadCloser
		err  error
	)

	if path == "-" {
		file = os.Stdin
	} else {
		file, err = os.Open(path)
		if err != nil {
			return nil, err
		}
	}
	defer file.Close()

	buf, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var order []buildv1alpha2.BuilderOrderEntry
	return order, yaml.Unmarshal(buf, &order)
}

func CreateOrder(buildpacks []string) []buildv1alpha2.BuilderOrderEntry {
	group := make([]buildv1alpha2.BuilderBuildpackRef, 0)

	// this regular expression splits out buildpack id and version
	var re = regexp.MustCompile(`(?m)^([^@]+)[@]?(.*)`)

	for _, buildpack := range buildpacks {
		submatch := re.FindStringSubmatch(buildpack)

		id := submatch[1]
		version := submatch[2]

		group = append(group, buildv1alpha2.BuilderBuildpackRef{
			BuildpackRef: corev1alpha1.BuildpackRef{
				BuildpackInfo: corev1alpha1.BuildpackInfo{
					Id:      id,
					Version: version,
				},
			},
		})
	}

	return []buildv1alpha2.BuilderOrderEntry{{Group: group}}
}

func CreateDetectionOrderRow(ref corev1alpha1.BuildpackRef) (string, string) {
	data := fmt.Sprintf("  %s", ref.Id)
	optional := ""

	if ref.Version != "" {
		data = fmt.Sprintf("%s@%s", data, ref.Version)
	}

	if ref.Optional {
		optional = "(Optional)"
	}

	return data, optional
}

func CoreOrderEntryToBuildOrderEntry(order []corev1alpha1.OrderEntry) []buildv1alpha2.BuilderOrderEntry {
	res := make([]buildv1alpha2.BuilderOrderEntry, len(order))
	for i, entry := range order {
		group := make([]buildv1alpha2.BuilderBuildpackRef, len(entry.Group))
		for j, ref := range entry.Group {
			group[j] = buildv1alpha2.BuilderBuildpackRef{
				BuildpackRef: ref,
			}
		}

		res[i] = buildv1alpha2.BuilderOrderEntry{
			Group: group,
		}
	}
	return res
}

// ReadOrderFromImage extracts the buildpack order from a CNB builder image's io.buildpacks.buildpack.order label
func ReadOrderFromImage(keychain authn.Keychain, fetcher Fetcher, imageRef string) ([]buildv1alpha2.BuilderOrderEntry, error) {
	img, err := fetcher.Fetch(keychain, imageRef)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch image %s", imageRef)
	}

	config, err := img.ConfigFile()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read image config")
	}

	orderJSON, ok := config.Config.Labels[builderOrderLabel]
	if !ok {
		return nil, errors.Errorf("image %s does not contain the %s label", imageRef, builderOrderLabel)
	}

	var cnbOrder []cnbOrderEntry
	if err := json.Unmarshal([]byte(orderJSON), &cnbOrder); err != nil {
		return nil, errors.Wrapf(err, "failed to parse %s label", builderOrderLabel)
	}

	// Convert CNB order format to kpack BuilderOrderEntry format
	order := make([]buildv1alpha2.BuilderOrderEntry, len(cnbOrder))
	for i, entry := range cnbOrder {
		group := make([]buildv1alpha2.BuilderBuildpackRef, len(entry.Group))
		for j, ref := range entry.Group {
			group[j] = buildv1alpha2.BuilderBuildpackRef{
				BuildpackRef: corev1alpha1.BuildpackRef{
					BuildpackInfo: corev1alpha1.BuildpackInfo{
						Id:      ref.ID,
						Version: ref.Version,
					},
					Optional: ref.Optional,
				},
			}
		}
		order[i] = buildv1alpha2.BuilderOrderEntry{
			Group: group,
		}
	}

	return order, nil
}
