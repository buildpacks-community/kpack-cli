// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore

import (
	"errors"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/k8s"
)

func NewListCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List cluster stores",
		Long:    "Prints a table of the most important information about cluster-scoped stores",
		Example: "kp clusterstore list",
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			storeList, err := cs.KpackClient.KpackV1alpha1().ClusterStores().List(metav1.ListOptions{})
			if err != nil {
				return err
			}

			if len(storeList.Items) == 0 {
				return errors.New("no clusterstores found")
			} else {
				return displayStoresTable(cmd, storeList)
			}

		},
		SilenceUsage: true,
	}

	return cmd
}

func displayStoresTable(cmd *cobra.Command, storeList *v1alpha1.ClusterStoreList) error {
	writer, err := commands.NewTableWriter(cmd.OutOrStdout(), "NAME", "READY")
	if err != nil {
		return err
	}

	for _, s := range storeList.Items {
		err := writer.AddRow(s.Name, getReadyText(s))
		if err != nil {
			return err
		}
	}

	return writer.Write()
}

func getReadyText(s v1alpha1.ClusterStore) string {
	cond := s.Status.GetCondition(corev1alpha1.ConditionReady)
	if cond == nil {
		return "Unknown"
	}
	return string(cond.Status)
}
