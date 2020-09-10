// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/builder"
	"github.com/pivotal/build-service-cli/pkg/k8s"
)

const (
	defaultStack             = "default"
	defaultStore             = "default"
	kubectlLastAppliedConfig = "kubectl.kubernetes.io/last-applied-configuration"
)

func NewCreateCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	var (
		tag       string
		namespace string
		stack     string
		store     string
		order     string
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
			name := args[0]

			cs, err := clientSetProvider.GetClientSet(namespace)
			if err != nil {
				return err
			}

			return create(name, tag, cs.Namespace, stack, store, order, cmd, cs)
		},
	}
	cmd.Flags().StringVarP(&tag, "tag", "t", "", "registry location where the builder will be created")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace")
	cmd.Flags().StringVarP(&stack, "stack", "s", defaultStack, "stack resource to use")
	cmd.Flags().StringVar(&store, "store", defaultStore, "buildpack store to use")
	cmd.Flags().StringVarP(&order, "order", "o", "", "path to buildpack order yaml")
	_ = cmd.MarkFlagRequired("tag")

	return cmd
}

func create(name, tag, namespace, stack, store, order string, cmd *cobra.Command, cs k8s.ClientSet) (err error) {
	bldr := &v1alpha1.Builder{
		TypeMeta: metaV1.TypeMeta{
			Kind:       v1alpha1.BuilderKind,
			APIVersion: "kpack.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: map[string]string{},
		},
		Spec: v1alpha1.NamespacedBuilderSpec{
			BuilderSpec: v1alpha1.BuilderSpec{
				Tag: tag,
				Stack: corev1.ObjectReference{
					Name: stack,
					Kind: v1alpha1.ClusterStackKind,
				},
				Store: corev1.ObjectReference{
					Name: store,
					Kind: v1alpha1.ClusterStoreKind,
				},
			},
			ServiceAccount: "default",
		},
	}

	bldr.Spec.Order, err = builder.ReadOrder(order)
	if err != nil {
		return err
	}

	marshal, err := json.Marshal(bldr)
	if err != nil {
		return err
	}

	bldr.Annotations[kubectlLastAppliedConfig] = string(marshal)

	_, err = cs.KpackClient.KpackV1alpha1().Builders(cs.Namespace).Create(bldr)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(cmd.OutOrStdout(), "\"%s\" created\n", bldr.Name)
	return err
}
