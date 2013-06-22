package fuse

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"
)

var _ = log.Println

const testTtl = 100 * time.Millisecond

func setupMemNodeTest(t *testing.T) (wd string, fs *MemNodeFs, clean func()) {
	tmp, err := ioutil.TempDir("", "go-fuse-memnode_test")
	if err != nil {
		t.Fatalf("TempDir failed: %v", err)
	}
	back := tmp + "/backing"
	os.Mkdir(back, 0700)
	fs = NewMemNodeFs(back)
	mnt := tmp + "/mnt"
	os.Mkdir(mnt, 0700)

	connector := NewFileSystemConnector(fs,
		&FileSystemOptions{
			EntryTimeout:    testTtl,
			AttrTimeout:     testTtl,
			NegativeTimeout: 0.0,
		})
	connector.SetDebug(VerboseTest())
	state := NewMountState(connector)
	state.Mount(mnt, nil)

	//me.state.SetDebug(false)
	state.SetDebug(VerboseTest())

	// Unthreaded, but in background.
	go state.Loop()
	return mnt, fs, func() {
		state.Unmount()
		os.RemoveAll(tmp)
	}

}

func TestMemNodeFsWrite(t *testing.T) {
	wd, _, clean := setupMemNodeTest(t)
	defer clean()

	want := "hello"
	err := ioutil.WriteFile(wd+"/test", []byte(want), 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	content, err := ioutil.ReadFile(wd + "/test")
	if string(content) != want {
		t.Fatalf("content mismatch: got %q, want %q", content, want)
	}
}

func TestMemNodeFs(t *testing.T) {
	wd, _, clean := setupMemNodeTest(t)
	defer clean()

	err := ioutil.WriteFile(wd+"/test", []byte{42}, 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	fi, err := os.Lstat(wd + "/test")
	if err != nil {
		t.Fatalf("Lstat failed: %v", err)
	}
	if fi.Size() != 1 {
		t.Errorf("Size after write incorrect: got %d want 1", fi.Size())
	}

	entries, err := ioutil.ReadDir(wd)
	if len(entries) != 1 || entries[0].Name() != "test" {
		t.Fatalf("Readdir got %v, expected 1 file named 'test'", entries)
	}
}

func TestMemNodeSetattr(t *testing.T) {
	wd, _, clean := setupMemNodeTest(t)
	defer clean()

	f, err := os.OpenFile(wd+"/test", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	defer f.Close()

	err = f.Truncate(4096)
	if err != nil {
		t.Fatalf("Truncate failed: %v", err)
	}

	fi, err := f.Stat()
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if fi.Size() != 4096 {
		t.Errorf("Size should be 4096 after Truncate: %d", fi.Size())
	}
}
