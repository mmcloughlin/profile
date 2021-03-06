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

		// All.
		{
			Name:    "all",
			Options: []func(*profile.Profile){profile.AllProfiles},
			Args: []string{
				"-cpuprofile=cpu.out",
				"-memprofile=mem.out",
				"-goroutineprofile=goroutine.out",
				"-threadcreateprofile=threadcreate.out",
				"-blockprofile=block.out",
				"-mutexprofile=mutex.out",
				"-trace=trace.out",
			},
			Files: []string{
				"cpu.out",
				"mem.out",
				"goroutine.out",
				"threadcreate.out",
				"block.out",
				"mutex.out",
				"trace.out",
			},
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

			// Confirm we have the files we expect.
			AssertDirContains(t, dir, c.Files)
		})
	}
}

func TestEnvConfiguration(t *testing.T) {
	dir := t.TempDir()
	Chdir(t, dir)

	// Set the environment variable for this test.
	key := "PROFILE"
	Setenv(t, key, "cpuprofile=cpu.out,memprofile=mem.out")

	// Run profiler.
	profile.Start(
		profile.AllProfiles,
		profile.ConfigEnvVar(key),
		profile.WithLogger(Logger(t)),
	).Stop()

	// Verify we have what we expect.
	AssertDirContains(t, dir, []string{"cpu.out", "mem.out"})
}

// TestEnvConfigurationEmpty is a regression test for the case where a
// configuration environment variable is specified but it's empty or unset.  In
// this case no profilers should be run.
func TestEnvConfigurationEmpty(t *testing.T) {
	dir := t.TempDir()
	Chdir(t, dir)

	// Use a key that should not be set. Bail in the absurdly unlikely case that
	// it is set.
	key := "PROFILE_VARIABLE_EMPTY"
	if os.Getenv(key) != "" {
		t.FailNow()
	}

	// Run profiler.
	profile.Start(
		profile.AllProfiles,
		profile.ConfigEnvVar(key),
		profile.WithLogger(Logger(t)),
	).Stop()

	// Verify the directory is empty.
	AssertDirContains(t, dir, nil)
}

// AssertDirContains asserts that dir contains non-empty files called filenames,
// and nothing else.
func AssertDirContains(t *testing.T, dir string, filenames []string) {
	t.Helper()

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
	sort.Strings(filenames)
	if !reflect.DeepEqual(got, filenames) {
		t.Logf("expect: %v", filenames)
		t.Logf("   got: %v", got)
		t.Error("unexpected file output")
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

// Setenv sets an environment variable for the duration of a test.
func Setenv(t *testing.T, key, value string) {
	t.Helper()

	prev, ok := os.LookupEnv(key)
	t.Cleanup(func() {
		if ok {
			if err := os.Setenv(key, prev); err != nil {
				t.Fatal(err)
			}
		} else {
			if err := os.Unsetenv(key); err != nil {
				t.Fatal(err)
			}
		}
	})

	if err := os.Setenv(key, value); err != nil {
		t.Fatal(err)
	}
}

// Logger builds a logger that writes to the test object.
func Logger(tb testing.TB) *log.Logger {
	tb.Helper()
	return log.New(Writer(tb), "test: ", 0)
}

type writer struct {
	tb testing.TB
}

// Writer builds a writer that logs all writes to the test object.
func Writer(tb testing.TB) io.Writer {
	tb.Helper()
	return writer{tb}
}

func (w writer) Write(p []byte) (n int, err error) {
	w.tb.Log(string(p))
	return len(p), nil
}
