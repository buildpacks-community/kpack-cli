// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package build_test

import (
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/commands/build"
	registryfakes "github.com/vmware-tanzu/kpack-cli/pkg/registry/fakes"
	"github.com/vmware-tanzu/kpack-cli/pkg/testhelpers"
)

func TestBuildStatusCommand(t *testing.T) {
	spec.Run(t, "TestBuildStatusCommand", testBuildStatusCommand)
}

func testBuildStatusCommand(t *testing.T, when spec.G, it spec.S) {
	const (
		image                       = "test-image"
		defaultNamespace            = "some-default-namespace"
		expectedOutputForMostRecent = `Image:     repo.com/image-3:tag
Status:    BUILDING
Reason:    TRIGGER

Started:     0001-01-01 05:00:00
Finished:    --

Pod Name:    pod-three

Builder:      some-repo.com/my-builder
Run Image:    some-repo.com/run-image

Source:    Local Source

BUILDPACK ID    BUILDPACK VERSION    HOMEPAGE
bp-id-1         bp-version-1         mysupercoolsite.com
bp-id-2         bp-version-2         mysupercoolsite2.com

`
		expectedOutputForBuildNumber = `Image:     repo.com/image-1:tag
Status:    SUCCESS
Reason:    CONFIG

Started:     0001-01-01 00:00:00
Finished:    0001-01-01 00:00:00

Pod Name:    pod-one

Builder:      some-repo.com/my-builder
Run Image:    some-repo.com/run-image

Source:    Local Source

BUILDPACK ID    BUILDPACK VERSION    HOMEPAGE
bp-id-1         bp-version-1         mysupercoolsite.com
bp-id-2         bp-version-2         mysupercoolsite2.com

`
	)

	cmdFunc := func(clientSet *fake.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeKpackProvider(clientSet, defaultNamespace)

		fakeFetcher := registryfakes.Fetcher{}
		fakeFetcher.AddImage("repo.com/image-1:tag", registryfakes.NewFakeLabeledImage("io.buildpacks.build.metadata", "{\"bom\":{\"some\":\"metadata\"}}", "some-digest"))

		fakeRegistryUtilProvider := &registryfakes.UtilProvider{
			FakeFetcher: &fakeFetcher,
		}
		return build.NewStatusCommand(clientSetProvider, fakeRegistryUtilProvider)
	}

	when("getting build status", func() {
		when("in the default namespace", func() {
			builds := testhelpers.BuildsToRuntimeObjs(testhelpers.MakeTestBuilds(image, defaultNamespace))

			when("the build exists", func() {
				when("the build flag is provided", func() {
					it("shows the build status", func() {
						testhelpers.CommandTest{
							Objects:        builds,
							Args:           []string{image, "-b", "1"},
							ExpectedOutput: expectedOutputForBuildNumber,
						}.TestKpack(t, cmdFunc)
					})
				})

				when("the build flag is not provided", func() {
					it("shows the build status of the most recent build", func() {
						testhelpers.CommandTest{
							Objects:        builds,
							Args:           []string{image},
							ExpectedOutput: expectedOutputForMostRecent,
						}.TestKpack(t, cmdFunc)
					})
				})
			})

			when("the build does not exist", func() {
				when("the build flag is provided", func() {
					it("prints an appropriate message", func() {
						testhelpers.CommandTest{
							Objects:             builds,
							Args:                []string{image, "-b", "123"},
							ExpectErr:           true,
							ExpectedErrorOutput: "Error: build \"123\" not found\n",
						}.TestKpack(t, cmdFunc)
					})
				})

				when("the build flag was not provided", func() {
					it("prints an appropriate message", func() {
						testhelpers.CommandTest{
							Args:                []string{image},
							ExpectErr:           true,
							ExpectedErrorOutput: "Error: no builds found\n",
						}.TestKpack(t, cmdFunc)
					})
				})
			})
		})

		when("in a given namespace", func() {
			const namespace = "some-namespace"
			builds := testhelpers.BuildsToRuntimeObjs(testhelpers.MakeTestBuilds(image, namespace))

			when("the build exists", func() {
				when("the build flag is provided", func() {
					it("gets the build status", func() {
						testhelpers.CommandTest{
							Objects:        builds,
							Args:           []string{image, "-b", "1", "-n", namespace},
							ExpectedOutput: expectedOutputForBuildNumber,
						}.TestKpack(t, cmdFunc)
					})
				})

				when("the build flag is not provided", func() {
					it("shows the build status of the most recent build", func() {
						testhelpers.CommandTest{
							Objects:        builds,
							Args:           []string{image, "-n", namespace},
							ExpectedOutput: expectedOutputForMostRecent,
						}.TestKpack(t, cmdFunc)
					})
				})
			})

			when("the build does not exist", func() {
				when("the build flag is provided", func() {
					it("prints an appropriate message", func() {
						testhelpers.CommandTest{
							Objects:             builds,
							Args:                []string{image, "-b", "123", "-n", namespace},
							ExpectErr:           true,
							ExpectedErrorOutput: "Error: build \"123\" not found\n",
						}.TestKpack(t, cmdFunc)
					})
				})

				when("the build flag was not provided", func() {
					it("prints an appropriate message", func() {
						testhelpers.CommandTest{
							Args:                []string{image, "-n", namespace},
							ExpectErr:           true,
							ExpectedErrorOutput: "Error: no builds found\n",
						}.TestKpack(t, cmdFunc)
					})
				})
			})
		})

		when("build status returns a reason and message", func() {
			it("displays status reason and status message", func() {
				expectedOutput := `Image:             repo.com/image-3:tag
Status:            BUILDING
Reason:            TRIGGER
Status Reason:     some-reason
Status Message:    some-message

Started:     0001-01-01 05:00:00
Finished:    --

Pod Name:    some-pod

Builder:      some-repo.com/my-builder
Run Image:    some-repo.com/run-image

Source:    Local Source

BUILDPACK ID    BUILDPACK VERSION    HOMEPAGE
bp-id-1         bp-version-1         mysupercoolsite.com
bp-id-2         bp-version-2         mysupercoolsite2.com

`
				bld := &v1alpha2.Build{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "bld-three",
						Namespace:         "some-default-namespace",
						CreationTimestamp: metav1.Time{Time: time.Time{}.Add(5 * time.Hour)},
						Labels: map[string]string{
							v1alpha2.ImageLabel:       image,
							v1alpha2.BuildNumberLabel: "3",
						},
						Annotations: map[string]string{
							v1alpha2.BuildReasonAnnotation: "TRIGGER",
						},
					},
					Spec: v1alpha2.BuildSpec{
						Builder: corev1alpha1.BuildBuilderSpec{
							Image: "some-repo.com/my-builder",
						},
					},
					Status: v1alpha2.BuildStatus{
						Status: corev1alpha1.Status{
							Conditions: corev1alpha1.Conditions{
								{
									Type:    corev1alpha1.ConditionSucceeded,
									Status:  corev1.ConditionUnknown,
									Reason:  "some-reason",
									Message: "some-message",
								},
							},
						},
						BuildMetadata: corev1alpha1.BuildpackMetadataList{
							{
								Id:       "bp-id-1",
								Version:  "bp-version-1",
								Homepage: "mysupercoolsite.com",
							},
							{
								Id:       "bp-id-2",
								Version:  "bp-version-2",
								Homepage: "mysupercoolsite2.com",
							},
						},
						Stack: corev1alpha1.BuildStack{
							RunImage: "some-repo.com/run-image",
						},
						LatestImage: "repo.com/image-3:tag",
						PodName:     "some-pod",
					},
				}
				testhelpers.CommandTest{
					Objects:        []runtime.Object{bld},
					Args:           []string{image},
					ExpectedOutput: expectedOutput,
				}.TestKpack(t, cmdFunc)
			})

			it("does not display when the condition is empty", func() {
				bld := &v1alpha2.Build{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "bld-three",
						Namespace:         "some-default-namespace",
						CreationTimestamp: metav1.Time{Time: time.Time{}.Add(5 * time.Hour)},
						Labels: map[string]string{
							v1alpha2.ImageLabel:       image,
							v1alpha2.BuildNumberLabel: "3",
						},
						Annotations: map[string]string{
							v1alpha2.BuildReasonAnnotation: "TRIGGER",
						},
					},
					Spec: v1alpha2.BuildSpec{
						Builder: corev1alpha1.BuildBuilderSpec{
							Image: "some-repo.com/my-builder",
						},
					},
					Status: v1alpha2.BuildStatus{
						Status: corev1alpha1.Status{
							Conditions: corev1alpha1.Conditions{},
						},
						BuildMetadata: corev1alpha1.BuildpackMetadataList{
							{
								Id:       "bp-id-1",
								Version:  "bp-version-1",
								Homepage: "mysupercoolsite.com",
							},
							{
								Id:       "bp-id-2",
								Version:  "bp-version-2",
								Homepage: "mysupercoolsite2.com",
							},
						},
						Stack: corev1alpha1.BuildStack{
							RunImage: "some-repo.com/run-image",
						},
						LatestImage: "repo.com/image-3:tag",
						PodName:     "pod-three",
					},
				}
				testhelpers.CommandTest{
					Objects:        []runtime.Object{bld},
					Args:           []string{image},
					ExpectedOutput: expectedOutputForMostRecent,
				}.TestKpack(t, cmdFunc)
			})

			when("changes are available on a build", func() {
				outputTemplateStr := `Image:     repo.com/image-3:tag
Status:    BUILDING
Reason:    {{.Reason}}
           {{.Change}}

Started:     0001-01-01 05:00:00
Finished:    --

Pod Name:    some-pod

Builder:      some-repo.com/my-builder
Run Image:    some-repo.com/run-image

Source:    Local Source

BUILDPACK ID    BUILDPACK VERSION    HOMEPAGE
bp-id-1         bp-version-1         mysupercoolsite.com
bp-id-2         bp-version-2         mysupercoolsite2.com

`

				og := newOutputGenerator(t, outputTemplateStr)
				diffBuilder := testhelpers.NewDiffBuilder(t)
				diffPadding := strings.Repeat(" ", len("Reason:")+commands.StatusWriterPadding)

				bld := &v1alpha2.Build{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "bld-three",
						Namespace:         "some-default-namespace",
						CreationTimestamp: metav1.Time{Time: time.Time{}.Add(5 * time.Hour)},
						Labels: map[string]string{
							v1alpha2.ImageLabel:       image,
							v1alpha2.BuildNumberLabel: "3",
						},
						Annotations: map[string]string{
							v1alpha2.BuildReasonAnnotation: "TRIGGER",
						},
					},
					Spec: v1alpha2.BuildSpec{
						Builder: corev1alpha1.BuildBuilderSpec{
							Image: "some-repo.com/my-builder",
						},
					},
					Status: v1alpha2.BuildStatus{
						Status: corev1alpha1.Status{
							Conditions: corev1alpha1.Conditions{},
						},
						BuildMetadata: corev1alpha1.BuildpackMetadataList{
							{
								Id:       "bp-id-1",
								Version:  "bp-version-1",
								Homepage: "mysupercoolsite.com",
							},
							{
								Id:       "bp-id-2",
								Version:  "bp-version-2",
								Homepage: "mysupercoolsite2.com",
							},
						},
						Stack: corev1alpha1.BuildStack{
							RunImage: "some-repo.com/run-image",
						},
						LatestImage: "repo.com/image-3:tag",
						PodName:     "some-pod",
					},
				}

				it("generates reason from the changes string", func() {
					bld.Annotations[v1alpha2.BuildReasonAnnotation] = "ignored-reason"
					bld.Annotations[v1alpha2.BuildChangesAnnotation] = testhelpers.CompactJSON(`
[
  {
    "reason": "expected-reason",
    "new": "new-change"
  }
]`)

					change := diffBuilder.New("new-change").Out()
					expectedOutput := og.output(t, "expected-reason", change)

					testhelpers.CommandTest{
						Objects:        []runtime.Object{bld},
						Args:           []string{image},
						ExpectedOutput: expectedOutput,
					}.TestKpack(t, cmdFunc)
				})

				when("single change", func() {
					when("TRIGGER", func() {
						it.Before(func() {
							bld.Annotations[v1alpha2.BuildChangesAnnotation] = testhelpers.CompactJSON(`
[
  {
    "reason": "TRIGGER",
    "new": "A new build was manually triggered on Fri, 20 Nov 2020 15:38:15 -0500"
  }
]`)
						})

						it("displays the correct reason and changes diff", func() {
							change := diffBuilder.New("A new build was manually triggered on Fri, 20 Nov 2020 15:38:15 -0500").Out()
							expectedOutput := og.output(t, "TRIGGER", change)

							testhelpers.CommandTest{
								Objects:        []runtime.Object{bld},
								Args:           []string{image},
								ExpectedOutput: expectedOutput,
							}.TestKpack(t, cmdFunc)
						})
					})

					when("COMMIT", func() {
						it.Before(func() {
							bld.Annotations[v1alpha2.BuildChangesAnnotation] = testhelpers.CompactJSON(`
[
  {
    "reason": "COMMIT",
    "old": "old-revision",
    "new": "new-revision"
  }
]`)
						})

						it("displays the correct reason and changes diff", func() {
							change := diffBuilder.SetPrefix("").
								Old("old-revision").SetPrefix(diffPadding).
								New("new-revision").Out()
							expectedOutput := og.output(t, "COMMIT", change)

							testhelpers.CommandTest{
								Objects:        []runtime.Object{bld},
								Args:           []string{image},
								ExpectedOutput: expectedOutput,
							}.TestKpack(t, cmdFunc)
						})
					})

					when("CONFIG", func() {
						it.Before(func() {
							bld.Annotations[v1alpha2.BuildChangesAnnotation] = testhelpers.CompactJSON(`
[
  {
    "reason": "CONFIG",
    "old": {
      "env": [
        {
          "name": "env-var-name",
          "value": "env-var-value"
        },
        {
          "name": "another-env-var-name",
          "value": "another-env-var-value"
        }
      ],
      "resources": {
        "limits": {
          "cpu": "500m",
          "memory": "2G"
        },
        "requests": {
          "cpu": "100m",
          "memory": "512M"
        }
      },
      "source": {
        "git": {
          "url": "some-git-url",
          "revision": "some-git-revision"
        },
        "subPath": "some-sub-path"
      }
    },
    "new": {
      "env": [
        {
          "name": "new-env-var-name",
          "value": "new-env-var-value"
        },
        {
          "name": "another-env-var-name",
          "value": "another-env-var-value"
        }
      ],
      "resources": {
        "limits": {
          "cpu": "300m",
          "memory": "1G"
        },
        "requests": {
          "cpu": "200m",
          "memory": "512M"
        }
      },
      "bindings": [
        {
          "name": "binding-name",
          "metadataRef": {
            "name": "some-metadata-ref"
          },
          "secretRef": {
            "name": "some-secret-ref"
          }
        }
      ],
      "source": {
        "blob": {
          "url": "some-blob-url"
        }
      }
    }
  }
]`)
						})

						it("displays the correct reason and changes diff", func() {
							change := diffBuilder.SetPrefix("").
								New("bindings:").SetPrefix(diffPadding).
								New("- metadataRef:").
								New("    name: some-metadata-ref").
								New("  name: binding-name").
								New("  secretRef:").
								New("    name: some-secret-ref").
								NoD("env:").
								Old("- name: env-var-name").
								Old("  value: env-var-value").
								New("- name: new-env-var-name").
								New("  value: new-env-var-value").
								NoD("- name: another-env-var-name").
								NoD("  value: another-env-var-value").
								NoD("resources:").
								NoD("  limits:").
								Old("    cpu: 500m").
								Old("    memory: 2G").
								New("    cpu: 300m").
								New("    memory: 1G").
								NoD("  requests:").
								Old("    cpu: 100m").
								New("    cpu: 200m").
								NoD("    memory: 512M").
								NoD("source:").
								Old("  git:").
								Old("    revision: some-git-revision").
								Old("    url: some-git-url").
								Old("  subPath: some-sub-path").
								New("  blob:").
								New("    url: some-blob-url").Out()
							expectedOutput := og.output(t, "CONFIG", change)

							testhelpers.CommandTest{
								Objects:        []runtime.Object{bld},
								Args:           []string{image},
								ExpectedOutput: expectedOutput,
							}.TestKpack(t, cmdFunc)
						})
					})

					when("BUILDPACK", func() {
						it.Before(func() {
							bld.Annotations[v1alpha2.BuildChangesAnnotation] = testhelpers.CompactJSON(`
[
  {
    "reason": "BUILDPACK",
    "old": [
      {
        "id": "another-buildpack-id",
        "version": "another-buildpack-old-version"
      },
      {
        "id": "some-buildpack-id",
        "version": "some-buildpack-old-version"
      }
    ],
    "new": [
      {
        "id": "some-buildpack-id",
        "version": "some-buildpack-new-version"
      }
    ]
  }
]`)
						})

						it("displays the correct reason and changes diff", func() {
							change := diffBuilder.SetPrefix("").
								Old("- id: another-buildpack-id").SetPrefix(diffPadding).
								Old("  version: another-buildpack-old-version").
								NoD("- id: some-buildpack-id").
								Old("  version: some-buildpack-old-version").
								New("  version: some-buildpack-new-version").Out()
							expectedOutput := og.output(t, "BUILDPACK", change)

							testhelpers.CommandTest{
								Objects:        []runtime.Object{bld},
								Args:           []string{image},
								ExpectedOutput: expectedOutput,
							}.TestKpack(t, cmdFunc)
						})
					})

					when("STACK", func() {
						it.Before(func() {
							bld.Annotations[v1alpha2.BuildChangesAnnotation] = testhelpers.CompactJSON(`
[
  {
    "reason": "STACK",
    "old": "sha256:87302783be0a0cab9fde5b68c9954b7e9150ca0d514ba542e9810c3c6f2984ad",
    "new": "sha256:87302783be0a0cab9fde5b68c9954b7e9150ca0d514ba542e9810c3c6f2984ae"
  }
]`)
						})

						it("displays the correct reason and changes diff", func() {
							change := diffBuilder.SetPrefix("").
								Old("sha256:87302783be0a0cab9fde5b68c9954b7e9150ca0d514ba542e9810c3c6f2984ad").SetPrefix(diffPadding).
								New("sha256:87302783be0a0cab9fde5b68c9954b7e9150ca0d514ba542e9810c3c6f2984ae").Out()
							expectedOutput := og.output(t, "STACK", change)

							testhelpers.CommandTest{
								Objects:        []runtime.Object{bld},
								Args:           []string{image},
								ExpectedOutput: expectedOutput,
							}.TestKpack(t, cmdFunc)
						})
					})
				})

				when("multiple changes", func() {
					it.Before(func() {
						bld.Annotations[v1alpha2.BuildChangesAnnotation] = testhelpers.CompactJSON(`
[
  {
    "reason": "TRIGGER",
    "old": "",
    "new": "A new build was manually triggered on Fri, 20 Nov 2020 15:38:15 -0500"
  },
  {
    "reason": "COMMIT",
    "old": "old-revision",
    "new": "new-revision"
  },
  {
    "reason": "CONFIG",
    "old": {
      "env": [
        {
          "name": "env-var-name",
          "value": "env-var-value"
        },
        {
          "name": "another-env-var-name",
          "value": "another-env-var-value"
        }
      ],
      "resources": {
        "limits": {
          "cpu": "500m",
          "memory": "2G"
        },
        "requests": {
          "cpu": "100m",
          "memory": "512M"
        }
      },
      "source": {
        "git": {
          "url": "some-git-url",
          "revision": "some-git-revision"
        },
        "subPath": "some-sub-path"
      }
    },
    "new": {
      "env": [
        {
          "name": "new-env-var-name",
          "value": "new-env-var-value"
        },
        {
          "name": "another-env-var-name",
          "value": "another-env-var-value"
        }
      ],
      "resources": {
        "limits": {
          "cpu": "300m",
          "memory": "1G"
        },
        "requests": {
          "cpu": "200m",
          "memory": "512M"
        }
      },
      "bindings": [
        {
          "name": "binding-name",
          "metadataRef": {
            "name": "some-metadata-ref"
          },
          "secretRef": {
            "name": "some-secret-ref"
          }
        }
      ],
      "source": {
        "blob": {
          "url": "some-blob-url"
        }
      }
    }
  },
  {
    "reason": "BUILDPACK",
    "old": [
      {
        "id": "another-buildpack-id",
        "version": "another-buildpack-old-version"
      },
      {
        "id": "some-buildpack-id",
        "version": "some-buildpack-old-version"
      }
    ],
    "new": [
      {
        "id": "some-buildpack-id",
        "version": "some-buildpack-new-version"
      }
    ]
  },
  {
    "reason": "STACK",
    "old": "sha256:87302783be0a0cab9fde5b68c9954b7e9150ca0d514ba542e9810c3c6f2984ad",
    "new": "sha256:87302783be0a0cab9fde5b68c9954b7e9150ca0d514ba542e9810c3c6f2984ae"
  }
]`)
					})

					it("displays the correct reason and changes diff", func() {
						change := diffBuilder.SetPrefix("").
							New("A new build was manually triggered on Fri, 20 Nov 2020 15:38:15 -0500").SetPrefix(diffPadding).
							Old("old-revision").
							New("new-revision").
							New("bindings:").
							New("- metadataRef:").
							New("    name: some-metadata-ref").
							New("  name: binding-name").
							New("  secretRef:").
							New("    name: some-secret-ref").
							NoD("env:").
							Old("- name: env-var-name").
							Old("  value: env-var-value").
							New("- name: new-env-var-name").
							New("  value: new-env-var-value").
							NoD("- name: another-env-var-name").
							NoD("  value: another-env-var-value").
							NoD("resources:").
							NoD("  limits:").
							Old("    cpu: 500m").
							Old("    memory: 2G").
							New("    cpu: 300m").
							New("    memory: 1G").
							NoD("  requests:").
							Old("    cpu: 100m").
							New("    cpu: 200m").
							NoD("    memory: 512M").
							NoD("source:").
							Old("  git:").
							Old("    revision: some-git-revision").
							Old("    url: some-git-url").
							Old("  subPath: some-sub-path").
							New("  blob:").
							New("    url: some-blob-url").
							Old("- id: another-buildpack-id").SetPrefix(diffPadding).
							Old("  version: another-buildpack-old-version").
							NoD("- id: some-buildpack-id").
							Old("  version: some-buildpack-old-version").
							New("  version: some-buildpack-new-version").
							Old("sha256:87302783be0a0cab9fde5b68c9954b7e9150ca0d514ba542e9810c3c6f2984ad").
							New("sha256:87302783be0a0cab9fde5b68c9954b7e9150ca0d514ba542e9810c3c6f2984ae").Out()
						expectedOutput := og.output(t, "TRIGGER,COMMIT,CONFIG,BUILDPACK,STACK", change)

						testhelpers.CommandTest{
							Objects:        []runtime.Object{bld},
							Args:           []string{image},
							ExpectedOutput: expectedOutput,
						}.TestKpack(t, cmdFunc)
					})
				})
			})
		})

		when("using the --bom flag", func() {
			builds := testhelpers.BuildsToRuntimeObjs(testhelpers.MakeTestBuilds(image, defaultNamespace))

			it("prints the registry image bom only", func() {
				testhelpers.CommandTest{
					Objects:        builds,
					Args:           []string{image, "-b", "1", "--bom"},
					ExpectedOutput: "{\"some\":\"metadata\"}\n",
				}.TestKpack(t, cmdFunc)
			})

			it("returns error when build is not successful", func() {
				testhelpers.CommandTest{
					Objects:             builds,
					Args:                []string{image, "--bom"},
					ExpectErr:           true,
					ExpectedErrorOutput: "Error: build has failed or has not finished\n",
				}.TestKpack(t, cmdFunc)
			})
		})
	})
}

type outputGenerator struct {
	template *template.Template
}

func newOutputGenerator(t *testing.T, templateStr string) outputGenerator {
	template, err := template.New("output").Parse(templateStr)
	assert.NoError(t, err)

	return outputGenerator{
		template: template,
	}
}

type templateInput struct {
	Reason string
	Change string
}

func (o outputGenerator) output(t *testing.T, reason, change string) string {
	var sb strings.Builder

	err := o.template.Execute(&sb, templateInput{
		Reason: reason,
		Change: change,
	})

	assert.NoError(t, err)
	return sb.String()
}
