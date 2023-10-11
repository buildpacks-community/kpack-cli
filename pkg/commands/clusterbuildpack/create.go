package clusterbuildpack

import (
	"context"
	"k8s.io/apimachinery/pkg/runtime"

	buildv1alpha2 "github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/spf13/cobra"
	"github.com/vmware-tanzu/kpack-cli/pkg/commands"
	"github.com/vmware-tanzu/kpack-cli/pkg/config"
	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
)

const (
	defaultServiceAccount = "default"
)

type CommandFlags struct {
	image string
}

func NewCreateCommand(clientSetProvider k8s.ClientSetProvider, newWaiter func(dynamic.Interface) commands.ResourceWaiter) *cobra.Command {
	var (
		flags CommandFlags
	)

	cmd := &cobra.Command{
		Use:   "create <name> --image <image>",
		Short: "Create a cluster buildpack",
		Long: `Create a cluster buildpack by providing command line arguments.

The default service account used is read from the "default.repository.serviceaccount" key in the "kp-config" ConfigMap within "kpack" namespace.
`,
		Example: `kp clusterbuildpack create my-cluster-buildpack --image gcr.io/paketo-buildpacks/java
kp clusterbuildpack create my-cluster-buildpack --image gcr.io/paketo-buildpacks/java:8.9.0
kp clusterbuildpack create my-cluster-buildpack --image gcr.io/paketo-buildpacks/java@sha256:fc1c6fba46b582f63b13490b89e50e93c95ce08142a8737f4a6b70c826c995de
`,
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
			return create(ctx, name, flags, ch, cs, newWaiter(cs.DynamicClient))
		},
	}

	cmd.Flags().StringVarP(&flags.image, "image", "i", "", "registry location where the cluster buildpack is located")
	commands.SetDryRunOutputFlags(cmd)
	_ = cmd.MarkFlagRequired("image")
	return cmd
}

func create(ctx context.Context, name string, flags CommandFlags, ch *commands.CommandHelper, cs k8s.ClientSet, w commands.ResourceWaiter) (err error) {
	svcAcc := config.NewKpConfigProvider(cs.K8sClient).GetKpConfig(ctx).ServiceAccount()

	bp := &buildv1alpha2.ClusterBuildpack{
		TypeMeta: metav1.TypeMeta{
			Kind:       buildv1alpha2.ClusterBuildpackKind,
			APIVersion: buildv1alpha2.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: map[string]string{},
		},
		Spec: buildv1alpha2.ClusterBuildpackSpec{
			ImageSource: corev1alpha1.ImageSource{
				Image: flags.image,
			},
			ServiceAccountRef: &svcAcc,
		},
	}

	err = k8s.SetLastAppliedCfg(bp)
	if err != nil {
		return err
	}

	if !ch.IsDryRun() {
		bp, err = cs.KpackClient.KpackV1alpha2().ClusterBuildpacks().Create(ctx, bp, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		if err = w.Wait(ctx, bp); err != nil {
			return err
		}
	}

	bpArray := []runtime.Object{bp}

	err = ch.PrintObjs(bpArray)
	if err != nil {
		return err
	}

	return ch.PrintResult("Cluster Buildpack %q created", bp.Name)
}
