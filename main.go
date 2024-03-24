package main

import (
	"encoding/json"
	"fmt"
	"locales/gitutils"
	"locales/input"
	"locales/maps"

	"github.com/fatih/color"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
)

func main() {
	opts, isset := input.GetFlags()

	if !isset {
		input.Help()
		return
	}

	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL: "file:///" + opts.Repo,
	})
	if err != nil {
		panic(err)
	}

	var headBranch *plumbing.Reference
	if opts.BranchModified == "" {
		headBranch, _ = gitutils.FindCurrentBranch(*r)
	} else {
		headBranch, _ = gitutils.FindBranchByName(*r, opts.BranchModified)
	}
	headLocales, err := findLocales(r, *headBranch, opts.Folder, &opts.LocaleModified)
	if err != nil {
		panic(err)
	}
	var developBranch *plumbing.Reference
	if opts.BranchOrigin == "" {
		developBranch = headBranch
	} else {
		developBranch, _ = gitutils.FindBranchByName(*r, opts.BranchOrigin)
	}
	developLocales, err := findLocales(r, *developBranch, opts.Folder, &opts.LocaleOrigin)
	if err != nil {
		panic(err)
	}

	lang, locale := maps.Divide(headLocales)
	dLang, dLocale := maps.Divide(developLocales)
	if len(locale) == 0 || len(dLocale) == 0 {
		return
	}

	diff := maps.Diff(maps.Dot[string, any](locale[0]), maps.Dot[string, any](dLocale[0]))

	fmt.Printf("%s %s > %s %s:\n", headBranch.Name(), lang[0], developBranch.Name(), dLang[0])
	if !diff.IsEmpty() {
		encoded, _ := json.MarshalIndent(maps.Undot(diff.AddedFields), "", "  ")
		color.Set(color.FgHiGreen)
		fmt.Println(string(encoded))
		color.Unset()
	} else {
		color.Cyan("No new translations added\n")
	}
	//TODO: compare all by locale

}

func findLocales(repository *git.Repository, ref plumbing.Reference, dirname string, locale *string) (map[string]map[string]any, error) {
	fileSelector := fmt.Sprintf("%s/*.json", dirname)
	if locale != nil {
		fileSelector = fmt.Sprintf("%s/%s.json", dirname, *locale)
	}
	found, err := gitutils.GetFilesFromRef(repository, ref.Hash(), fileSelector)
	if err != nil {
		return nil, err
	}
	results := make(map[string]map[string]any, len(found))
	for filename, contents := range found {
		var m map[string]any
		json.Unmarshal(contents, &m)
		results[filename] = m
	}
	return results, nil
}
