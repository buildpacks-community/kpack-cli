// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package secret

import (
	"sort"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/k8s"
)

func NewListCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	var (
		namespace string
	)

	command := cobra.Command{
		Use:   "list",
		Short: "List secrets",
		Long: `Prints a table of the most important information about secrets in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.`,
		Example:      "kp secret list\nkp secret list -n my-namespace",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet(namespace)
			if err != nil {
				return err
			}

			serviceAccount, err := cs.K8sClient.CoreV1().ServiceAccounts(cs.Namespace).Get("default", metav1.GetOptions{})
			if err != nil {
				return err
			}

			if len(serviceAccount.Secrets) == 0 && len(serviceAccount.ImagePullSecrets) == 0 {
				return errors.Errorf("no secrets found in %q namespace", cs.Namespace)
			} else {
				return displaySecretsTable(cmd, serviceAccount)
			}
		},
	}

	command.Flags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace")

	return &command
}

func displaySecretsTable(cmd *cobra.Command, sa *corev1.ServiceAccount) error {
	managedSecrets, err := readManagedSecrets(sa)
	if err != nil {
		return err
	}

	secretNameSet := map[string]interface{}{}
	for _, item := range append(sa.Secrets) {
		secretNameSet[item.Name] = nil
	}
	for _, item := range append(sa.ImagePullSecrets) {
		secretNameSet[item.Name] = nil
	}

	var secretNames []string
	for name := range secretNameSet {
		secretNames = append(secretNames, name)
	}
	sort.Strings(secretNames)

	writer, err := commands.NewTableWriter(cmd.OutOrStdout(), "NAME", "TARGET")
	if err != nil {
		return err
	}

	for _, name := range secretNames {
		err := writer.AddRow(name, managedSecrets[name])
		if err != nil {
			return err
		}
	}

	return writer.Write()
}
