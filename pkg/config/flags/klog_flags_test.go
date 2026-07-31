/*
Copyright 2025 The Kubernetes Authors.

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

package flags

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

// --------------------- HELPERS ---------------------

func newTestFlagSet(t *testing.T, logFile string, logToStderr, alsoLogToStderr bool) *pflag.FlagSet {
	t.Helper()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.String("log-file", "", "")
	fs.Bool("logtostderr", true, "")
	fs.Bool("alsologtostderr", false, "")

	if logFile != "" {
		if err := fs.Set("log-file", logFile); err != nil {
			t.Fatalf("set log-file: %v", err)
		}
	}
	if err := fs.Set("logtostderr", boolToString(logToStderr)); err != nil {
		t.Fatalf("set logtostderr: %v", err)
	}
	if err := fs.Set("alsologtostderr", boolToString(alsoLogToStderr)); err != nil {
		t.Fatalf("set alsologtostderr: %v", err)
	}
	return fs
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// assertStreams checks the actual streams against expected values
func assertStreams(t *testing.T, actualInfo, actualErr any, expectedInfo, expectedErr any) {
	t.Helper()

	check := func(name string, actual, expected any) {
		if expected == nil {
			if actual != nil {
				t.Fatalf("%s: expected nil, got %T", name, actual)
			}
			return
		}

		if reflect.TypeOf(actual) != reflect.TypeOf(expected) {
			t.Fatalf("%s type mismatch: expected %T, got %T", name, expected, actual)
		}

		if f, ok := expected.(*os.File); ok {
			if actual.(*os.File) != f {
				t.Fatalf("%s pointer mismatch: expected same *os.File", name)
			}
		}
		if expected == os.Stderr {
			if actual != os.Stderr {
				t.Fatalf("%s mismatch: expected os.Stderr", name)
			}
		}

		// For io.MultiWriter, type check is sufficient
	}

	check("InfoStream", actualInfo, expectedInfo)
	check("ErrorStream", actualErr, expectedErr)
}

// --------------------- TESTS ---------------------

func TestComputeLoggingOptions_DefaultToStderr(t *testing.T) {
	fs := newTestFlagSet(t, "", true, false)
	opts, err := ComputeLoggingOptions(fs)
	if err != nil {
		t.Fatalf("ComputeLoggingOptions error: %v", err)
	}
	assertStreams(t, opts.InfoStream, opts.ErrorStream, os.Stderr, os.Stderr)
}

func TestComputeLoggingOptions_LogFileIgnoredWhenLogToStderrTrue(t *testing.T) {
	tmp := t.TempDir()
	logPath := filepath.Join(tmp, "subdir", "ca.log")

	fs := newTestFlagSet(t, logPath, true, false)
	opts, err := ComputeLoggingOptions(fs)
	if err != nil {
		t.Fatalf("ComputeLoggingOptions error: %v", err)
	}

	if _, statErr := os.Stat(logPath); !os.IsNotExist(statErr) {
		t.Fatalf("expected no log file created at %q, statErr=%v", logPath, statErr)
	}

	assertStreams(t, opts.InfoStream, opts.ErrorStream, os.Stderr, os.Stderr)
}

func TestComputeLoggingOptions_LogFileOnly(t *testing.T) {
	tmp := t.TempDir()
	logPath := filepath.Join(tmp, "nested", "onlyfile.log")

	fs := newTestFlagSet(t, logPath, false, false)
	opts, err := ComputeLoggingOptions(fs)
	if err != nil {
		t.Fatalf("ComputeLoggingOptions error: %v", err)
	}

	if _, statErr := os.Stat(logPath); statErr != nil {
		t.Fatalf("expected log file created at %q, err=%v", logPath, statErr)
	}

	// Streams should be the same file writer
	file := opts.InfoStream.(*os.File)
	assertStreams(t, opts.InfoStream, opts.ErrorStream, file, file)

	// Write and verify content
	msg := "hello-file-only\n"
	if _, err := file.Write([]byte(msg)); err != nil {
		t.Fatalf("write to file failed: %v", err)
	}
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log file failed: %v", err)
	}
	if !strings.Contains(string(data), msg) {
		t.Fatalf("log file does not contain expected content")
	}
}

func TestComputeLoggingOptions_LogFileAlsoToStderr(t *testing.T) {
	tmp := t.TempDir()
	logPath := filepath.Join(tmp, "withstderr", "combo.log")

	fs := newTestFlagSet(t, logPath, false, true)
	opts, err := ComputeLoggingOptions(fs)
	if err != nil {
		t.Fatalf("ComputeLoggingOptions error: %v", err)
	}

	// File must exist
	if _, statErr := os.Stat(logPath); statErr != nil {
		t.Fatalf("expected log file created at %q, err=%v", logPath, statErr)
	}

	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("failed to open log file to build expected writer: %v", err)
	}
	defer file.Close()

	expectedWriter := io.MultiWriter(os.Stderr, file)

	// Assert streams against expected MultiWriter
	assertStreams(t, opts.InfoStream, opts.ErrorStream, expectedWriter, expectedWriter)

	// Write and verify content appears in the file
	msg := "hello-also-stderr\n"
	if _, err := opts.InfoStream.Write([]byte(msg)); err != nil {
		t.Fatalf("write to InfoStream failed: %v", err)
	}
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log file failed: %v", err)
	}
	if !strings.Contains(string(data), msg) {
		t.Fatalf("log file does not contain expected content")
	}
}

func TestComputeLoggingOptions_CreateDirError(t *testing.T) {
	tmp := t.TempDir()
	notADirPath := filepath.Join(tmp, "notadir")
	if err := os.WriteFile(notADirPath, []byte("x"), 0o644); err != nil {
		t.Fatalf("prepare file failed: %v", err)
	}
	logPath := filepath.Join(notADirPath, "child", "ca.log")

	fs := newTestFlagSet(t, logPath, false, false)
	_, err := ComputeLoggingOptions(fs)
	if err == nil || !strings.Contains(err.Error(), "failed to create log directory") {
		t.Fatalf("expected create dir error, got: %v", err)
	}
}

func TestComputeLoggingOptions_OpenFileError(t *testing.T) {
	tmp := t.TempDir()
	dirAsFile := filepath.Join(tmp, "adir")
	if err := os.MkdirAll(dirAsFile, 0o755); err != nil {
		t.Fatalf("prepare dir failed: %v", err)
	}

	fs := newTestFlagSet(t, dirAsFile, false, false)
	_, err := ComputeLoggingOptions(fs)
	if err == nil || !strings.Contains(err.Error(), "failed to open log file") {
		t.Fatalf("expected open file error, got: %v", err)
	}
}

func TestComputeLoggingOptions_WriterInterfaceUsable(t *testing.T) {
	tmp := t.TempDir()
	logPath := filepath.Join(tmp, "writeusable", "ca.log")

	fs := newTestFlagSet(t, logPath, false, false)
	opts, err := ComputeLoggingOptions(fs)
	if err != nil {
		t.Fatalf("ComputeLoggingOptions error: %v", err)
	}

	file := opts.ErrorStream.(*os.File)
	n, err := file.Write([]byte("err-stream\n"))
	if err != nil || n == 0 {
		t.Fatalf("failed writing via ErrorStream: n=%d err=%v", n, err)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log file failed: %v", err)
	}
	if !strings.Contains(string(data), "err-stream") {
		t.Fatalf("log file does not contain expected err-stream content")
	}
}
