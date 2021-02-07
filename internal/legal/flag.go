package legal

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

// RegisterFlag registers a legal flag to the provided flagset.
// When the legal flag is provided, it prints legal information and exits.
//
// When flagset is nil, uses flag.CommandLine.
func RegisterFlag(flagset *flag.FlagSet) {
	if flagset == nil {
		flagset = flag.CommandLine
	}

	flagset.Var(legalFlag{}, "legal", "Print legal notices and exit")
}

// legalFlag is a boolean flag that implements flag.Value.
// When it is set, it prints legal information and exits.
type legalFlag struct{}

func (legalFlag) String() string { return "" }

func (l legalFlag) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	if v {
		l.PrintAndExit()
	}
	return nil
}

func (legalFlag) IsBoolFlag() bool { return true }

func (legalFlag) PrintAndExit() {
	fmt.Println("This executable contains code from several different go packages. ")
	fmt.Println("Some of these packages require licensing information to be made available to the end user. ")
	fmt.Println(Notices)
	os.Exit(0)
}
