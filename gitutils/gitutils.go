package gitutils

import (
	"fmt"
	"locales/selectors"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func FindBranchByName(r git.Repository, name string) (*plumbing.Reference, error) {
	t, _ := r.Remote("origin")
	refs, err := t.List(&git.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, ref := range refs {
		if ref.Name() == plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", name)) {
			return ref, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func FindCurrentBranch(r git.Repository) (*plumbing.Reference, error) {
	return r.Head()
}

func FindEntries(tree *object.Tree, filePath selectors.Selector) map[string]object.TreeEntry {
	chunks := filePath.Chunks()
	results := make(map[string]object.TreeEntry)
	for _, entry := range tree.Entries {
		if !selectors.Satisfies(entry.Name, chunks[0]) {
			continue
		}
		if len(chunks) == 1 {
			results[entry.Name] = entry
			continue
		}
		nested, _ := tree.Tree(entry.Name)
		nestedEntries := FindEntries(nested, selectors.From(chunks[1:]))
		for k, v := range nestedEntries {
			results[entry.Name+"/"+k] = v
		}
	}
	return results
}

func GetFilesFromRef(repository *git.Repository, ref plumbing.Hash, filename string) (map[string][]byte, error) {
	commit, err := repository.CommitObject(ref)
	if err != nil {
		return nil, err
	}

	tree, err := repository.TreeObject(commit.TreeHash)
	if err != nil {
		return nil, err
	}

	found := FindEntries(tree, selectors.Selector{Selector: filename})
	if len(found) == 0 {
		return nil, nil
	}

	results := make(map[string][]byte, len(found))
	for name, entry := range found {
		data, err := readFileContents(*repository, entry)
		if err != nil {
			continue
		}
		results[name] = data
	}

	return results, nil
}

func readFileContents(r git.Repository, entry object.TreeEntry) ([]byte, error) {
	blob, err := r.BlobObject(entry.Hash)
	if err != nil {
		return nil, err
	}

	reader, err := blob.Reader()
	if err != nil {
		return nil, err
	}

	data := make([]byte, blob.Size)

	n, err := reader.Read(data)
	defer reader.Close()
	if err != nil {
		return nil, err
	}

	if int64(n) != blob.Size {
		return nil, fmt.Errorf("wrong size")
	}

	return data, nil
}
