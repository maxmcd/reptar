package reptar

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
)

func TestArchive(t *testing.T) {
	tests := []struct {
		name string
		in   []entry
	}{
		{
			name: "one file",
			in:   entries(file("foo")),
		},
		{
			name: "files and dir",
			in: entries(
				dir("thing"),
				file("foo"),
				emptyFile("empty"),
			),
		},
		{
			name: "nested files",
			in: entries(
				dir("thing"),
				file("foo"),
				file("thing/bar"),
			),
		},
		{
			name: "unusual suspects",
			in: entries(
				dir("thing"),
				file("thing/bar"),
				symlink("thing/bar", "symlink"),
				hardlink("thing/bar", "hardlink"),
				fifo("thing/pipe"),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFileDir := t.TempDir()
			for _, e := range tt.in {
				if err := e(testFileDir); err != nil {
					t.Fatal(err)
				}
			}
			checkDir(t, testFileDir)
		})
	}
}

func checkDir(t *testing.T, testDir string) {
	var buf bytes.Buffer
	if err := Archive(testDir, &buf); err != nil {
		t.Fatal(err)
	}
	copyOfOriginal := t.TempDir()
	if err := Unarchive(&buf, copyOfOriginal); err != nil {
		t.Fatal(err)
	}

	if hashDir(t, testDir) != hashDir(t, copyOfOriginal) {
		{
			cmd := "diff -qr " + testDir + " " + copyOfOriginal
			b, err := exec.Command("bash", "-c", cmd).CombinedOutput()
			fmt.Println(string(b))
			if err != nil {
				fmt.Println(cmd)
			}
			if err != nil {
				t.Fatal(err)
			}
		}
		cmd := "git diff --color=never --no-index " + testDir + " " + copyOfOriginal
		b, err := exec.Command("bash", "-c", cmd).CombinedOutput()
		if err != nil {
			fmt.Println(cmd)
			fmt.Println(string(b))
		}
		if err != nil {
			t.Fatal(err)
		}
	}
}

func hashDir(t *testing.T, location string) string {
	h := sha256.New()
	if err := Archive(location, h); err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

type entry func(dir string) error

func entries(v ...entry) []entry {
	return v
}

var j func(...string) string = filepath.Join

func file(name string) func(dir string) error {
	return func(dir string) error {
		f, err := os.Create(j(dir, name))
		if err != nil {
			return err
		}
		if _, err := io.CopyN(f, rand.Reader, 9e6); err != nil {
			return err
		}
		return f.Close()
	}
}

func emptyFile(name string) func(dir string) error {
	return func(dir string) error {
		f, err := os.Create(j(dir, name))
		if err != nil {
			return err
		}
		return f.Close()
	}
}

func dir(name string) func(dir string) error {
	return func(dir string) error {
		return os.Mkdir(j(dir, name), 0755)
	}
}

func hardlink(name, link string) func(dir string) error {
	return func(dir string) error {
		return os.Link(j(dir, name), j(dir, link))
	}
}

func symlink(name, link string) func(dir string) error {
	return func(dir string) error {
		return os.Symlink(j(dir, name), j(dir, link))
	}
}

func fifo(name string) func(dir string) error {
	return func(dir string) error {
		return syscall.Mkfifo(j(dir, name), 0755)
	}
}
