// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package custombuilder

import (
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"

	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
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
		namespace string
		stack     string
		store     string
		order     string
	)

	cmd := &cobra.Command{
		Use:   "create <name> <tag>",
		Short: "Create a custom builder",
		Long: `Create a custom builder by providing command line arguments.
This custom builder will be created only if it does not exist in the provided namespace.

namespace defaults to the kubernetes current-context namespace.`,
		Example: `kp cb create my-builder my-registry.com/my-builder-tag --order /path/to/order.yaml --stack tiny --store my-store
kp cb create my-builder my-registry.com/my-builder-tag --order /path/to/order.yaml`,
		Args:         cobra.ExactArgs(2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			tag := args[1]

			cs, err := clientSetProvider.GetClientSet(namespace)
			if err != nil {
				return err
			}

			cb := &expv1alpha1.CustomBuilder{
				TypeMeta: metaV1.TypeMeta{
					Kind:       expv1alpha1.CustomBuilderKind,
					APIVersion: "experimental.kpack.pivotal.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        name,
					Namespace:   cs.Namespace,
					Annotations: map[string]string{},
				},
				Spec: expv1alpha1.CustomNamespacedBuilderSpec{
					CustomBuilderSpec: expv1alpha1.CustomBuilderSpec{
						Tag: tag,
						Stack: corev1.ObjectReference{
							Name: stack,
							Kind: expv1alpha1.ClusterStackKind,
						},
						Store: corev1.ObjectReference{
							Name: store,
							Kind: expv1alpha1.ClusterStoreKind,
						},
					},
					ServiceAccount: "default",
				},
			}

			cb.Spec.Order, err = builder.ReadOrder(order)
			if err != nil {
				return err
			}

			marshal, err := json.Marshal(cb)
			if err != nil {
				return err
			}

			cb.Annotations[kubectlLastAppliedConfig] = string(marshal)

			_, err = cs.KpackClient.ExperimentalV1alpha1().CustomBuilders(cs.Namespace).Create(cb)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "\"%s\" created\n", cb.Name)
			return err
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace")
	cmd.Flags().StringVarP(&stack, "stack", "s", defaultStack, "stack resource to use")
	cmd.Flags().StringVar(&store, "store", defaultStore, "buildpack store to use")
	cmd.Flags().StringVarP(&order, "order", "o", "", "path to buildpack order yaml")

	return cmd
}
