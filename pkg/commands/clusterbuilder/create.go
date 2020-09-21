// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterbuilder

import (
	"encoding/json"
	"path"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"

	"github.com/pivotal/build-service-cli/pkg/builder"
	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/k8s"
)

const (
	kpNamespace              = "kpack"
	apiVersion               = "kpack.io/v1alpha1"
	defaultStack             = "default"
	defaultStore             = "default"
	kubectlLastAppliedConfig = "kubectl.kubernetes.io/last-applied-configuration"
)

func NewCreateCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	var (
		flags CommandFlags
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
			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			ch, err := commands.NewCommandHelper(cmd)
			if err != nil {
				return err
			}

			name := args[0]

			return create(name, flags, ch, cs)
		},
	}

	cmd.Flags().StringVarP(&flags.tag, "tag", "t", "", "registry location where the builder will be created")
	cmd.Flags().StringVarP(&flags.stack, "stack", "s", defaultStack, "stack resource to use")
	cmd.Flags().StringVar(&flags.store, "store", defaultStore, "buildpack store to use")
	cmd.Flags().StringVarP(&flags.order, "order", "o", "", "path to buildpack order yaml")
	cmd.Flags().BoolVarP(&flags.dryRun, "dry-run", "", false, "only print the object that would be sent, without sending it")
	cmd.Flags().StringVar(&flags.output, "output", "", "output format. supported formats are: yaml, json")
	return cmd
}

type CommandFlags struct {
	tag    string
	stack  string
	store  string
	order  string
	dryRun bool
	output string
}

func create(name string, flags CommandFlags, ch *commands.CommandHelper, cs k8s.ClientSet) error {
	configHelper := k8s.DefaultConfigHelper(cs)

	if flags.tag == "" {
		repository, err := configHelper.GetCanonicalRepository()
		if err != nil {
			return err
		}

		flags.tag = path.Join(repository, name)
	}

	serviceAccount, err := configHelper.GetCanonicalServiceAccount()
	if err != nil {
		return err
	}

	cb := &v1alpha1.ClusterBuilder{
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
			ServiceAccountRef: corev1.ObjectReference{
				Namespace: kpNamespace,
				Name:      serviceAccount,
			},
		},
	}

	cb.Spec.Order, err = builder.ReadOrder(flags.order)
	if err != nil {
		return err
	}

	marshal, err := json.Marshal(cb)
	if err != nil {
		return err
	}

	cb.Annotations[kubectlLastAppliedConfig] = string(marshal)

	if !ch.IsDryRun() {
		cb, err = cs.KpackClient.KpackV1alpha1().ClusterBuilders().Create(cb)
		if err != nil {
			return err
		}
	}

	err = ch.PrintObj(cb)
	if err != nil {
		return err
	}

	return ch.PrintResult("%q created", cb.Name)
}
