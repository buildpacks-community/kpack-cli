package lifecycle

import (
	"fmt"
	"os"
	"path"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/pivotal/kpack/pkg/registry/imagehelpers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/commands"
	"github.com/pivotal/build-service-cli/pkg/k8s"
	"github.com/pivotal/build-service-cli/pkg/registry"
)

const (
	kpNamespace = "kpack"

	lifecycleConfigMapName = "lifecycle-image"
	lifecycleImageName     = "lifecycle"
	lifecycleImageKey      = "image"
	lifecycleMetadataLabel = "io.buildpacks.lifecycle.metadata"
)

func NewUpdateCommand(clientSetProvider k8s.ClientSetProvider, rup registry.UtilProvider) *cobra.Command {
	var (
		image  string
		tlsCfg registry.TLSConfig
	)

	cmd := &cobra.Command{
		Use:     "update --image <image-tag>",
		Short:   "Update lifecycle image used by kpack",
		Long:    `Update lifecycle image used by kpack

The Lifecycle image will be uploaded to the canonical repository.
Therefore, you must have credentials to access the registry on your machine.

The canonical repository is read from the "canonical.repository" key of the "kp-config" ConfigMap within "kpack" namespace.
`,
		Example: "kp lifecycle update --image my-registry.com/lifecycle",
		Args: commands.ExactArgsWithUsage(0),
		SilenceUsage: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if image == "" {
				return fmt.Errorf("required flag(s) \"image\" not set\n\n%s", cmd.UsageString())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			if err = updateLifecycleImage(image, tlsCfg, cs, rup); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Updated lifecycle image")
			return nil
		},
	}
	cmd.Flags().StringVarP(&image, "image", "i", "", "location of the image")
	commands.SetTLSFlags(cmd, &tlsCfg)
	return cmd
}

func updateLifecycleImage(srcImgTag string, tlsCfg registry.TLSConfig, cs k8s.ClientSet, rup registry.UtilProvider) error {
	img, err := rup.Fetcher().Fetch(srcImgTag, tlsCfg)
	if err != nil {
		return err
	}

	if err = validateImage(img); err != nil {
		return err
	}

	dstRepo, err := k8s.DefaultConfigHelper(cs).GetCanonicalRepository()
	if err != nil {
		return err
	}

	dstImgTag, err := rup.Relocator(true).Relocate(img, path.Join(dstRepo, lifecycleImageName), os.Stdout, tlsCfg)
	if err != nil {
		return err
	}

	return updateConfigMapWithImage(dstImgTag, cs)
}

func validateImage(img v1.Image) error {
	hasLabel, err := imagehelpers.HasLabel(img, lifecycleMetadataLabel)
	if err != nil {
		return err
	}

	if !hasLabel {
		return errors.New("image missing lifecycle metadata")
	}
	return nil
}

func updateConfigMapWithImage(dstImgTag string, cs k8s.ClientSet) error {
	cm, err := cs.K8sClient.CoreV1().ConfigMaps(kpNamespace).Get(lifecycleConfigMapName, metav1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		return errors.Errorf("configmap %q not found in %q namespace", lifecycleConfigMapName, kpNamespace)
	} else if err != nil {
		return err
	}

	cm.Data[lifecycleImageKey] = dstImgTag
	_, err = cs.K8sClient.CoreV1().ConfigMaps(kpNamespace).Update(cm)
	return err
}
