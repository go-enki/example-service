package models

import "fmt"

type Greeting struct {
	Template string
	Name     string
	Rendered string
}

func (g *Greeting) Render() string {
	return fmt.Sprintf(g.Template, g.Name)
}

func (g *Greeting) Validate() error {
	if g.Name == "" {
		return ErrGreetingEmptyName
	}

	return nil
}