// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"fmt"
	"io"
	"io/ioutil"
	"reflect"

	"github.com/pivotal/kpack/pkg/apis/build"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/vmware-tanzu/kpack-cli/pkg/k8s"
	"github.com/vmware-tanzu/kpack-cli/pkg/kpackcompat"
)

type CommandHelper struct {
	dryRun          bool
	dryRunImgUpload bool
	output          bool
	wait            bool
	timestamps      bool

	outWriter  io.Writer
	errWriter  io.Writer
	objPrinter k8s.ObjectPrinter

	typeToGVK map[reflect.Type]schema.GroupVersionKind
}

const (
	DryRunFlag          = "dry-run"
	DryRunImgUploadFlag = "dry-run-with-image-upload"
	OutputFlag          = "output"
	WaitFlag            = "wait"
	TimestampsFlag      = "timestamps"
)

func NewCommandHelper(cmd *cobra.Command) (*CommandHelper, error) {
	dryRun, err := GetBoolFlag(DryRunFlag, cmd)
	if err != nil {
		return nil, err
	}

	dryRunImgUpload, err := GetBoolFlag(DryRunImgUploadFlag, cmd)
	if err != nil {
		return nil, err
	}

	output, err := GetStringFlag(OutputFlag, cmd)
	if err != nil {
		return nil, err
	}

	wait, err := GetBoolFlag(WaitFlag, cmd)
	if err != nil {
		return nil, err
	}

	timestamps, err := GetBoolFlag(TimestampsFlag, cmd)
	if err != nil {
		return nil, err
	}

	var objPrinter k8s.ObjectPrinter

	outputResource := len(output) > 0
	if outputResource {
		objPrinter, err = k8s.NewObjectPrinter(output)
		if err != nil {
			return nil, err
		}
	}

	return &CommandHelper{
		dryRun:          dryRun,
		dryRunImgUpload: dryRunImgUpload,
		output:          outputResource,
		wait:            wait,
		timestamps:      timestamps,
		outWriter:       cmd.OutOrStdout(),
		errWriter:       cmd.ErrOrStderr(),
		objPrinter:      objPrinter,
		typeToGVK:       getTypeToGVKLookup(),
	}, nil
}

func (ch CommandHelper) IsDryRun() bool {
	return ch.dryRun || ch.dryRunImgUpload
}

func (ch CommandHelper) IsUploading() bool {
	return !ch.dryRun || ch.dryRunImgUpload
}

func (ch CommandHelper) ValidateOnly() bool {
	return ch.dryRun && !ch.output
}

func (ch CommandHelper) CanChangeState() bool {
	return !ch.dryRun
}

func (ch CommandHelper) ShouldWait() bool {
	return ch.wait && !ch.IsDryRun() && !ch.output
}

func (ch CommandHelper) ShowTimestamp() bool {
	return ch.timestamps
}

func (ch CommandHelper) PrintObjs(objs []runtime.Object) error {
	if !ch.output {
		return nil
	}
	for _, obj := range objs {
		err := ch.checkKind(obj)

		if err != nil {
			return err
		}
	}
	err := ch.objPrinter.PrintObject(objs, ch.outWriter)

	if err != nil {
		return err
	}
	return nil
}

func (ch CommandHelper) checkKind(obj runtime.Object) error {
	oGVK := obj.GetObjectKind().GroupVersionKind()
	if oGVK.Version == "" || oGVK.Kind == "" {
		nGVK, ok := ch.typeToGVK[reflect.TypeOf(obj)]
		if !ok {
			return errors.Errorf("failed to output. unknown type %q", reflect.TypeOf(obj))
		}
		obj.GetObjectKind().SetGroupVersionKind(nGVK)
	}
	return nil
}

func (ch CommandHelper) PrintChangeResult(change bool, format string, args ...interface{}) error {
	if ch.dryRunImgUpload {
		format += " (dry run with image upload)"
	} else if ch.dryRun {
		format += " (dry run)"
	} else if !change {
		format += " (no change)"
	}
	_, err := ch.OutOrDiscardWriter().Write([]byte(fmt.Sprintf(format+"\n", args...)))
	return err
}

func (ch CommandHelper) PrintResult(format string, args ...interface{}) error {
	if ch.dryRunImgUpload {
		format += " (dry run with image upload)"
	} else if ch.dryRun {
		format += " (dry run)"
	}
	_, err := ch.OutOrDiscardWriter().Write([]byte(fmt.Sprintf(format+"\n", args...)))
	return err
}

func (ch CommandHelper) PrintStatus(format string, args ...interface{}) error {
	if ch.dryRunImgUpload {
		format += " (dry run with image upload)"
	} else if ch.dryRun {
		format += " (dry run)"
	}
	_, err := ch.OutOrErrWriter().Write([]byte(fmt.Sprintf(format+"\n", args...)))
	return err
}

func (ch CommandHelper) Printlnf(format string, args ...interface{}) error {
	_, err := fmt.Fprintf(ch.OutOrErrWriter(), format+"\n", args...)
	return err
}

func (ch CommandHelper) OutOrErrWriter() io.Writer {
	if ch.output {
		return ch.errWriter
	} else {
		return ch.outWriter
	}
}

func (ch CommandHelper) OutOrDiscardWriter() io.Writer {
	if ch.output {
		return ioutil.Discard
	} else {
		return ch.outWriter
	}
}

func (ch CommandHelper) Writer() io.Writer {
	return ch.OutOrErrWriter()
}

func GetBoolFlag(name string, cmd *cobra.Command) (bool, error) {
	flag := cmd.Flags().Lookup(name)
	if flag == nil {
		return false, nil
	}

	if !cmd.Flags().Changed(name) {
		return false, nil
	}

	value, err := cmd.Flags().GetBool(name)
	if err != nil {
		return value, err
	}
	return value, nil
}

func GetStringFlag(name string, cmd *cobra.Command) (string, error) {
	flag := cmd.Flags().Lookup(name)
	if flag == nil {
		return "", nil
	}

	if !cmd.Flags().Changed(name) {
		return "", nil
	}

	value, err := cmd.Flags().GetString(name)
	if err != nil {
		return value, err
	}
	return value, nil
}

func getTypeToGVKLookup() map[reflect.Type]schema.GroupVersionKind {
	v1GV := schema.GroupVersion{Group: v1.GroupName, Version: "v1"}
	buildGV := schema.GroupVersion{Group: build.GroupName, Version: kpackcompat.LatestKpackAPIVersion}

	return map[reflect.Type]schema.GroupVersionKind{
		reflect.TypeOf(&v1.Secret{}):                 v1GV.WithKind("Secret"),
		reflect.TypeOf(&v1.ServiceAccount{}):         v1GV.WithKind("ServiceAccount"),
		reflect.TypeOf(&v1.ConfigMap{}):              v1GV.WithKind("ConfigMap"),
		reflect.TypeOf(&v1alpha2.Image{}):            buildGV.WithKind("Image"),
		reflect.TypeOf(&v1alpha2.Builder{}):          buildGV.WithKind(v1alpha2.BuilderKind),
		reflect.TypeOf(&v1alpha2.ClusterStack{}):     buildGV.WithKind(v1alpha2.ClusterStackKind),
		reflect.TypeOf(&v1alpha2.ClusterStore{}):     buildGV.WithKind(v1alpha2.ClusterStoreKind),
		reflect.TypeOf(&v1alpha2.ClusterBuilder{}):   buildGV.WithKind(v1alpha2.ClusterBuilderKind),
		reflect.TypeOf(&v1alpha2.ClusterLifecycle{}): buildGV.WithKind(v1alpha2.ClusterLifecycleKind),
	}
}
