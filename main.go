package main

import (
	"fmt"
	"strings"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

// Release - A release that could consist of multiple commits and has a version (semver).
type Release struct {
	MajorVersion   int
	MinorVersion   int
	PatchVersion   int
	ReleaseDate    string
	CommitMessages []string
}

func main() {
	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL: "https://github.com/emil-nasso/release-demo-repo.git",
	})
	checkIfError(err)
	// ... retrieves the branch pointed by HEAD
	ref, err := r.Head()
	checkIfError(err)
	// ... retrieves the commit object
	commit, err := r.CommitObject(ref.Hash())
	checkIfError(err)
	// ... retrieves the commit history
	history, err := commit.History()
	checkIfError(err)

	// Walk through all commits, in from the first one to the last one (reverse order)
	// and add a release when [Release] is found in a message
	releases := make([]Release, 0)
	unreleased := newRelease()
	for i := len(history) - 1; i >= 0; i-- {
		message := history[i].Message
		unreleased.CommitMessages = append(unreleased.CommitMessages, message)

		if strings.Contains(message, "[Release]") {
			unreleased.ReleaseDate = history[i].Author.When.Format(object.DateFormat)
			releases = append(releases, unreleased)
			unreleased = newRelease()
		}
	}
	// Calculate semver version for all releases based on occurence of [Breaking],
	// [Feature] and [Bug] in the commit messages
	setVersions(&releases)
	// Display it all.
	printReleases(unreleased, releases)
}

func checkIfError(err error) {
	if err != nil {
		panic(err)
	}
}

func newRelease() Release {
	return Release{
		CommitMessages: []string{},
	}
}

func printReleases(unreleased Release, releases []Release) {
	fmt.Println("# Change log")

	fmt.Println("## [Unreleased]")
	fmt.Println("### [Changed]")
	fmt.Printf("- %s\n\n", formatCommitMessages(unreleased.CommitMessages))

	for i := len(releases) - 1; i >= 0; i-- {
		release := releases[i]
		fmt.Printf(
			"## [%d.%d.%d] - (%s):\n",
			release.MajorVersion, release.MinorVersion, release.PatchVersion,
			release.ReleaseDate,
		)
		fmt.Println("### Changed")
		fmt.Printf("- %s\n\n", formatCommitMessages(releases[i].CommitMessages))
	}
}

func formatCommitMessages(messages []string) string {
	formatted := strings.Join(messages, "")
	formatted = strings.Trim(formatted, "\n")
	return strings.Replace(formatted, "\n", "\n- ", -1)
}

func setVersions(releases *[]Release) {
	version := [3]int{1, 0, 0}
	for i, release := range *releases {
		if i > 0 {
			messages := strings.Join(release.CommitMessages, "")
			if strings.Contains(messages, "[Breaking]") {
				version = [3]int{version[0] + 1, 0, 0}
			} else if strings.Contains(messages, "[Feature]") {
				version = [3]int{version[0], version[1] + 1, 0}
			} else if strings.Contains(messages, "[Bug]") {
				version = [3]int{version[0], version[1], version[2] + 1}
			}
		}
		(*releases)[i].MajorVersion = version[0]
		(*releases)[i].MinorVersion = version[1]
		(*releases)[i].PatchVersion = version[2]
	}
}
