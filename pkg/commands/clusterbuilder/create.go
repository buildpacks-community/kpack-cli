// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterbuilder

import (
	"encoding/json"
	"fmt"
	"path"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"

	"github.com/pivotal/build-service-cli/pkg/builder"
	"github.com/pivotal/build-service-cli/pkg/k8s"
)

const (
	kpNamespace              = "kpack"
	apiVersion               = "kpack.io/v1alpha1"
	kubectlLastAppliedConfig = "kubectl.kubernetes.io/last-applied-configuration"
)

func NewCreateCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	var (
		tag   string
		stack string
		store string
		order string
	)

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a cluster builder",
		Long: `Create a cluster builder by providing command line arguments.
The cluster builder will be created only if it does not exist.

Tag when not specified, defaults to a combination of the canonical repository and specified builder name.
The canonical repository is read from the "canonical.repository" key in the "kp-config" ConfigMap within "kpack" namespace.
`,
		Example: `kp cb create my-builder --order /path/to/order.yaml --stack tiny --store my-store
kp cb create my-builder --order /path/to/order.yaml
kp cb create my-builder --tag my-registry.com/my-builder-tag --order /path/to/order.yaml --stack tiny --store my-store
kp cb create my-builder --tag my-registry.com/my-builder-tag --order /path/to/order.yaml`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			configHelper := k8s.DefaultConfigHelper(cs)

			if tag == "" {
				repository, err := configHelper.GetCanonicalRepository()
				if err != nil {
					return err
				}

				tag = path.Join(repository, name)
			}

			serviceAccount, err := configHelper.GetCanonicalServiceAccount()
			if err != nil {
				return err
			}

			ccb := &v1alpha1.ClusterBuilder{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1alpha1.ClusterBuilderKind,
					APIVersion: apiVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        name,
					Annotations: map[string]string{},
				},
				Spec: v1alpha1.ClusterBuilderSpec{
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
					ServiceAccountRef: corev1.ObjectReference{
						Namespace: kpNamespace,
						Name:      serviceAccount,
					},
				},
			}

			ccb.Spec.Order, err = builder.ReadOrder(order)
			if err != nil {
				return err
			}

			marshal, err := json.Marshal(ccb)
			if err != nil {
				return err
			}

			ccb.Annotations[kubectlLastAppliedConfig] = string(marshal)

			_, err = cs.KpackClient.KpackV1alpha1().ClusterBuilders().Create(ccb)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "\"%s\" created\n", ccb.Name)
			return err
		},
	}
	cmd.Flags().StringVarP(&tag, "tag", "t", "", "registry location where the builder will be created")
	cmd.Flags().StringVarP(&stack, "stack", "s", "default", "stack resource to use")
	cmd.Flags().StringVar(&store, "store", "default", "buildpack store to use")
	cmd.Flags().StringVarP(&order, "order", "o", "", "path to buildpack order yaml")

	return cmd
}
