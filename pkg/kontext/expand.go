/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kontext

import (
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
)

const (
	// StoragePath is where in the container image the files are placed.
	StoragePath = "/var/run/kontext"
)

func copy(src, dest string, info os.FileInfo) error {
	from, err := os.Open(src)
	if err != nil {
		return err
	}
	defer from.Close()

	to, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, info.Mode())
	if err != nil {
		return err
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	if err != nil {
		return err
	}
	return nil
}

func Expand(ctx context.Context) error {
	targetPath, err := os.Getwd()
	if err != nil {
		return err
	}

	return filepath.Walk(StoragePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// If the context is canceled, then bail out early.
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if path == StoragePath {
			return nil
		}
		relativePath := path[len(StoragePath)+1:]
		target := filepath.Join(targetPath, relativePath)

		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		if !info.Mode().IsRegular() {
			log.Printf("Skipping irregular file: %q", relativePath)
			return nil
		}
		if err := os.MkdirAll(filepath.Dir(target), 0444); err != nil {
			return err
		}
		return copy(path, target, info)
	})
}
