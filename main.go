// +build ignore
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"html/template"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var aggregateTpl string = `
package main

type {{ .TypeName }} struct {

	// Fields to track internal state go
	// here, don't store vanity attributes
	// on the aggregate itself, only fields
	// needed to satisfy business logic go
	// in this part of the applicaton.

	// These are just examples of fields
	// you might want to care about
	isEnabled  bool
	validUntil *time.Time

}

func ({{ .Abbr }} *{{ .TypeName }}) ReactTo(ev Event) error {
	switch ev {
	default:
		return errors.Errorf("{{ .TypeName}} aggregate doesn't know what to do with %s", ev)
	}
	return nil
}
`

func main() {

	var rootCmd = &cobra.Command{
		Use:   "ls-cms",
		Short: "LS-Cms is a bad name for a CQRS/DDD app generator",
		Long: `A Fast and Flexible Static Site Generator built with
                love by spf13 and friends in Go.
                Complete documentation is available at http://hugo.spf13.com`,
	}

	var cmdList = &cobra.Command{
		Use:   "list",
		Short: "Sub-command lists various types of things registered in the app",
		Long:  `enumerates things in the working directory and prints them, see subcommands for info`,
	}

	var cmdListAggregate = &cobra.Command{
		Use:   "aggregate",
		Short: "generate an aggregate with a specicif name",
		Long:  `long description of aggregate gen command`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {

			// get pwd
			pwd, err := os.Getwd()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			// parse everything in sight
			fileSet := token.NewFileSet()
			astPkgs, err := parser.ParseDir(fileSet, pwd, func(info os.FileInfo) bool {
				name := info.Name()
				return !info.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go")
			}, parser.ParseComments)
			if err != nil {
				log.Fatal(err)
			}

			// Inspect the AST because we are pros.
			for _, pkgAst := range astPkgs {
				ast.Inspect(pkgAst, func(n ast.Node) bool {
					var s string
					switch x := n.(type) {
					// case *ast.BasicLit:
					// 	s = x.Value
					case *ast.TypeSpec:
						s = x.Name.Name
					}
					if s != "" {
						fmt.Printf("%s:\t%s\n", fileSet.Position(n.Pos()), s)
					}
					return true
				})
			}

		},
	}

	cmdList.AddCommand(cmdListAggregate)

	var cmdGen = &cobra.Command{
		Use:   "gen",
		Short: "........",
		Long:  `fksfkdslfjsljflsjlfskl`,
	}

	var cmdGenAggregate = &cobra.Command{
		Use:   "aggregate [name]",
		Short: "generate an aggregate with a specicif name",
		Long:  `long description of aggregate gen command`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {

			ident := NewIdentifier(args[0])

			var t *template.Template = template.Must(template.New("aggregate").Parse(aggregateTpl))

			type Aggregate struct {
				TypeName string
				Abbr     string
			}

			agg := Aggregate{
				TypeName: ident.TypeName(),
				Abbr:     ident.Abbr(),
			}

			err := t.Execute(os.Stdout, agg)
			if err != nil {
				log.Println("executing template:", err)
			}

		},
	}

	cmdGen.AddCommand(cmdGenAggregate)

	rootCmd.AddCommand(cmdList, cmdGen)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
