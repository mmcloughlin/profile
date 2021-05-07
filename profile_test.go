package profile_test

import (
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/mmcloughlin/profile"
)

func TestFlagsConfiguration(t *testing.T) {
	cases := []struct {
		Name       string
		Options    []func(*profile.Profile)
		Args       []string
		ParseError bool
		Files      []string
	}{
		// Each method on its own.
		{
			Name:    "cpu",
			Options: []func(*profile.Profile){profile.CPUProfile},
			Args:    []string{"-cpuprofile=cpu.out"},
			Files:   []string{"cpu.out"},
		},
		{
			Name:    "mem",
			Options: []func(*profile.Profile){profile.MemProfile},
			Args:    []string{"-memprofile=mem.out"},
			Files:   []string{"mem.out"},
		},
		{
			Name:    "goroutine",
			Options: []func(*profile.Profile){profile.GoroutineProfile},
			Args:    []string{"-goroutineprofile=goroutine.out"},
			Files:   []string{"goroutine.out"},
		},
		{
			Name:    "threadcreateprofile",
			Options: []func(*profile.Profile){profile.ThreadcreationProfile},
			Args:    []string{"-threadcreateprofile=threadcreate.out"},
			Files:   []string{"threadcreate.out"},
		},
		{
			Name:    "block",
			Options: []func(*profile.Profile){profile.BlockProfile},
			Args:    []string{"-blockprofile=block.out"},
			Files:   []string{"block.out"},
		},
		{
			Name:    "mutex",
			Options: []func(*profile.Profile){profile.MutexProfile},
			Args:    []string{"-mutexprofile=mutex.out"},
			Files:   []string{"mutex.out"},
		},
		{
			Name:    "trace",
			Options: []func(*profile.Profile){profile.TraceProfile},
			Args:    []string{"-trace=trace.out"},
			Files:   []string{"trace.out"},
		},

		// Defaults: when no options are provided.
		{
			Name: "default_noargs",
		},
		{
			Name:  "default_cpu",
			Args:  []string{"-cpuprofile=cpu.out"},
			Files: []string{"cpu.out"},
		},

		// Multi-mode profiling.
		{
			Name:    "multi_enable_both",
			Options: []func(*profile.Profile){profile.CPUProfile, profile.MemProfile},
			Args:    []string{"-cpuprofile=cpu.out", "-memprofile=mem.out"},
			Files:   []string{"cpu.out", "mem.out"},
		},
		{
			Name:    "multi_enable_one",
			Options: []func(*profile.Profile){profile.CPUProfile, profile.MemProfile},
			Args:    []string{"-memprofile=mem.out"},
			Files:   []string{"mem.out"},
		},
	}
	for _, c := range cases {
		c := c // scopelint
		t.Run(c.Name, func(t *testing.T) {
			dir := t.TempDir()
			Chdir(t, dir)

			// Initialize.
			p := profile.New(c.Options...)
			p.Configure(profile.WithLogger(Logger(t)))

			// Configure via flags.
			f := flag.NewFlagSet("profile", flag.ContinueOnError)
			p.SetFlags(f)

			err := f.Parse(c.Args)
			if (err != nil) != c.ParseError {
				t.Logf("expected parse error: %v", c.ParseError)
				t.Logf("got: %v", err)
				t.FailNow()
			}

			// Run.
			p.Start().Stop()

			// Confirm we have the files we expect, and nothing else.
			entries, err := ioutil.ReadDir(dir)
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			for _, entry := range entries {
				if !entry.Mode().IsRegular() {
					t.Errorf("%s is not regular file", entry.Name())
				}
				if entry.Size() == 0 {
					t.Errorf("file %v is empty", entry.Name())
				}
				got = append(got, entry.Name())
			}

			sort.Strings(got)
			sort.Strings(c.Files)
			if !reflect.DeepEqual(got, c.Files) {
				t.Logf("expect: %v", c.Files)
				t.Logf("   got: %v", got)
				t.Error("unexpected file output")
			}
		})
	}
}

// Chdir changes into a given directory for the duration of a test.
func Chdir(t *testing.T, dir string) {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatal(err)
		}
	})
}

// Logger builds a logger that writes to the test object.
func Logger(tb testing.TB) *log.Logger {
	return log.New(Writer(tb), "test: ", 0)
}

type writer struct {
	tb testing.TB
}

// Writer builds a writer that logs all writes to the test object.
func Writer(tb testing.TB) io.Writer {
	return writer{tb}
}

func (w writer) Write(p []byte) (n int, err error) {
	w.tb.Log(string(p))
	return len(p), nil
}
