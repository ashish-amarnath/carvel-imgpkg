// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPushPull(t *testing.T) {
	env := BuildEnv(t)
	imgpkg := Imgpkg{t, Logger{}, env.ImgpkgPath}

	assetsPath := filepath.Join("assets", "simple-app")
	path := filepath.Join(os.TempDir(), "imgpkg-test-basic")

	cleanUp := func() { os.RemoveAll(path) }
	cleanUp()
	defer cleanUp()

	imgpkg.Run([]string{"push", "-i", env.Image, "-f", assetsPath})
	imgpkg.Run([]string{"pull", "-i", env.Image, "-o", path})

	expectedFiles := []string{
		"README.md",
		"LICENSE",
		"config/config.yml",
		"config/inner-dir/README.txt",
	}

	for _, file := range expectedFiles {
		compareFiles(filepath.Join(assetsPath, file), filepath.Join(path, file), t)
	}
}

func TestPushMultipleFiles(t *testing.T) {
	env := BuildEnv(t)
	imgpkg := Imgpkg{t, Logger{}, env.ImgpkgPath}

	assetsPath := filepath.Join("assets", "simple-app")
	path := filepath.Join(os.TempDir(), "imgpkg-test-push-multiple-files")

	cleanUp := func() { os.RemoveAll(path) }
	cleanUp()
	defer cleanUp()

	imgpkg.Run([]string{
		"push", "-i", env.Image,
		"-f", filepath.Join(assetsPath, "LICENSE"),
		"-f", filepath.Join(assetsPath, "README.md"),
		"-f", filepath.Join(assetsPath, "config"),
	})

	imgpkg.Run([]string{"pull", "-i", env.Image, "-o", path})

	expectedFiles := map[string]string{
		"README.md":                   "README.md",
		"LICENSE":                     "LICENSE",
		"config/config.yml":           "config.yml",
		"config/inner-dir/README.txt": "inner-dir/README.txt",
	}

	for assetFile, downloadedFile := range expectedFiles {
		compareFiles(filepath.Join(assetsPath, assetFile), filepath.Join(path, downloadedFile), t)
	}
}
