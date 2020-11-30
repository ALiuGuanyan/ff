package ff_test

import (
	"context"
	"github.com/ALiuGuanyan/ff/v3"
	"github.com/ALiuGuanyan/ff/v3/ffcli"
	"github.com/ALiuGuanyan/ff/v3/fftest"
	"github.com/ALiuGuanyan/pflag"
	"os"
	"testing"
	"time"
)

func TestParseBasics(t *testing.T) {
	t.Parallel()

	for _, testcase := range []struct {
		name string
		env  map[string]string
		file string
		args []string
		opts []ff.Option
		want fftest.Vars
	}{
		{
			name: "empty",
			args: []string{},
			want: fftest.Vars{},
		},
		{
			name: "args only",
			args: []string{"-s", "foo", "-i", "123", "-b", "-d", "24m"},
			want: fftest.Vars{Str: "foo", Int: 123, Bool: true, Duration: 24 * time.Minute},
		},
		{
			name: "file only",
			file: "testdata/1.conf",
			want: fftest.Vars{Str: "bar", Int: 99, Bool: true, Duration: time.Hour},
		},
		{
			name: "env only",
			env:  map[string]string{"TEST_PARSE_STR": "baz", "TEST_PARSE_FLOAT": "0.99", "TEST_PARSE_DURATION": "100s"},
			opts: []ff.Option{ff.WithEnvVarPrefix("TEST_PARSE")},
			want: fftest.Vars{Str: "baz", Float: 0.99, Duration: 100 * time.Second},
		},
		{
			name: "file args",
			file: "testdata/2.conf",
			args: []string{"-s", "foo", "-i", "1234"},
			want: fftest.Vars{Str: "foo", Int: 1234, Duration: 3 * time.Second},
		},
		{
			name: "env args",
			env:  map[string]string{"TEST_PARSE_STR": "should be overridden", "TEST_PARSE_BOOL": "true"},
			args: []string{"-s", "explicit wins", "-i", "7"},
			opts: []ff.Option{ff.WithEnvVarPrefix("TEST_PARSE")},
			want: fftest.Vars{Str: "explicit wins", Int: 7, Bool: true},
		},
		{
			name: "file env",
			env:  map[string]string{"TEST_PARSE_STR": "env takes priority", "TEST_PARSE_BOOL": "true"},
			file: "testdata/3.conf",
			opts: []ff.Option{ff.WithEnvVarPrefix("TEST_PARSE")},
			want: fftest.Vars{Str: "env takes priority", Int: 99, Bool: true, Duration: 34 * time.Second},
		},
		{
			name: "file env args",
			file: "testdata/4.conf",
			env:  map[string]string{"TEST_PARSE_STR": "from env", "TEST_PARSE_INT": "300", "TEST_PARSE_FLOAT": "0.15", "TEST_PARSE_BOOL": "true"},
			args: []string{"-s", "from arg", "-i", "100"},
			opts: []ff.Option{ff.WithEnvVarPrefix("TEST_PARSE")},
			want: fftest.Vars{Str: "from arg", Int: 100, Float: 0.15, Bool: true, Duration: time.Minute},
		},
		{
			name: "repeated args",
			args: []string{"-s", "foo", "-s", "bar", "-d", "1m", "-d", "1h", "-x", "1", "-x", "2", "-x", "3"},
			want: fftest.Vars{Str: "bar", Duration: time.Hour, Slice: []string{"1", "2", "3"}},
		},
		{
			name: "priority repeats",
			env:  map[string]string{"TEST_PARSE_STR": "s.env", "TEST_PARSE_STRING_SLICE": "x.env.1"},
			file: "testdata/5.conf",
			args: []string{"-s", "s.arg.1", "-s", "s.arg.2", "-x", "x.arg.1", "-x", "x.arg.2"},
			opts: []ff.Option{ff.WithEnvVarPrefix("TEST_PARSE")},
			want: fftest.Vars{Str: "s.arg.2", Slice: []string{"x.arg.1", "x.arg.2"}}, // highest prio wins and no others are called
		},
		{
			name: "PlainParser solo bool",
			file: "testdata/solo_bool.conf",
			want: fftest.Vars{Str: "x", Bool: true},
		},
		{
			name: "PlainParser string with spaces",
			file: "testdata/spaces.conf",
			want: fftest.Vars{Str: "i am the very model of a modern major general"},
		},
		{
			name: "default comma behavior",
			env:  map[string]string{"TEST_PARSE_STR": "one,two,three", "TEST_PARSE_STRING_SLICE": "one,two,three"},
			opts: []ff.Option{ff.WithEnvVarPrefix("TEST_PARSE")},
			want: fftest.Vars{Str: "one,two,three", Slice: []string{"one","two","three"}},
		},
		{
			name: "WithEnvVarSplit",
			env:  map[string]string{"TEST_PARSE_STR": "one,two,three", "TEST_PARSE_STRING_SLICE": "one,two,three"},
			opts: []ff.Option{ff.WithEnvVarPrefix("TEST_PARSE"), ff.WithEnvVarSplit(",")},
			want: fftest.Vars{Str: "three", Slice: []string{"one", "two", "three"}},
		},
		{
			name: "WithEnvVarNoPrefix",
			env:  map[string]string{"TEST_PARSE_STR": "foo", "STR": "bar"},
			opts: []ff.Option{ff.WithEnvVarNoPrefix()},
			want: fftest.Vars{Str: "bar"},
		},
		{
			name: "WithIgnoreUndefined env",
			env:  map[string]string{"TEST_PARSE_UNDEFINED": "one", "TEST_PARSE_STR": "one"},
			opts: []ff.Option{ff.WithEnvVarPrefix("TEST_PARSE"), ff.WithIgnoreUndefined(true)},
			want: fftest.Vars{Str: "one"},
		},
		{
			name: "WithIgnoreUndefined file true",
			file: "testdata/undefined.conf",
			opts: []ff.Option{ff.WithIgnoreUndefined(true)},
			want: fftest.Vars{Str: "one"},
		},
		{
			name: "WithIgnoreUndefined file false",
			file: "testdata/undefined.conf",
			opts: []ff.Option{ff.WithIgnoreUndefined(false)},
			want: fftest.Vars{WantParseErrorString: "config file flag"},
		},
		{
			name: "env var split comma whitespace",
			env:  map[string]string{"TEST_PARSE_STR": "one, two, three ", "TEST_PARSE_STRING_SLICE": "one, two, three "},
			opts: []ff.Option{ff.WithEnvVarPrefix("TEST_PARSE"), ff.WithEnvVarSplit(",")},
			want: fftest.Vars{Str: " three ", Slice: []string{"one", " two", " three "}},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			if testcase.file != "" {
				testcase.opts = append(testcase.opts, ff.WithConfigFile(testcase.file), ff.WithConfigFileParser(ff.PlainParser))
			}

			if len(testcase.env) > 0 {
				for k, v := range testcase.env {
					defer os.Setenv(k, os.Getenv(k))
					err := os.Setenv(k, v)
					if err != nil {
						t.Error(err)
					}
				}
			}

			fs, vars := fftest.Pair()
			vars.ParseError = ff.Parse(fs, testcase.args, testcase.opts...)
			if err := fftest.Compare(&testcase.want, vars); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestParseIssue16(t *testing.T) {
	t.Parallel()

	for _, testcase := range []struct {
		name string
		data string
		want string
	}{
		{
			name: "hash in value",
			data: "s bar#baz",
			want: "bar#baz",
		},
		{
			name: "EOL comment with space",
			data: "s bar # baz",
			want: "bar",
		},
		{
			name: "EOL comment no space",
			data: "s bar #baz",
			want: "bar",
		},
		{
			name: "only comment with space",
			data: "# foo bar\n",
			want: "",
		},
		{
			name: "only comment no space",
			data: "#foo bar\n",
			want: "",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			filename, cleanup := fftest.TempFile(t, testcase.data)
			defer cleanup()

			fs, vars := fftest.Pair()
			vars.ParseError = ff.Parse(fs, []string{},
				ff.WithConfigFile(filename),
				ff.WithConfigFileParser(ff.PlainParser),
			)

			want := fftest.Vars{Str: testcase.want}
			if err := fftest.Compare(&want, vars); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestParseConfigFile(t *testing.T) {
	t.Parallel()

	for _, testcase := range []struct {
		name         string
		missing      bool
		allowMissing bool
		parseError   error
	}{
		{
			name: "has config file",
		},
		{
			name:       "config file missing",
			missing:    true,
			parseError: os.ErrNotExist,
		},
		{
			name:         "config file missing + allow missing",
			missing:      true,
			allowMissing: true,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			filename := "dummy"
			if !testcase.missing {
				var cleanup func()
				filename, cleanup = fftest.TempFile(t, "")
				defer cleanup()
			}

			options := []ff.Option{ff.WithConfigFile(filename), ff.WithConfigFileParser(ff.PlainParser)}
			if testcase.allowMissing {
				options = append(options, ff.WithAllowMissingConfigFile(true))
			}

			fs, vars := fftest.Pair()
			vars.ParseError = ff.Parse(fs, []string{}, options...)

			want := fftest.Vars{WantParseErrorIs: testcase.parseError}
			if err := fftest.Compare(&want, vars); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestParseConfigFileVia(t *testing.T) {
	t.Parallel()

	var (
		rootFS = pflag.NewFlagSet("root", pflag.ContinueOnError)
		config = rootFS.StringP("config-file", "c", "", "")
		i      = rootFS.IntP("int", "i",0, "")
		s      = rootFS.StringP("str", "s","", "")
		subFS  = pflag.NewFlagSet("subcommand", pflag.ContinueOnError)
		d      = subFS.DurationP("duration","d", time.Second, "")
		b      = subFS.BoolP("bool", "b", false, "")
	)

	subCommand := &ffcli.Command{
		Name:    "subcommand",
		FlagSet: subFS,
		Options: []ff.Option{
			ff.WithConfigFileParser(ff.PlainParser),
			ff.WithConfigFileVia(config),
			ff.WithIgnoreUndefined(true),
		},
		Exec: func(ctx context.Context, args []string) error { return nil },
	}

	root := &ffcli.Command{
		Name:    "root",
		FlagSet: rootFS,
		Options: []ff.Option{
			ff.WithConfigFileParser(ff.PlainParser),
			ff.WithConfigFileFlag("config-file"),
			ff.WithIgnoreUndefined(true),
		},
		Exec:        func(ctx context.Context, args []string) error { return nil },
		Subcommands: []*ffcli.Command{subCommand},
	}

	err := root.ParseAndRun(context.Background(), []string{"--config-file", "testdata/1.conf", "subcommand"})
	if err != nil {
		t.Fatal(err)
	}

	if want, have := time.Hour, *d; want != have {
		t.Errorf("d: want %v, have %v", want, have)
	}
	if want, have := true, *b; want != have {
		t.Errorf("b: want %v, have %v", want, have)
	}
	if want, have := "bar", *s; want != have {
		t.Errorf("s: want %q, have %q", want, have)
	}
	if want, have := 99, *i; want != have {
		t.Errorf("i: want %d, have %d", want, have)
	}

}
