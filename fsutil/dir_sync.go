package fsutil

import (
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"

	xxhash "github.com/OneOfOne/xxhash"
	"github.com/pkg/errors"
)

func doCopyFile(srcDir string, srcFi os.FileInfo, destPath string) error {
	srcFile, err := os.Open(filepath.Join(srcDir, srcFi.Name()))
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	if runtime.GOOS != "windows" {
		if err := dstFile.Chmod(srcFi.Mode()); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(srcDir string, srcFi os.FileInfo, destPath string) error {
	err := doCopyFile(srcDir, srcFi, destPath)
	if err != nil {
		return err
	}
	return os.Chtimes(destPath, srcFi.ModTime(), srcFi.ModTime())
}

func mkDir(dirPath string, dirMode os.FileMode, dirModTime time.Time) error {
	if err := os.Mkdir(dirPath, dirMode); err != nil {
		return err
	}
	return os.Chtimes(dirPath, dirModTime, dirModTime)
}

func hashFile(srcFilePath string) (uint64, error) {

	h := xxhash.New64()
	srcFile, err := os.Open(srcFilePath)
	if err != nil {
		return 0, err
	}
	defer srcFile.Close()

	if _, err = io.Copy(h, srcFile); err != nil {
		return 0, err
	}
	return h.Sum64(), nil

}

func syncFolder(folderAbsPath, folderRelativePath, destDir string, destCreated, mirror bool) error {
	dirEntry, err := os.ReadDir(folderAbsPath)
	if err != nil {
		return err
	}

	destAbsPath := filepath.Join(destDir, folderRelativePath)
	destFilesMap := make(map[string]os.FileInfo, len(dirEntry))
	fileInfos := make([]os.FileInfo, 0)

	for _, entry := range dirEntry {
		fileInfo, err := entry.Info()
		if err != nil {
			return err
		}
		fileInfos = append(fileInfos, fileInfo)
	}

	if !destCreated {
		_, err := os.Stat(destAbsPath)
		if err != nil && !os.IsNotExist(err) {
			return err
		} else if err == nil {
			destFileInfos, err := os.ReadDir(destAbsPath)
			if err != nil {
				return err
			}
			for _, fi := range destFileInfos {
				fileInfo, err := fi.Info()
				if err != nil {
					return err
				}
				destFilesMap[filepath.Base(fi.Name())] = fileInfo
			}
		} else {
			srcStats, err := os.Stat(folderAbsPath)
			if err != nil {
				return err
			}
			if err := mkDir(destAbsPath, srcStats.Mode(), srcStats.ModTime()); err != nil {
				return err
			}
		}
	}

	destFilesMap, err = syncFolderContent(fileInfos, destFilesMap, destAbsPath, folderRelativePath, folderAbsPath, destDir, mirror)
	if err != nil {
		return err
	}

	if mirror {
		for _, v := range destFilesMap {
			if err := os.RemoveAll(path.Join(destAbsPath, v.Name())); err != nil {
				return err
			}
		}
	}

	return nil

}

func syncFolderContent(fileInfos []fs.FileInfo, destFilesMap map[string]fs.FileInfo, destAbsPath, folderRelativePath,
	folderAbsPath, destDir string, mirror bool) (map[string]fs.FileInfo, error) {
	for _, srcFi := range fileInfos {
		srcName := srcFi.Name()
		destFi := destFilesMap[srcName]
		destPath := filepath.Join(destAbsPath, srcName)
		srcRelPath := filepath.Join(folderRelativePath, srcName)
		srcAbsPath := filepath.Join(folderAbsPath, srcName)

		if destFi == nil {
			if srcFi.IsDir() {
				if err := mkDir(destPath, srcFi.Mode(), srcFi.ModTime()); err != nil {
					return destFilesMap, err
				}
				err := syncFolder(srcAbsPath, srcRelPath, destDir, true, mirror)
				if err != nil {
					return destFilesMap, err
				}
			} else if err := copyFile(folderAbsPath, srcFi, destPath); err != nil {
				return destFilesMap, err
			}
		} else {
			if srcFi.IsDir() {
				err := syncFolder(srcAbsPath, srcRelPath, destDir, false, mirror)
				if err != nil {
					return destFilesMap, err
				}
			} else {
				if srcFi.Size() != destFi.Size() {
					if err := copyFile(folderAbsPath, srcFi, destPath); err != nil {
						return destFilesMap, err
					}
				} else {
					srcHash, err := hashFile(srcAbsPath)
					if err != nil {
						return destFilesMap, err
					}
					dstHash, err := hashFile(destPath)
					if err != nil {
						return destFilesMap, err
					}
					if srcHash != dstHash {
						err = copyFile(srcAbsPath, srcFi, destPath)
						if err != nil {
							return destFilesMap, err
						}
					}
				}
			}
			delete(destFilesMap, srcName)
		}
	}
	return destFilesMap, nil
}

func Sync(sourceDirs []string, destDir string, mirror bool) error {

	sourceDirsInfo := make([]os.FileInfo, len(sourceDirs))
	for idx, sDir := range sourceDirs {
		sfi, err := os.Stat(sDir)
		if err != nil {
			return errors.Wrapf(err, "source directory %s does not exist, cannot continue", sDir)
		}
		sourceDirsInfo[idx] = sfi
	}

	_, err := os.Stat(destDir)
	if err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(destDir, 0755); err != nil {
			return errors.Wrapf(err, "cannot create destination directory %s", destDir)
		}
	} else if err != nil {
		return errors.Wrapf(err, "cannot access destination directory %s", destDir)
	}

	for _, sDir := range sourceDirs {
		if err = syncFolder(sDir, filepath.Base(sDir), destDir, false, mirror); err != nil {
			return errors.Wrapf(err, "cannot sync directory %s", sDir)
		}
	}

	return nil
}
