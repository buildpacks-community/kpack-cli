// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"context"
	"fmt"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"

	"github.com/pivotal/build-service-cli/pkg/builder"
	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/k8s"
)

const (
	defaultStack = "default"
	defaultStore = "default"
)

func NewCreateCommand(clientSetProvider k8s.ClientSetProvider, newWaiter func(dynamic.Interface) commands.ResourceWaiter) *cobra.Command {
	var (
		flags CommandFlags
	)

	cmd := &cobra.Command{
		Use:   "create <name> --tag <tag>",
		Short: "Create a builder",
		Long: `Create a builder by providing command line arguments.
The builder will be created only if it does not exist in the provided namespace.

A buildpack order must be provided with either the path to an order yaml or via the --buildpack flag.
Multiple buildpacks provided via the --buildpack flag will be added to the same order group. 

The namespace defaults to the kubernetes current-context namespace.`,
		Example: `kp builder create my-builder --tag my-registry.com/my-builder-tag --order /path/to/order.yaml --stack tiny --store my-store
kp builder create my-builder --tag my-registry.com/my-builder-tag --order /path/to/order.yaml
kp builder create my-builder --tag my-registry.com/my-builder-tag --buildpack my-buildpack-id --buildpack my-other-buildpack@1.0.1`,
		Args:         commands.ExactArgsWithUsage(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet(flags.namespace)
			if err != nil {
				return err
			}

			ch, err := commands.NewCommandHelper(cmd)
			if err != nil {
				return err
			}

			name := args[0]
			flags.namespace = cs.Namespace

			ctx := cmd.Context()
			return create(ctx, name, flags, ch, cs, newWaiter(cs.DynamicClient))
		},
	}

	cmd.Flags().StringVarP(&flags.tag, "tag", "t", "", "registry location where the builder will be created")
	cmd.Flags().StringVarP(&flags.namespace, "namespace", "n", "", "kubernetes namespace")
	cmd.Flags().StringVarP(&flags.stack, "stack", "s", defaultStack, "stack resource to use")
	cmd.Flags().StringVar(&flags.store, "store", defaultStore, "buildpack store to use")
	cmd.Flags().StringVarP(&flags.order, "order", "o", "", "path to buildpack order yaml")
	cmd.Flags().StringSliceVarP(&flags.buildpacks, "buildpack", "b", []string{}, "buildpack id and optional version in the form of either '<buildpack>@<version>' or '<buildpack>'\n  repeat for each buildpack in order, or supply once with comma-separated list")
	commands.SetDryRunOutputFlags(cmd)
	_ = cmd.MarkFlagRequired("tag")
	return cmd
}

type CommandFlags struct {
	tag        string
	namespace  string
	stack      string
	store      string
	order      string
	buildpacks []string
}

func create(ctx context.Context, name string, flags CommandFlags, ch *commands.CommandHelper, cs k8s.ClientSet, w commands.ResourceWaiter) (err error) {
	bldr := &v1alpha1.Builder{
		TypeMeta: metav1.TypeMeta{
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

	if len(flags.buildpacks) > 0 && flags.order != "" {
		return fmt.Errorf("cannot use --order and --buildpack together")
	}

	if len(flags.buildpacks) > 0 {
		bldr.Spec.Order = builder.CreateOrder(flags.buildpacks)
	}

	if flags.order != "" {
		bldr.Spec.Order, err = builder.ReadOrder(flags.order)
		if err != nil {
			return err
		}
	}

	err = k8s.SetLastAppliedCfg(bldr)
	if err != nil {
		return err
	}

	if !ch.IsDryRun() {
		bldr, err = cs.KpackClient.KpackV1alpha1().Builders(cs.Namespace).Create(ctx, bldr, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		if err := w.Wait(ctx, bldr); err != nil {
			return err
		}
	}

	err = ch.PrintObj(bldr)
	if err != nil {
		return err
	}

	return ch.PrintResult("Builder %q created", bldr.Name)
}
