package deepparser

import "testing"

func TestFindDefinitionFromAst_enum(t *testing.T) {
	var typer Typer
	typer.AddFromDir("IntEnum", "examples")
	if len(typer.Ordered) == 0 {
		t.Fatal("not index")
	}
	def := typer.Ordered[0]
	if def == nil {
		t.Fatal("nil enum")
	}
	t.Log(def)

	if !def.IsTypeAlias() {
		t.Fatal("is not type alias")
	}

	values := def.FindEnumValues()
	if len(values) != 4 {
		t.Fatal("should be 4 values but got", len(values))
	}
	for _, val := range values {
		t.Log(val.Name, "=", val.Value)
	}
}
