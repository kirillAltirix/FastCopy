package fastcopy

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sync"
	"sync/atomic"
	"time"
)

func Copy(srcDir, dstDir, filter string) (int, error) {
	// create regex files filter
	regFilter := regexp.MustCompile(filter)

	// process files in destination dir
	m := make(map[string]time.Time)
	if err := initFilesMapFromDst(dstDir, regFilter, &m); err != nil {
		slog.Error("Processing files in destination path failed", slog.String("error", err.Error()))
		return 0, err
	}

	res, err := copy(srcDir, dstDir, regFilter, &m)
	if err != nil {
		slog.Error("Copying failed", slog.String("error", err.Error()))
	}
	return res, err
}

func copyFile(srcPath, dstPath, filename string) error {
	f1, _ := filepath.Abs(srcPath + "/" + filename)
	f2, _ := filepath.Abs(dstPath + "/" + filename)
	cmd := exec.Command("cmd", "/C", "copy "+fmt.Sprintf("%v", f1)+" "+fmt.Sprintf("%v", f2))
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func initFilesMapFromDst(dstDir string, regFilter *regexp.Regexp, m *map[string]time.Time) error {
	filesTo, err := os.ReadDir(dstDir)
	if err != nil {
		return err
	}

	for _, file := range filesTo {
		if regFilter.Match([]byte(file.Name())) {
			info, _ := file.Info()
			(*m)[info.Name()] = info.ModTime()
		}
	}
	return nil
}

func copy(srcDir, dstDir string, regFilter *regexp.Regexp, m *map[string]time.Time) (int, error) {
	filesFrom, err := os.ReadDir(srcDir)
	if err != nil {
		return 0, err
	}

	var numFilesCopied int32

	wg := sync.WaitGroup{}
	for _, file := range filesFrom {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// check that file matches users filter
			if !regFilter.Match([]byte(file.Name())) {
				return
			}
			info, _ := file.Info()
			// check that the file was really changes using modification time
			if modTime, ok := (*m)[info.Name()]; ok && modTime == info.ModTime() {
				return
			}
			// do copy
			if err := copyFile(srcDir, dstDir, info.Name()); err != nil {
				slog.Error("Copy file failed", slog.String("file name", info.Name()), slog.String("error", err.Error()))
				return
			}
			atomic.AddInt32(&numFilesCopied, 1)
		}()
	}
	wg.Wait()

	return int(numFilesCopied), nil
}
