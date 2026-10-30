package game

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

const installedVersionFile = ".yayeet-version"

var (
	lookPath     = exec.LookPath
	extractSeven = extractWith7z
)

func InstallFromArchive(ctx context.Context, archivePath, installationPath, version string, report ProgressFunc) error {
	version = strings.TrimSpace(version)
	if err := ValidateVersionName(version); err != nil {
		return err
	}
	if archivePath == "" {
		return errors.New("archive path is empty")
	}
	if installationPath == "" {
		return errors.New("installation path is empty")
	}
	if _, err := os.Stat(archivePath); err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	if err := os.MkdirAll(installationPath, 0o755); err != nil {
		return fmt.Errorf("create installation directory: %w", err)
	}

	tempDir, err := os.MkdirTemp("", "yayeet-local-*")
	if err != nil {
		return fmt.Errorf("create temporary extract directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	if err := extractSeven(ctx, archivePath, tempDir, report); err != nil {
		return err
	}

	source, err := findWindowsNoEditor(tempDir)
	if err != nil {
		return err
	}
	if err := replaceWindowsNoEditor(installationPath, source); err != nil {
		return err
	}
	if err := writeInstalledVersion(installationPath, version); err != nil {
		return err
	}
	if report != nil {
		report(InstallProgress{Current: 1, Total: 1})
	}
	return nil
}

func ValidateVersionName(version string) error {
	if version == "" {
		return errors.New("version is empty")
	}
	if filepath.Base(version) != version || strings.ContainsAny(version, `/\`) {
		return errors.New("version must be a plain name like 0100_0009")
	}
	return nil
}

func VersionHintFromArchiveName(archivePath string) string {
	base := filepath.Base(archivePath)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	if err := ValidateVersionName(base); err != nil {
		return ""
	}
	return base
}

func writeInstalledVersion(installationPath, version string) error {
	if err := ValidateVersionName(version); err != nil {
		return err
	}
	path := filepath.Join(installationPath, installedVersionFile)
	if err := os.WriteFile(path, []byte(version+"\n"), 0o644); err != nil {
		return fmt.Errorf("write installed version marker: %w", err)
	}
	return nil
}

func readInstalledVersion(installationPath string) (string, error) {
	data, err := os.ReadFile(filepath.Join(installationPath, installedVersionFile))
	if err != nil {
		return "", err
	}
	version := strings.TrimSpace(string(data))
	if err := ValidateVersionName(version); err != nil {
		return "", err
	}
	return version, nil
}

func findWindowsNoEditor(root string) (string, error) {
	var found string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() || entry.Name() != "WindowsNoEditor" {
			return nil
		}
		executable := filepath.Join(path, "VotV.exe")
		if _, err := os.Stat(executable); err != nil {
			return nil
		}
		if found != "" {
			return errors.New("archive contains multiple WindowsNoEditor game directories")
		}
		found = path
		return filepath.SkipDir
	})
	if err != nil {
		return "", fmt.Errorf("scan extracted archive: %w", err)
	}
	if found == "" {
		return "", errors.New("archive does not contain WindowsNoEditor/VotV.exe")
	}
	return found, nil
}

func replaceWindowsNoEditor(installationPath, source string) error {
	dest := filepath.Join(installationPath, "WindowsNoEditor")
	backup := dest + ".yayeet-old"
	_ = os.RemoveAll(backup)

	if _, err := os.Stat(dest); err == nil {
		if err := os.Rename(dest, backup); err != nil {
			return fmt.Errorf("move existing game files aside: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat existing game files: %w", err)
	}

	if err := moveDirectory(source, dest); err != nil {
		if _, backupErr := os.Stat(backup); backupErr == nil {
			_ = os.RemoveAll(dest)
			_ = os.Rename(backup, dest)
		}
		return err
	}
	_ = os.RemoveAll(backup)
	return nil
}

func moveDirectory(source, dest string) error {
	if err := os.Rename(source, dest); err == nil {
		return nil
	} else if !isCrossDevice(err) {
		return fmt.Errorf("move game files into installation: %w", err)
	}
	if err := copyDirectory(source, dest); err != nil {
		_ = os.RemoveAll(dest)
		return err
	}
	if err := os.RemoveAll(source); err != nil {
		return fmt.Errorf("remove temporary game files: %w", err)
	}
	return nil
}

func isCrossDevice(err error) bool {
	return errors.Is(err, syscall.EXDEV)
}

func copyDirectory(source, dest string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dest, rel)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return os.MkdirAll(target, info.Mode().Perm())
		}
		if entry.Type()&fs.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(link, target)
		}
		return copyFile(path, target, info.Mode().Perm())
	})
}

func copyFile(source, dest string, mode fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func extractWith7z(ctx context.Context, archivePath, dest string, report ProgressFunc) error {
	sevenZ, err := lookPath("7z")
	if err != nil {
		sevenZ, err = lookPath("7za")
	}
	if err != nil {
		return errors.New("7z is required to install a local game archive; install p7zip or 7zip and try again")
	}

	cmd := exec.CommandContext(ctx, sevenZ, "x", "-y", "-bsp1", "-bse1", "-o"+dest, archivePath)
	output, writer := io.Pipe()
	cmd.Stdout = writer
	cmd.Stderr = writer
	if err := cmd.Start(); err != nil {
		_ = writer.Close()
		return fmt.Errorf("start 7z: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		scanner := bufio.NewScanner(output)
		scanner.Split(scanLinesOrReturns)
		for scanner.Scan() {
			percent, ok := parseSevenZipPercent(scanner.Text())
			if !ok || report == nil {
				continue
			}
			report(InstallProgress{Current: percent, Total: 100})
		}
		done <- scanner.Err()
	}()

	waitErr := cmd.Wait()
	_ = writer.Close()
	scanErr := <-done
	if waitErr != nil {
		return fmt.Errorf("extract archive with 7z: %w", waitErr)
	}
	if scanErr != nil {
		return fmt.Errorf("read 7z output: %w", scanErr)
	}
	return nil
}

func scanLinesOrReturns(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	for index, value := range data {
		if value == '\n' || value == '\r' {
			return index + 1, data[:index], nil
		}
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func parseSevenZipPercent(line string) (int, bool) {
	line = strings.TrimSpace(line)
	if !strings.HasSuffix(line, "%") {
		return 0, false
	}
	number := strings.TrimSpace(strings.TrimSuffix(line, "%"))
	percent, err := strconv.Atoi(number)
	if err != nil || percent < 0 || percent > 100 {
		return 0, false
	}
	return percent, true
}
