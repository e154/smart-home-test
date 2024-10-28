// This file is part of the Smart Home
// Program complex distribution https://github.com/e154/smart-home
// Copyright (C) 2024, Filippov Alex
//
// This library is free software: you can redistribute it and/or
// modify it under the terms of the GNU Lesser General Public
// License as published by the Free Software Foundation; either
// version 3 of the License, or (at your option) any later version.
//
// This library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// Library General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public
// License along with this library.  If not, see
// <https://www.gnu.org/licenses/>.

package common

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

func ExtractTarGz(gzipStream io.Reader, target string) error {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return err
	}

	if err = os.MkdirAll(target, 0755); err != nil {
		return errors.Wrap(fmt.Errorf("failed create dir %s: ", target), err.Error())
	}

	tarReader := tar.NewReader(uncompressedStream)

	for {
		file, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return errors.Wrap(fmt.Errorf("extract tar failed: "), err.Error())
		}

		path := filepath.Join(target, file.Name)

		switch file.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(path, 0755); err != nil {
				return errors.Wrap(fmt.Errorf("extract mkdir failed"), err.Error())
			}
		case tar.TypeReg:

			targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, file.FileInfo().Mode())
			if err != nil {
				return errors.Wrap(fmt.Errorf("open file failed"), err.Error())
			}
			defer targetFile.Close()

			if _, err = io.Copy(targetFile, tarReader); err != nil {
				return errors.Wrap(fmt.Errorf("copy file failed"), err.Error())
			}
		}

	}

	return nil
}
