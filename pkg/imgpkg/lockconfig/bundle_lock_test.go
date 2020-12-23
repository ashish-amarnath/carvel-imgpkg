// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package lockconfig_test

import (
	"strings"
	"testing"

	"github.com/k14s/imgpkg/pkg/imgpkg/lockconfig"
)

func TestBundleLockNonDigestUnmarshalError(t *testing.T) {
	data := `
apiVersion: imgpkg.carvel.dev/v1alpha1
kind: BundleLock
bundle:
  image: nginx:v1
`

	_, err := lockconfig.NewBundleLockFromBytes([]byte(data))
	if err == nil {
		t.Fatalf("Expected non-nil error")
	}
	if !strings.Contains(err.Error(), "Expected ref to be in digest form, got 'nginx:v1'") {
		t.Fatalf("Expected err to check digest form, but err was: '%s'", err)
	}
}

func TestBundleLockWithUnknownKeys(t *testing.T) {
	data := `
apiVersion: imgpkg.carvel.dev/v1alpha1
kind: BundleLock
spec:
  image: nginx:v1
`

	_, err := lockconfig.NewBundleLockFromBytes([]byte(data))
	if err == nil {
		t.Fatalf("Expected non-nil error")
	}
	if !strings.Contains(err.Error(), `unknown field "spec"`) {
		t.Fatalf("Expected error for unknown key, got: %s", err)
	}
}
