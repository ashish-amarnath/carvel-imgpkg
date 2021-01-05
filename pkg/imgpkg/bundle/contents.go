package bundle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cppforlife/go-cli-ui/ui"
	regname "github.com/google/go-containerregistry/pkg/name"
	ctlimg "github.com/k14s/imgpkg/pkg/imgpkg/image"
	"github.com/k14s/imgpkg/pkg/imgpkg/lockconfig"
	lf "github.com/k14s/imgpkg/pkg/imgpkg/lockfiles"
	"github.com/k14s/imgpkg/pkg/imgpkg/plainimage"
)

const (
	ImgpkgDir      = ".imgpkg"
	ImagesLockFile = "images.yml"
)

type Contents struct {
	paths         []string
	excludedPaths []string
}

func NewContents(paths []string, excludedPaths []string) Contents {
	return Contents{paths: paths, excludedPaths: excludedPaths}
}

func (b Contents) Push(uploadRef regname.Tag, registry ctlimg.Registry, ui ui.UI) (string, error) {
	err := b.validate(registry)
	if err != nil {
		return "", err
	}

	labels := map[string]string{BundleConfigLabel: "true"}
	return plainimage.NewContents(b.paths, b.excludedPaths).Push(uploadRef, labels, registry, ui)
}

func (b Contents) PresentsAsBundle() (bool, error) {
	imgpkgDirs, err := b.findImgpkgDirs()
	if err != nil {
		return false, err
	}

	err = b.validateImgpkgDirs(imgpkgDirs)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (b Contents) validate(registry ctlimg.Registry) error {
	imgpkgDirs, err := b.findImgpkgDirs()
	if err != nil {
		return err
	}

	err = b.validateImgpkgDirs(imgpkgDirs)
	if err != nil {
		return err
	}

	imagesLock, err := lockconfig.NewImagesLockFromPath(filepath.Join(imgpkgDirs[0], lf.ImageLockFile))
	if err != nil {
		return err
	}

	bundles, err := b.checkForBundles(registry, imagesLock.Images)
	if err != nil {
		return fmt.Errorf("Checking image lock for bundles: %s", err)
	}

	if len(bundles) != 0 {
		return fmt.Errorf("Expected image lock to not contain bundle reference: '%v'", strings.Join(bundles, "', '"))
	}

	return nil
}

func (b Contents) checkForBundles(reg ctlimg.Registry, imageRefs []lockconfig.ImageRef) ([]string, error) {
	var bundles []string
	for _, img := range imageRefs {
		imgRef := img.Image
		parsedRef, err := regname.ParseReference(imgRef)
		if err != nil {
			return nil, err
		}
		image, err := reg.Image(parsedRef)
		if err != nil {
			return nil, err
		}

		isBundle, err := lf.IsBundle(image) // TODO please come back to me
		if err != nil {
			return nil, err
		}

		if isBundle {
			bundles = append(bundles, imgRef)
		}
	}
	return bundles, nil
}

func (b *Contents) findImgpkgDirs() ([]string, error) {
	var bundlePaths []string
	for _, path := range b.paths {
		err := filepath.Walk(path, func(currPath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if filepath.Base(currPath) != lf.BundleDir {
				return nil
			}

			currPath, err = filepath.Abs(currPath)
			if err != nil {
				return err
			}

			bundlePaths = append(bundlePaths, currPath)

			return nil
		})

		if err != nil {
			return []string{}, err
		}
	}

	return bundlePaths, nil
}

func (b Contents) validateImgpkgDirs(imgpkgDirs []string) error {
	if len(imgpkgDirs) != 1 {
		return fmt.Errorf("Expected one '%s' dir, got %d: %s", lf.BundleDir, len(imgpkgDirs), strings.Join(imgpkgDirs, ", "))
	}

	path := imgpkgDirs[0]

	// make sure it is a child of one input dir
	for _, flagPath := range b.paths {
		flagPath, err := filepath.Abs(flagPath)
		if err != nil {
			return err
		}

		if filepath.Dir(path) == flagPath {
			return nil
		}
	}

	return fmt.Errorf("Expected '%s' directory, to be a direct child of one of: %s; was %s", lf.BundleDir, strings.Join(b.paths, ", "), path)
}
