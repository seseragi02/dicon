package internal

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"

	"golang.org/x/tools/imports"
)

type Generator struct {
	buf         bytes.Buffer
	PackageName string
}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) Printf(format string, args ...interface{}) {
	fmt.Fprintf(&g.buf, format, args...)
}

func (g *Generator) Generate(it *InterfaceType, fs []FuncType) error {
	g.PackageName = it.PackageName
	g.appendHeader(it)
	g.appendStructDefs(it)
	g.appendMethod(fs, "")
	return nil
}

func (g *Generator) GenerateMock(it *InterfaceType, targets []InterfaceType) error {
	if g.PackageName == "" {
		g.PackageName = it.PackageName
	}
	g.appendHeader(it)
	for _, i := range targets {
		g.appendMockStruct(&i)
	}
	return nil
}

func (g *Generator) Out(w io.Writer, filename string) error {
	dist, err := imports.Process(filename, g.buf.Bytes(), &imports.Options{Comments: true})
	if err != nil {
		fmt.Printf("%s\n", g.buf.Bytes())
		return err
	}
	if _, err := io.Copy(w, bytes.NewReader(dist)); err != nil {
		return err
	}
	return nil
}

func (g *Generator) appendHeader(it *InterfaceType) {
	g.Printf("// Code generated by \"dicon\"; DO NOT EDIT.\n")
	g.Printf("\n")
	g.Printf("package %s\n", g.PackageName)
	g.Printf("\n")
	g.Printf("import (\n")
	g.Printf("\"log\"\n")
	g.Printf("\"github.com/pkg/errors\"\n")
	g.Printf(")\n")
}

func (g *Generator) appendStructDefs(it *InterfaceType) {
	g.Printf("type dicontainer struct {\n")
	g.Printf("store map[string]interface{}\n")
	g.Printf("}\n")
	g.Printf("func NewDIContainer() %s {\n", it.Name)
	g.Printf("return &dicontainer{\n")
	g.Printf("store: map[string]interface{}{},\n")
	g.Printf("}\n")
	g.Printf("}\n")
	g.Printf("\n")
}

func (g *Generator) appendMethod(funcs []FuncType, _ string) {
	for _, f := range funcs {
		g.Printf("func (d *dicontainer) %s()", f.Name)
		if len(f.ReturnTypes) != 2 {
			log.Fatalf("Must be (instance, error) signature but %v", f.ReturnTypes)
		}

		returnType := f.ReturnTypes[0]
		g.Printf("(%s, error) {\n", returnType.ConvertName(g.PackageName))

		g.Printf("if i, ok := d.store[\"%s\"]; ok {\n", f.Name)
		g.Printf("instance, ok := i.(%s)\n", returnType.ConvertName(g.PackageName))
		g.Printf("if ok {\n")
		g.Printf("return instance, nil\n")
		g.Printf("}\n")
		g.Printf("return nil, fmt.Errorf(\"invalid instance is cached %%v\", instance)\n")
		g.Printf("}\n")

		dep := make([]string, 0, len(f.ArgumentTypes))
		for i, a := range f.ArgumentTypes {
			g.Printf("dep%d, err := d.%s()\n", i, a.SimpleName())
			g.Printf("if err != nil {\n")
			g.Printf("return nil, errors.Wrap(err, \"resolve %s failed at DICON\")\n", a.SimpleName())
			g.Printf("}\n")
			dep = append(dep, fmt.Sprintf("dep%d", i))
		}

		g.Printf("instance, err := %sNew%s(%s)\n", g.relativePackageName(f.PackageName), f.Name, strings.Join(dep, ", "))
		g.Printf("if err != nil {\n")
		g.Printf("return nil, errors.Wrap(err, \"creation %s failed at DICON\")\n", f.Name)
		g.Printf("}\n")
		g.Printf("d.store[\"%s\"] = instance\n", f.Name)
		g.Printf("return instance, nil\n")
		g.Printf("}\n")
	}
}

func (g *Generator) appendMockStruct(it *InterfaceType) {
	g.Printf("type %sMock struct {\n", it.Name)
	args := map[string][]string{}
	returns := map[string][]string{}

	for _, f := range it.Funcs {
		var ags []string
		for i, a := range f.ArgumentTypes {
			ags = append(ags, fmt.Sprintf("a%d %s", i, a.ConvertName(g.PackageName)))
		}
		args[f.Name] = ags

		var rets []string
		for _, r := range f.ReturnTypes {
			rets = append(rets, r.ConvertName(g.PackageName))
		}
		returns[f.Name] = rets
		g.Printf("%sMock func(%s)", f.Name, strings.Join(ags, ","))
		if len(rets) == 1 {
			g.Printf("%s", strings.Join(rets, ","))
		} else if len(rets) != 0 {
			g.Printf("(%s)", strings.Join(rets, ","))
		}
		g.Printf("\n")
	}

	g.Printf("}\n")
	g.Printf("\n")
	g.Printf("func New%sMock() *%sMock {\n", it.Name, it.Name)
	g.Printf("return &%sMock{}\n", it.Name)
	g.Printf("}\n")
	g.Printf("\n")

	for _, f := range it.Funcs {
		ags := args[f.Name]
		rets := returns[f.Name]

		g.Printf("func (mk *%sMock) %s(%s) ", it.Name, f.Name, strings.Join(ags, ","))
		if len(rets) == 1 {
			g.Printf("%s", rets[0])
		} else if len(rets) != 0 {
			g.Printf("(%s)", strings.Join(rets, ","))
		}
		g.Printf(" {\n")
		var a []string
		for i, _ := range ags {
			a = append(a, fmt.Sprintf("a%d", i))
		}
		g.Printf("return mk.%sMock(%s)\n", f.Name, strings.Join(a, ","))
		g.Printf("}\n")
	}
}

func (g *Generator) relativePackageName(packageName string) string {
	if g.PackageName == packageName {
		return ""
	}
	return packageName + "."
}
