package deepparser

import (
	"bytes"
	"github.com/fatih/structtag"
	"github.com/reddec/godetector"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"path/filepath"
)

// Deeply parsed types. Currently supports only structs
type Typer struct {
	Ordered       []*Definition          // Inspected and parsed definition in order of inspection
	Parsed        map[string]*Definition // Indexed definition where index is <path>@<type>
	BeforeInspect func(def *Definition)  // Invoke hook before inspection (ex: RemoveJsonIgnoredFields)
}

// Add recursively pre-parsed structure definition
func (tsg *Typer) Add(def *Definition) {
	uid := def.Import.Path + "@" + def.TypeName
	_, ok := tsg.Parsed[uid]
	if ok {
		return
	}
	if tsg.Parsed == nil {
		tsg.Parsed = make(map[string]*Definition)
	}
	tsg.Ordered = append(tsg.Ordered, def)
	if tsg.BeforeInspect != nil {
		tsg.BeforeInspect(def)
	}
	tsg.Parsed[uid] = def

	for _, f := range def.StructFields() {
		alias := DetectPackageInType(f.AST.Type)
		typeName := RebuildTypeNameWithoutPackage(f.AST.Type)
		def := FindDefinitionFromAst(typeName, alias, def.File, def.FileDir)

		if def != nil {
			tsg.Add(def)
		}
	}
}

// Parse and add recursively type from directory. Do nothing if not found
func (tsg *Typer) AddFromDir(typeName string, dir string) {
	def := FindDefinitionFromAst(typeName, "", nil, dir)
	if def == nil {
		return
	}
	tsg.Add(def)
}

// Parse and add recursively type from specific file
func (tsg *Typer) AddFromFile(typeName string, filename string) {
	tsg.AddFromDir(typeName, filepath.Dir(filename))
}

// Parse and add type using full import name using current working directory
func (tsg *Typer) AddFromImport(typeName string, importPath string) {
	location, err := godetector.FindPackageDefinitionDir(importPath, ".")
	if err != nil {
		return
	}
	tsg.AddFromDir(typeName, location)
}

type Definition struct {
	Import   godetector.Import
	Decl     *ast.GenDecl
	Type     *ast.TypeSpec
	TypeName string
	FS       *token.FileSet
	FileDir  string
	File     *ast.File
}

func FindDefinitionFromAst(typeName, alias string, file *ast.File, fileDir string) *Definition {
	var importDef godetector.Import
	if alias != "" {
		v, err := godetector.ResolveImport(alias, file, fileDir)
		if err != nil {
			log.Println("failed resolve import for", alias, "from dir", fileDir, ":", err)
			return nil
		}
		importDef = *v
	} else {
		v, err := godetector.InspectImportByDir(fileDir)
		if err != nil {
			log.Println("failed inspect", fileDir, ":", err)
			return nil
		}
		importDef = *v
	}

	var fs token.FileSet
	importFile, err := parser.ParseDir(&fs, importDef.Location, nil, parser.AllErrors)
	if err != nil {
		log.Println("failed parse", importDef.Location, ":", err)
		return nil
	}
	for _, packageDefintion := range importFile {
		for _, packageFile := range packageDefintion.Files {
			for _, decl := range packageFile.Decls {
				if v, ok := decl.(*ast.GenDecl); ok && v.Tok == token.TYPE {
					for _, spec := range v.Specs {
						if st, ok := spec.(*ast.TypeSpec); ok && st.Name.Name == typeName {
							return &Definition{
								Import:   importDef,
								Decl:     v,
								Type:     st,
								FS:       &fs,
								TypeName: typeName,
								FileDir:  importDef.Location,
								File:     packageFile,
							}
						}
					}
				}
			}
		}
	}
	return nil
}

type StField struct {
	Name      string
	Type      string
	Tag       string
	Comment   string
	AST       *ast.Field
	Omitempty bool
}

func (def *Definition) IsStruct() bool {
	_, ok := def.Type.Type.(*ast.StructType)
	return ok
}

func (def *Definition) StructFields() []*StField {
	st, ok := def.Type.Type.(*ast.StructType)
	if !ok {
		return nil
	}
	if st.Fields == nil || len(st.Fields.List) == 0 {
		return nil
	}
	var ans []*StField
	for _, field := range st.Fields.List {
		if len(field.Names) == 0 {
			continue
		}
		if !ast.IsExported(field.Names[0].Name) {
			continue
		}
		var comment string
		if field.Comment != nil {
			comment = field.Comment.Text()
		}
		f := &StField{
			Name:    field.Names[0].Name,
			Tag:     field.Names[0].Name,
			Type:    AstPrint(field.Type, def.FS),
			Comment: comment,
			AST:     field,
		}
		ans = append(ans, f)
		if field.Tag == nil {
			continue
		}
		s := field.Tag.Value
		s = s[1 : len(s)-1]
		val, err := structtag.Parse(s)
		if err != nil {
			log.Println("failed parse tags:", err)
			continue
		}

		if jsTag, err := val.Get("json"); err == nil && jsTag != nil {
			if jsTag.Name != "-" {
				f.Tag = jsTag.Name
			}
			f.Omitempty = jsTag.HasOption("omitempty")
		}
	}
	return ans
}

func (def *Definition) RemoveJSONIgnoredFields() {
	st, ok := def.Type.Type.(*ast.StructType)
	if !ok {
		return
	}
	if st.Fields == nil || len(st.Fields.List) == 0 {
		return
	}
	var filtered []*ast.Field
	for _, field := range st.Fields.List {
		filtered = append(filtered, field)
		if field.Tag == nil {
			continue
		}
		s := field.Tag.Value
		s = s[1 : len(s)-1]
		val, err := structtag.Parse(s)
		if err != nil {
			log.Println("failed parse tags:", err)
			continue
		}
		if !ast.IsExported(field.Names[0].Name) {
			filtered = filtered[:len(filtered)-1]
			continue
		}

		if jsTag, err := val.Get("json"); err == nil && jsTag != nil {
			if jsTag.Value() == "-" {
				filtered = filtered[:len(filtered)-1]
			}
		}
	}
	st.Fields.List = filtered
}

func AstPrint(t ast.Node, fs *token.FileSet) string {
	var buf bytes.Buffer
	printer.Fprint(&buf, fs, t)
	return buf.String()
}

func DetectPackageInType(t ast.Expr) string {
	if acc, ok := t.(*ast.SelectorExpr); ok {
		return acc.X.(*ast.Ident).Name
	} else if ptr, ok := t.(*ast.StarExpr); ok {
		return DetectPackageInType(ptr.X)
	} else if arr, ok := t.(*ast.ArrayType); ok {
		return DetectPackageInType(arr.Elt)
	}
	return ""
}

func RebuildOps(t ast.Expr) string {
	if ptr, ok := t.(*ast.StarExpr); ok {
		return "*" + RebuildOps(ptr.X)
	}
	if arr, ok := t.(*ast.ArrayType); ok {
		return "[]" + RebuildOps(arr.Elt)
	}
	return ""
}

func RebuildTypeNameWithoutPackage(t ast.Expr) string {
	if v, ok := t.(*ast.Ident); ok {
		return v.Name
	}
	if ptr, ok := t.(*ast.StarExpr); ok {
		return RebuildTypeNameWithoutPackage(ptr.X)
	}
	if acc, ok := t.(*ast.SelectorExpr); ok {
		return acc.Sel.Name
	}
	if arr, ok := t.(*ast.ArrayType); ok {
		return RebuildTypeNameWithoutPackage(arr.Elt)
	}
	return ""
}

func RemoveJsonIgnoredFields(def *Definition) {
	def.RemoveJSONIgnoredFields()
}
