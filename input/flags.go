package input

import (
	"flag"
	"os"
	"strings"
)

type Options struct {
	Repo           string
	BranchModified string
	BranchOrigin   string
	LocaleModified string
	LocaleOrigin   string
	Folder         string
}

var flags struct {
	Repo     string
	Locales  string
	Origin   string
	Modified string
}

func init() {
	flag.StringVar(&flags.Repo, "r", ".", "path to project folder - must be a git repository")
	flag.StringVar(&flags.Locales, "p", "locales", "path to locales folder from repository root")
	flag.StringVar(&flags.Modified, "o", "", "branch & locale name to compare against. Defaults to develop:[compared locale]")
	flag.StringVar(&flags.Origin, "m", "", "branch & locale name to compare, e.g. 'develop:en'. Defaults to [current branch]:ru")
}

func GetFlags() (Options, bool) {
	flag.Parse()

	if flag.NFlag() == 0 {
		return Options{}, false
	}

	var (
		repo     = &flags.Repo
		folder   = &flags.Locales
		origin   = &flags.Origin
		modified = flags.Modified
	)

	branchAndLocaleBase := strings.Split(*origin, ":")
	branchBase := branchAndLocaleBase[0]
	localeBase := ""
	if len(branchAndLocaleBase) > 1 {
		localeBase = branchAndLocaleBase[1]
	}
	branchAndLocaleModified := strings.Split(modified, ":")
	branchModified := branchAndLocaleModified[0]
	localeModified := "ru"
	if len(branchAndLocaleModified) > 1 {
		localeModified = branchAndLocaleModified[1]
	}
	if localeBase == "" {
		localeBase = localeModified
	}
	if branchBase == "" {
		branchBase = "develop"
	}
	if *repo == "." {
		wd, _ := os.Getwd()
		repo = &wd
	}

	return Options{
		Repo:           *repo,
		Folder:         *folder,
		LocaleOrigin:   localeBase,
		LocaleModified: localeModified,
		BranchOrigin:   branchBase,
		BranchModified: branchModified,
	}, true
}

func Help() {
	flag.Usage()
}
