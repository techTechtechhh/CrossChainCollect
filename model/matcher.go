package model

type Matcher interface {
	Match(Results) (Results, Results, Results, error)
}
