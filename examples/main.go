package main

import (
	"html/template"
	"os"
)

func main() {
	example()
}
func htmlT(html string) *template.Template {
	return template.Must(template.New("").Parse(html))
}

func example() {
	// html template example
	type Inventory struct {
		Material string
		Count    uint
	}
	sweaters := Inventory{"wool", 17}

	tmpl :=
		htmlT(`
			<div class="h-1">
				<ul>
				<li> a list</li>
				</ul>
				{{.Count}}
			 </div>`)
	err := tmpl.Execute(os.Stdout, sweaters)
	if err != nil {
		panic(err)
	}
}
