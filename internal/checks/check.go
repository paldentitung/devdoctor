package checks

type Check interface {
	Run() Result
}

type Group struct {
	Name   string
	Checks []Check
}
