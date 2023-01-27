package main

import (
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
	"os"
	"strings"
	"text/template"
)

type Event struct {
	Name string
}

type Events struct {
    Package string
	Events []Event
}

const loadMode = packages.NeedName |
	packages.NeedFiles |
	packages.NeedCompiledGoFiles |
	packages.NeedImports |
	packages.NeedDeps |
	packages.NeedTypes |
	packages.NeedSyntax |
	packages.NeedTypesInfo

func main() {
    var path string
    var pkgName string

	switch os.Args[1] {
	case "player":
        path = "github.com/df-mc/dragonfly/server/player"
        pkgName = "player"
	case "world": path = ""
        path = "github.com/df-mc/dragonfly/server/world"
        pkgName = "world"
	case "server": path = ""
	}

    events := Events{
        Package: pkgName,
    }

	loadConfig := new(packages.Config)
	loadConfig.Mode = loadMode
	loadConfig.Fset = token.NewFileSet()

	pkgs, _ := packages.Load(
		loadConfig,
		path,
	)

	pkg := pkgs[0]

	for _, syn := range pkg.Syntax {
		for _, dec := range syn.Decls {
			if gen, ok := dec.(*ast.GenDecl); ok && gen.Tok == token.TYPE {
				for _, spec := range gen.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						obj, ok := pkg.TypesInfo.Defs[ts.Name]

						if !ok {
							continue
						}

						typeName, ok := obj.(*types.TypeName)

						if !ok {
							continue
						}

						t := typeName.Type().String()
						if strings.Contains(t, "Event") {
							s := strings.Split(
								t,
								".",
							)[2]

                            s = strings.ReplaceAll(s, "Event", "")

							events.Events = append(
								events.Events,
								Event{s},
							)
						}
					}
				}
			}
		}
	}

	tmplFile, _ := os.ReadFile("tmpl.txt")
	tmpl, _ := template.New("handlermanager").Parse(string(tmplFile))

    file, _ := os.Create("handlermanager.go")
	tmpl.Execute(file, events)
}
