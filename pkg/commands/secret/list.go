// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package secret

import (
	"sort"
	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/buildpacks-community/kpack-cli/pkg/commands"
	"github.com/buildpacks-community/kpack-cli/pkg/k8s"
)

func NewListCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	var (
		namespace      string
		serviceAccount string
	)

	command := cobra.Command{
		Use:   "list",
		Short: "List secrets attached to a service account",
		Long: `List secrets for a service account in the provided namespace.

A secret attached to a service account that does not exist in the specified namespace will be listed as AVAILABLE "false".

The namespace defaults to the kubernetes current-context namespace.

The service account defaults to "default".`,
		Example:      "kp secret list\nkp secret list -n my-namespace",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet(namespace)
			if err != nil {
				return err
			}

			serviceAccount, err := cs.K8sClient.CoreV1().ServiceAccounts(cs.Namespace).Get(cmd.Context(), serviceAccount, metav1.GetOptions{})
			if err != nil {
				return err
			}
			secretsList, err := cs.K8sClient.CoreV1().Secrets(cs.Namespace).List(cmd.Context(), metav1.ListOptions{})
			if err != nil {
				return err
			}

			if len(serviceAccount.Secrets) == 0 && len(serviceAccount.ImagePullSecrets) == 0 {
				return errors.Errorf("no secrets found in %q namespace for %q service account", cs.Namespace, serviceAccount.Name)
			} else {
				return displaySecretsTable(cmd, serviceAccount, secretsList)
			}
		},
	}

	command.Flags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace")
	command.Flags().StringVar(&serviceAccount, "service-account", "default", "service account to list secrets for")

	return &command
}

func displaySecretsTable(cmd *cobra.Command, sa *corev1.ServiceAccount, secretsList *corev1.SecretList) error {
	secretNames, err := getServiceAccountSecretsInfo(sa, secretsList)
	if err != nil {
		return errors.WithMessage(err, "could not retrieve secrets information from service account.")
	}
	writer, err := commands.NewTableWriter(cmd.OutOrStdout(), "NAME", "TARGET", "AVAILABLE")
	if err != nil {
		return err
	}

	for _, secret := range secretNames {
		err = writer.AddRow(secret.name, secret.target, strconv.FormatBool(secret.isAvailable))
		if err != nil {
			return err
		}
	}

	return writer.Write()
}

func getServiceAccountSecretsInfo(sa *corev1.ServiceAccount, secretsList *corev1.SecretList) ([]struct {
	name        string
	target      string
	isAvailable bool
}, error) {
	managedSecrets, err := readManagedSecrets(sa)
	if err != nil {
		return nil, err
	}
	secretNameSet := map[string]interface{}{}
	for _, item := range append(sa.Secrets) {
		secretNameSet[item.Name] = nil
	}
	for _, item := range append(sa.ImagePullSecrets) {
		secretNameSet[item.Name] = nil
	}

	var secretNames []struct {
		name        string
		target      string
		isAvailable bool
	}
	for name := range secretNameSet {
		found := false
		for _, secret := range secretsList.Items {
			if secret.Name == name {
				found = true
				break
			}
		}
		secretNames = append(secretNames, struct {
			name        string
			target      string
			isAvailable bool
		}{name, managedSecrets[name], found})
	}
	sort.Slice(secretNames, func(i, j int) bool {
		return secretNames[i].name < secretNames[j].name
	})
	return secretNames, nil
}
