package qformat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type S struct {
	A int
}

func (s *S) Y() string {
	return "YY"
}

type SS struct {
	S1 S
	S2 S
}

func (ss *SS) X() int {
	return 17
}

type ABC struct {
	A string
	B string
	C string
}

func TestOne(t *testing.T) {
	s := S{42}
	s2 := SS{s, s}

	q := Q{}
	q["one"] = "Alpha"
	q["s"] = s
	q["s2"] = &s2
	assert.Equal(t,
		"Hello Alpha <<Missing: 'two'>> <<Missing: 'three'>> <<Positional arg 4 not present>> hats 42 42 {42} 17",
		q.Sprintf("Hello {one} {two} {three} {4} {0} {s.A} {s2.S1.A} {s2.S1} {s2.X}", "hats"),
	)

	assert.Equal(t, "pointer method on non-pointer test: YY", q.Sprintf("pointer method on non-pointer test: {s2.S1.Y}"))

	q["."] = &ABC{"up", "down", "strange"}

	assert.Equal(t, "Dot Test: up, down, strange", q.Sprintf("Dot Test: {A}, {B}, {C}"))

}

type Level3 struct{}
type Level2 struct{ l *Level3 }
type Level1 struct{ l *Level2 }

func (l *Level3) Name() string { return "ok" }
func (l *Level2) L3() *Level3  { return l.l }
func (l *Level1) L2() *Level2  { return l.l }

func TestNested(t *testing.T) {
	ThreeLayer := Level1{l: &Level2{l: &Level3{}}}
	q := Q{}
	q["."] = &ThreeLayer

	assert.Equal(t, "ok", q.Sprintf("{L2.L3.Name}"))
}

type Namer interface {
	Name() string
}

type NameHolder struct{ n Namer }

func (n *NameHolder) Namer() Namer { return n.n }

func TestInterface(t *testing.T) {
	InterfaceTest := &NameHolder{&Level3{}}
	q := Q{}
	q["."] = InterfaceTest

	assert.Equal(t, "ok", q.Sprintf("{Namer.Name}"))
}
