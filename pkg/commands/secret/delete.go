// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package secret

import (
	"fmt"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
)

func NewDeleteCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	var (
		namespace      string
		serviceAccount string
	)

	command := cobra.Command{
		Use:   "delete <name>",
		Short: "Delete secret",
		Long: `Deletes a specific secret in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.`,
		Example:      "kp secret delete my-secret",
		Args:         commands.ExactArgsWithUsage(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet(namespace)
			if err != nil {
				return err
			}

			ctx := cmd.Context()

			serviceAccount, err := cs.K8sClient.CoreV1().ServiceAccounts(cs.Namespace).Get(ctx, serviceAccount, metav1.GetOptions{})
			if err != nil {
				return err
			}

			updatedSA, err := deleteSecretsFromServiceAccount(serviceAccount, args[0])
			if err != nil {
				return err
			}

			patch, err := k8s.CreatePatch(serviceAccount, updatedSA)
			if err != nil {
				return err
			}

			if len(patch) > 0 {
				_, err = cs.K8sClient.CoreV1().ServiceAccounts(cs.Namespace).Patch(ctx, updatedSA.Name, types.MergePatchType, patch, metav1.PatchOptions{})
				if err != nil {
					return err
				}
			}

			err = cs.K8sClient.CoreV1().Secrets(cs.Namespace).Delete(ctx, args[0], metav1.DeleteOptions{})
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Secret %q deleted\n", args[0])
			return err
		},
	}

	command.Flags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace")
	command.Flags().StringVar(&serviceAccount, "service-account", "default", "service account name to use")

	return &command
}

func deleteSecretsFromServiceAccount(sa *corev1.ServiceAccount, name string) (*corev1.ServiceAccount, error) {
	updatedSA := sa.DeepCopy()
	managedSecrets, err := readManagedSecrets(updatedSA)
	if err != nil {
		return nil, err
	}

	for i, s := range updatedSA.Secrets {
		if s.Name == name {
			updatedSA.Secrets = append(updatedSA.Secrets[:i], updatedSA.Secrets[i+1:]...)
			delete(managedSecrets, s.Name)
			break
		}
	}
	for i, s := range updatedSA.ImagePullSecrets {
		if s.Name == name {
			updatedSA.ImagePullSecrets = append(updatedSA.ImagePullSecrets[:i], updatedSA.ImagePullSecrets[i+1:]...)
			delete(managedSecrets, s.Name)
			break
		}
	}

	err = writeManagedSecrets(managedSecrets, updatedSA)
	if err != nil {
		return nil, err
	}

	return updatedSA, nil
}
