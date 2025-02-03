package action

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"

	fishcomp "github.com/gopasspw/gopass/internal/completion/fish"
	zshcomp "github.com/gopasspw/gopass/internal/completion/zsh"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/tree"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/urfave/cli/v2"
	"github.com/hashicorp/go-set"
	"golang.org/x/exp/slices"
)

// Complete prints a list of all password names to os.Stdout.
func (s *Action) Complete(c *cli.Context) {
	ctx := ctxutil.WithGlobalFlags(c)
	_, err := s.Store.IsInitialized(ctx) // important to make sure the structs are not nil.
	if err != nil {
		out.Errorf(ctx, "Store not initialized: %s", err)

		return
	}
	list, err := s.Store.List(ctx, tree.INF)
	if err != nil {
		return
	}

	outs := set.New[string](20)

	re := regexp.MustCompile(`\\(.)`)
	in := re.ReplaceAllString(c.Args().First(), `$1`)

	lst := strings.Split(in, "/")

	for _, v := range list {
		if strings.HasPrefix(v, in) {
			v_split := strings.Split(v, "/")

			out := strings.Join(v_split[:len(lst)], "/")

			if len(v_split) > len(lst) {
				out += "/"
			}

			outs.Insert(out)
		}
	}

	var outs_lst = outs.Slice()
	slices.Sort(outs_lst)

	for _, v := range outs_lst {
		fmt.Fprintln(stdout, v)
	}
}

// CompletionOpenBSDKsh returns an OpenBSD ksh script used for auto completion.
func (s *Action) CompletionOpenBSDKsh(a *cli.App) error {
	out := `
PASS_LIST=$(gopass ls -f)
set -A complete_gopass -- $PASS_LIST %s
`

	if a == nil {
		return fmt.Errorf("can not parse command options")
	}

	opts := make([]string, 0, len(a.Commands))
	for _, opt := range a.Commands {
		opts = append(opts, opt.Name)
		if len(opt.Aliases) > 0 {
			opts = append(opts, strings.Join(opt.Aliases, " "))
		}
	}

	fmt.Fprintf(stdout, out, strings.Join(opts, " "))

	return nil
}

// CompletionBash returns a bash script used for auto completion.
func (s *Action) CompletionBash(c *cli.Context) error {
	out := `_gopass_bash_autocomplete() {
	local cur opts base
	COMPREPLY=()

	cur="${COMP_WORDS[COMP_CWORD]}"
	opts=$( ${COMP_WORDS[@]:0:$COMP_CWORD}  "${cur}"  --generate-bash-completion )
	local IFS=$'\n'

	COMPREPLY=( $( compgen -W "${opts}" -- "${cur}" ) )

	if [ ${#COMPREPLY[@]} -eq 1 ]; then
		if [ "${COMPREPLY[0]: -1}" != "/" ]; then
			compopt +o nospace
		fi
	fi

	return 0
}

`
	out += "complete -o filenames -F _gopass_bash_autocomplete -o nospace -o nosort " + s.Name
	if runtime.GOOS == "windows" {
		out += "\ncomplete -o filenames -F _gopass_bash_autocomplete -o nospace -o nosort " + s.Name + ".exe"
	}
	fmt.Fprintln(stdout, out)

	return nil
}

// CompletionFish returns an autocompletion script for fish.
func (s *Action) CompletionFish(a *cli.App) error {
	if a == nil {
		return fmt.Errorf("app is nil")
	}
	comp, err := fishcomp.GetCompletion(a)
	if err != nil {
		return err
	}

	fmt.Fprintln(stdout, comp)

	return nil
}

// CompletionZSH returns a zsh completion script.
func (s *Action) CompletionZSH(a *cli.App) error {
	comp, err := zshcomp.GetCompletion(a)
	if err != nil {
		return err
	}

	fmt.Fprintln(stdout, comp)

	return nil
}
