package coreself_test

import (
	"reflect"
	"sort"
	"testing"

	"github.com/cogpy/echo9llama/core/coreself"
)

func TestKernelExportedMethodAllowlist(t *testing.T) {
	typeOfKernel := reflect.TypeOf((*coreself.Kernel)(nil))
	methods := make([]string, 0, typeOfKernel.NumMethod())
	for index := range typeOfKernel.NumMethod() {
		methods = append(methods, typeOfKernel.Method(index).Name)
	}
	sort.Strings(methods)
	want := []string{"Close", "ExportCapsule", "Preflight", "StateSnapshot", "Status", "VerifyCapsule"}
	if !reflect.DeepEqual(methods, want) {
		t.Fatalf("exported Kernel methods = %v, want %v", methods, want)
	}
}
