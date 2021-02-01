// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"fmt"
	"io"
	"io/ioutil"
	"reflect"

	"github.com/pivotal/kpack/pkg/apis/build"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/pivotal/build-service-cli/pkg/k8s"
)

type CommandHelper struct {
	dryRun          bool
	dryRunImgUpload bool
	output          bool
	wait            bool

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
		outWriter:       cmd.OutOrStdout(),
		errWriter:       cmd.ErrOrStderr(),
		objPrinter:      objPrinter,
		typeToGVK:       getTypeToGVKLookup(),
	}, nil
}

func (ch CommandHelper) IsDryRun() bool {
	return ch.dryRun || ch.dryRunImgUpload
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

func (ch CommandHelper) PrintObjs(objs []runtime.Object) error {
	for _, obj := range objs {
		if err := ch.PrintObj(obj); err != nil {
			return err
		}
	}
	return nil
}

func (ch CommandHelper) PrintObj(obj runtime.Object) error {
	if !ch.output {
		return nil
	}

	oGVK := obj.GetObjectKind().GroupVersionKind()
	if oGVK.Version == "" || oGVK.Kind == "" {
		nGVK, ok := ch.typeToGVK[reflect.TypeOf(obj)]
		if !ok {
			return errors.Errorf("failed to output. unknown type %q", reflect.TypeOf(obj))
		}
		obj.GetObjectKind().SetGroupVersionKind(nGVK)
	}
	err := ch.objPrinter.PrintObject(obj, ch.outWriter)
	obj.GetObjectKind().SetGroupVersionKind(oGVK)
	return err
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
	buildGV := schema.GroupVersion{Group: build.GroupName, Version: "v1alpha1"}

	return map[reflect.Type]schema.GroupVersionKind{
		reflect.TypeOf(&v1.Secret{}):               v1GV.WithKind("Secret"),
		reflect.TypeOf(&v1.ServiceAccount{}):       v1GV.WithKind("ServiceAccount"),
		reflect.TypeOf(&v1.ConfigMap{}):            v1GV.WithKind("ConfigMap"),
		reflect.TypeOf(&v1alpha1.Image{}):          buildGV.WithKind("Image"),
		reflect.TypeOf(&v1alpha1.Builder{}):        buildGV.WithKind(v1alpha1.BuilderKind),
		reflect.TypeOf(&v1alpha1.ClusterStack{}):   buildGV.WithKind(v1alpha1.ClusterStackKind),
		reflect.TypeOf(&v1alpha1.ClusterStore{}):   buildGV.WithKind(v1alpha1.ClusterStoreKind),
		reflect.TypeOf(&v1alpha1.ClusterBuilder{}): buildGV.WithKind(v1alpha1.ClusterBuilderKind),
	}
}
