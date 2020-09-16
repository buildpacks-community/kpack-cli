// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"encoding/json"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/builder"
	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/k8s"
)

const (
	defaultStack             = "default"
	defaultStore             = "default"
	kubectlLastAppliedConfig = "kubectl.kubernetes.io/last-applied-configuration"
)

func NewCreateCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	var (
		flags CommandFlags
	)

	cmd := &cobra.Command{
		Use:   "create <name> --tag <tag>",
		Short: "Create a builder",
		Long: `Create a builder by providing command line arguments.
The builder will be created only if it does not exist in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.`,
		Example: `kp builder create my-builder --tag my-registry.com/my-builder-tag --order /path/to/order.yaml --stack tiny --store my-store
kp builder create my-builder --tag my-registry.com/my-builder-tag --order /path/to/order.yaml`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet(flags.namespace)
			if err != nil {
				return err
			}

			name := args[0]
			flags.namespace = cs.Namespace

			return create(name, flags, cmd.OutOrStdout(), cs)
		},
	}

	cmd.Flags().StringVarP(&flags.tag, "tag", "t", "", "registry location where the builder will be created")
	cmd.Flags().StringVarP(&flags.namespace, "namespace", "n", "", "kubernetes namespace")
	cmd.Flags().StringVarP(&flags.stack, "stack", "s", defaultStack, "stack resource to use")
	cmd.Flags().StringVar(&flags.store, "store", defaultStore, "buildpack store to use")
	cmd.Flags().StringVarP(&flags.order, "order", "o", "", "path to buildpack order yaml")
	cmd.Flags().BoolVarP(&flags.dryRun, "dry-run", "", false, "only print the object that would be sent, without sending it")
	cmd.Flags().StringVarP(&flags.outputFormat, "output", "", "yaml", "output format. supported formats are: yaml, json")
	_ = cmd.MarkFlagRequired("tag")
	return cmd
}

type CommandFlags struct {
	tag   string
	namespace string
	stack string
	store string
	order string
	dryRun bool
	outputFormat string
}

func create(name string, flags CommandFlags, writer io.Writer, cs k8s.ClientSet) (err error) {
	bldr := &v1alpha1.Builder{
		TypeMeta: metaV1.TypeMeta{
			Kind:       v1alpha1.BuilderKind,
			APIVersion: "kpack.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   flags.namespace,
			Annotations: map[string]string{},
		},
		Spec: v1alpha1.NamespacedBuilderSpec{
			BuilderSpec: v1alpha1.BuilderSpec{
				Tag: flags.tag,
				Stack: corev1.ObjectReference{
					Name: flags.stack,
					Kind: v1alpha1.ClusterStackKind,
				},
				Store: corev1.ObjectReference{
					Name: flags.store,
					Kind: v1alpha1.ClusterStoreKind,
				},
			},
			ServiceAccount: "default",
		},
	}

	bldr.Spec.Order, err = builder.ReadOrder(flags.order)
	if err != nil {
		return err
	}

	marshal, err := json.Marshal(bldr)
	if err != nil {
		return err
	}

	bldr.Annotations[kubectlLastAppliedConfig] = string(marshal)

	if flags.dryRun {
		printer, err := commands.NewResourcePrinter(flags.outputFormat)
		if err != nil {
			return err
		}

		return printer.PrintObject(bldr, writer)
	}

	_, err = cs.KpackClient.KpackV1alpha1().Builders(cs.Namespace).Create(bldr)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(writer, "\"%s\" created\n", bldr.Name)
	return err
}
