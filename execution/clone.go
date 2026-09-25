package execution

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cainlara/gogit-branch/core"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

// cloneArgs is the validated argument set for the clone subcommand
// (data-model.md §1). Name and Email are empty when not supplied as
// arguments — empty means "not provided" here; the prompted flow's
// empty-answer semantics live in CloneRepository itself (FR-005).
type cloneArgs struct {
	URL   string
	Name  string
	Email string
	Anon  bool
}

// parseCloneArgs validates the tokens following `clone`/`cl` BEFORE any side
// effect, so every rejected shape guarantees "nothing is cloned" (spec FR-007,
// research.md D6). Rules, in order:
//
//  1. URL required and always the first token — `-anon` in first position is
//     not a URL (contract §1: "gogit clone -anon" errors).
//  2. `-anon` is matched as an exact token anywhere after the URL; combined
//     with any name/email token it is contradictory and rejected (A1→B).
//  3. Name/email arrive as a pair — exactly one is rejected (FR-007).
//  4. Any surplus token is rejected (guard allows at most 4 extras: URL + 2 +
//     -anon, so this catches `clone u n e extra`).
//
// Pure: no git, no TTY — fully unit-testable (data-model.md §1 table).
func parseCloneArgs(args []string) (cloneArgs, error) {
	if len(args) == 0 || args[0] == "-anon" || args[0] == "" {
		return cloneArgs{}, errors.New("repository URL required (usage: clone <repo-url> [name email] [-anon])")
	}

	parsed := cloneArgs{URL: args[0]}
	var positionals []string

	for _, token := range args[1:] {
		if token == "-anon" {
			parsed.Anon = true
			continue
		}

		positionals = append(positionals, token)
	}

	if parsed.Anon && len(positionals) > 0 {
		return cloneArgs{}, errors.New("conflicting arguments: -anon cannot be combined with a user name and email")
	}

	if parsed.Anon {
		return parsed, nil
	}

	switch len(positionals) {
	case 0:
		return parsed, nil
	case 2:
		parsed.Name = positionals[0]
		parsed.Email = positionals[1]

		if parsed.Name == "" || parsed.Email == "" {
			return cloneArgs{}, errors.New("user name and email must both be non-empty (or omit both, or use -anon)")
		}

		return parsed, nil
	case 1:
		return cloneArgs{}, errors.New("user name and email must be provided together (or neither, or use -anon)")
	default:
		return cloneArgs{}, fmt.Errorf("unexpected arguments after the repository URL: %v", positionals[2:])
	}
}

// CloneRepository is the single entry point for the `clone`/`cl` subcommand
// (Constitution II: one subcommand = one execution file, one exported entry
// point). It validates arguments first (nothing runs on a bad shape), then
// clones with announce + live progress into the launch directory, then decides
// identity: -anon skips, an argument pair applies directly, otherwise prompts
// run after the successful clone (FR-005). Failures are returned for main.go
// to print (Constitution IV); skip/success messages are informational output
// on success paths (research.md D8).
func CloneRepository(gitClient *core.GitClient, args []string) error {
	parsed, err := parseCloneArgs(args)
	if err != nil {
		return err
	}

	fmt.Println()
	color.Cyan("Cloning repository")
	color.Cyan(fmt.Sprintf("Cloning %s...", parsed.URL))

	bar := newStatusStreamRenderer()
	defer bar.finish()

	stopInterruptWatch := clearBarOnInterrupt(bar)
	defer stopInterruptWatch()

	bar.start(progressStageCloning)

	if _, err := gitClient.CloneWithProgress(parsed.URL, bar.onProgress); err != nil {
		bar.finish()

		return err
	}

	bar.finish()

	return applyIdentity(gitClient, parsed)
}

// applyIdentity is the post-clone identity decision (contract §2 step 5):
// -anon skips everything; an argument pair applies directly; otherwise the
// prompted flow runs (FR-005). Every skip path prints ONE informational line
// and returns nil — the command succeeds (FR-008, FR-016); only a config
// failure returns an error, and the clone is kept either way (FR-010).
func applyIdentity(gitClient *core.GitClient, parsed cloneArgs) error {
	switch {
	case parsed.Anon:
		reportIdentitySkipped("requested with -anon")

		return nil
	case parsed.Name != "" || parsed.Email != "":
		if err := gitClient.SetLocalIdentity(core.TargetDirFromURL(parsed.URL), parsed.Name, parsed.Email); err != nil {
			return err
		}

		color.Green(fmt.Sprintf("%s Cloned repository — identity set for this repository", EMOJI_ROCKET))

		return nil
	default:
		return promptForIdentity(gitClient, parsed.URL)
	}
}

// reportIdentitySkipped prints the single skip line shared by every
// skip outcome (-anon, empty answers, cancelled/non-interactive prompts).
func reportIdentitySkipped(reason string) {
	color.Green(fmt.Sprintf("%s Cloned repository — identity setup skipped (%s)", EMOJI_HERB, reason))
}

// promptForIdentity is the prompted identity flow for URL-only runs
// (spec FR-005, clarification A4): ask name, then email, AFTER the successful
// clone. Neither prompt validates — an empty answer is accepted and skips its
// field; both empty means no identity setup at all (the -anon effect). Any
// prompt.Run() error — Ctrl+C, EOF, or a non-interactive session with no
// terminal to prompt on — takes the single skip path: clone kept, one
// informational line, return nil, command succeeds (FR-008, FR-016,
// research.md D4). A cancelled second prompt discards the first answer too:
// identity is all-or-nothing once prompting stops.
func promptForIdentity(gitClient *core.GitClient, url string) error {
	name, nameOK := runIdentityPrompt("User Name")
	if !nameOK {
		reportIdentitySkipped("prompt cancelled or unavailable")

		return nil
	}

	email, emailOK := runIdentityPrompt("User Email")
	if !emailOK {
		reportIdentitySkipped("prompt cancelled or unavailable")

		return nil
	}

	if name == "" && email == "" {
		reportIdentitySkipped("no values entered")

		return nil
	}

	if err := gitClient.SetLocalIdentity(core.TargetDirFromURL(url), name, email); err != nil {
		return err
	}

	color.Green(fmt.Sprintf("%s Cloned repository — identity set for this repository", EMOJI_ROCKET))

	return nil
}

// runIdentityPrompt renders one freeform prompt with NO Validate (empty
// answers must pass, FR-005) and normalizes incidental surrounding whitespace.
// ok=false covers both cancellation and a prompt that cannot be shown at all
// — the caller treats them identically (research.md D4).
func runIdentityPrompt(label string) (string, bool) {
	prompt := promptui.Prompt{Label: label}

	result, err := prompt.Run()
	if err != nil {
		return "", false
	}

	return strings.TrimSpace(result), true
}
