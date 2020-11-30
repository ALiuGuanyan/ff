package fftest

import (
	"errors"
	"fmt"
	"github.com/ALiuGuanyan/pflag"
	"strings"
	"time"
)

// Pair returns a predefined flag set, and a predefined set of variables that
// have been registered into it. Tests can call parse on the flag set with a
// variety of flags, config files, and env vars, and check the resulting effect
// on the vars.
func Pair() (*pflag.FlagSet, *Vars) {
	fs := pflag.NewFlagSet("fftest", pflag.ContinueOnError)

	var v Vars
	fs.StringVarP(&v.Str, "str", "s","", "string")
	fs.IntVarP(&v.Int, "int","i", 0, "int")
	fs.Float64VarP(&v.Float, "float","f", 0., "float64")
	fs.BoolVarP(&v.Bool, "bool", "b",false, "bool")
	fs.DurationVarP(&v.Duration, "duration","d", 0*time.Second, "time.Duration")
	fs.StringSliceVarP(&v.Slice, "string-slice", "x", []string{},"collection of strings (repeatable)")

	return fs, &v
}

// Vars are a common set of variables used for testing.
type Vars struct {
	Str string
	Int int
	Float float64
	Bool bool
	Duration time.Duration
	Slice []string

	// ParseError should be assigned as the result of Parse in tests.
	ParseError error

	// If a test case expects an input to generate a parse error,
	// it can specify that error here. The Compare helper will
	// look for it using errors.Is.
	WantParseErrorIs error

	// If a test case expects an input to generate a parse error,
	// it can specify part of that error string here. The Compare
	// helper will look for it using strings.Contains.
	WantParseErrorString string
}

// Compare one set of vars with another
// and return an error on any difference.
func Compare(want, have *Vars) error {
	if want.WantParseErrorIs != nil || want.WantParseErrorString != "" {
		if want.WantParseErrorIs != nil && have.ParseError == nil {
			return fmt.Errorf("want error (%v), have none", want.WantParseErrorIs)
		}

		if want.WantParseErrorString != "" && have.ParseError == nil {
			return fmt.Errorf("want error (%q), have none", want.WantParseErrorString)
		}

		if want.WantParseErrorIs == nil && want.WantParseErrorString == "" && have.ParseError != nil {
			return fmt.Errorf("want clean parse, have error (%v)", have.ParseError)
		}

		if want.WantParseErrorIs != nil && have.ParseError != nil && !errors.Is(have.ParseError, want.WantParseErrorIs) {
			return fmt.Errorf("want wrapped error (%#+v), have error (%#+v)", want.WantParseErrorIs, have.ParseError)
		}

		if want.WantParseErrorString != "" && have.ParseError != nil && !strings.Contains(have.ParseError.Error(), want.WantParseErrorString) {
			return fmt.Errorf("want error string (%q), have error (%v)", want.WantParseErrorString, have.ParseError)
		}

		return nil
	}

	if want.Str != have.Str {
		return fmt.Errorf("var S: want %q, have %q", want.Str, have.Str)
	}
	if want.Int != have.Int {
		return fmt.Errorf("var I: want %d, have %d", want.Int, have.Int)
	}
	if want.Float != have.Float {
		return fmt.Errorf("var F: want %f, have %f", want.Float, have.Float)
	}
	if want.Bool != have.Bool {
		return fmt.Errorf("var Bool: want %v, have %v", want.Bool, have.Bool)
	}
	if want.Duration != have.Duration {
		return fmt.Errorf("var Duration: want %s, have %s", want.Duration, have.Duration)
	}

	for i := 0; i < len(want.Slice); i++ {
		if want.Slice[i] != have.Slice[i] {
			return fmt.Errorf("var X: want %v, have %v", want.Slice, have.Slice)
		}
	}


	return nil
}

