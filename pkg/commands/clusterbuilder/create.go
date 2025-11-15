// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterbuilder

import (
	"context"
	"fmt"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"

	"github.com/buildpacks-community/kpack-cli/pkg/builder"
	"github.com/buildpacks-community/kpack-cli/pkg/commands"
	"github.com/buildpacks-community/kpack-cli/pkg/config"
	"github.com/buildpacks-community/kpack-cli/pkg/dockercreds"
	"github.com/buildpacks-community/kpack-cli/pkg/k8s"
	"github.com/buildpacks-community/kpack-cli/pkg/registry"
)

const (
	apiVersion   = "kpack.io/v1alpha2"
	defaultStack = "default"
)

func NewCreateCommand(clientSetProvider k8s.ClientSetProvider, newWaiter func(dynamic.Interface) commands.ResourceWaiter) *cobra.Command {
	var (
		flags     CommandFlags
		tlsConfig registry.TLSConfig
	)

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a cluster builder",
		Long: `Create a cluster builder by providing command line arguments.
The cluster builder will be created only if it does not exist.

A buildpack order must be provided with either the path to an order yaml, via the --buildpack flag, or extracted from a builder image using --order-from.
Multiple buildpacks provided via the --buildpack flag will be added to the same order group.

Tag when not specified, defaults to a combination of the default repository and specified builder name.
The default repository is read from the "default.repository" key in the "kp-config" ConfigMap within "kpack" namespace.
`,
		Example: `kp clusterbuilder create my-builder --order /path/to/order.yaml --stack tiny --store my-store
kp clusterbuilder create my-builder --buildpack my-buildpack-id --buildpack my-other-buildpack@1.0.1
kp clusterbuilder create my-builder --order-from paketobuildpacks/builder-jammy-base --stack tiny --store my-store
kp clusterbuilder create my-builder --tag my-registry.com/my-builder-tag --order /path/to/order.yaml --stack tiny --store my-store
kp clusterbuilder create my-builder --tag my-registry.com/my-builder-tag --buildpack my-buildpack-id --buildpack my-other-buildpack@1.0.1`,
		Args:         commands.ExactArgsWithUsage(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			ch, err := commands.NewCommandHelper(cmd)
			if err != nil {
				return err
			}

			name := args[0]
			ctx := cmd.Context()

			fetcher := registry.NewDefaultFetcher(tlsConfig)
			return create(ctx, name, flags, ch, cs, fetcher, newWaiter(cs.DynamicClient))
		},
	}

	cmd.Flags().StringVarP(&flags.tag, "tag", "t", "", "registry location where the builder will be created")
	cmd.Flags().StringVarP(&flags.stack, "stack", "s", defaultStack, "stack resource to use")
	cmd.Flags().StringVar(&flags.store, "store", "", "buildpack store to use")
	cmd.Flags().StringVarP(&flags.order, "order", "o", "", "path to buildpack order yaml")
	cmd.Flags().StringSliceVarP(&flags.buildpacks, "buildpack", "b", []string{}, "buildpack id and optional version in the form of either '<buildpack>@<version>' or '<buildpack>'\n  repeat for each buildpack in order, or supply once with comma-separated list")
	cmd.Flags().StringVar(&flags.orderFrom, "order-from", "", "builder image to extract buildpack order from")
	commands.SetDryRunOutputFlags(cmd)
	commands.SetTLSFlags(cmd, &tlsConfig)
	return cmd
}

type CommandFlags struct {
	tag        string
	stack      string
	store      string
	order      string
	buildpacks []string
	orderFrom  string
}

func create(ctx context.Context, name string, flags CommandFlags, ch *commands.CommandHelper, cs k8s.ClientSet, fetcher builder.Fetcher, waiter commands.ResourceWaiter) error {
	kpConfig := config.NewKpConfigProvider(cs.K8sClient).GetKpConfig(ctx)

	if flags.tag == "" {
		repo, err := kpConfig.DefaultRepository()
		if err != nil {
			return err
		}

		flags.tag = fmt.Sprintf("%s:clusterbuilder-%s", repo, name)
	}

	cb := &v1alpha2.ClusterBuilder{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha2.ClusterBuilderKind,
			APIVersion: apiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: map[string]string{},
		},
		Spec: v1alpha2.ClusterBuilderSpec{
			BuilderSpec: v1alpha2.BuilderSpec{
				Tag: flags.tag,
				Stack: corev1.ObjectReference{
					Name: flags.stack,
					Kind: v1alpha2.ClusterStackKind,
				},
			},
			ServiceAccountRef: kpConfig.ServiceAccount(),
		},
	}

	// Validate that only one order source is provided
	orderSourceCount := 0
	if len(flags.buildpacks) > 0 {
		orderSourceCount++
	}
	if flags.order != "" {
		orderSourceCount++
	}
	if flags.orderFrom != "" {
		orderSourceCount++
	}
	if orderSourceCount > 1 {
		return fmt.Errorf("only one of --order, --buildpack, or --order-from can be specified")
	}

	// Set the order based on the provided flag
	var err error
	if len(flags.buildpacks) > 0 {
		cb.Spec.Order = builder.CreateOrder(flags.buildpacks)
	} else if flags.order != "" {
		cb.Spec.Order, err = builder.ReadOrder(flags.order)
		if err != nil {
			return err
		}
	} else if flags.orderFrom != "" {
		keychain := dockercreds.DefaultKeychain
		cb.Spec.Order, err = builder.ReadOrderFromImage(keychain, fetcher, flags.orderFrom)
		if err != nil {
			return err
		}
	}

	if flags.store != "" {
		cb.Spec.Store = corev1.ObjectReference{
			Name: flags.store,
			Kind: v1alpha2.ClusterStoreKind,
		}
	}

	err = k8s.SetLastAppliedCfg(cb)
	if err != nil {
		return err
	}

	if !ch.IsDryRun() {
		cb, err = cs.KpackClient.KpackV1alpha2().ClusterBuilders().Create(ctx, cb, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		if err := waiter.Wait(ctx, cb); err != nil {
			return err
		}
	}

	cbArray := []runtime.Object{cb}

	err = ch.PrintObjs(cbArray)
	if err != nil {
		return err
	}

	return ch.PrintResult("ClusterBuilder %q created", cb.Name)
}
