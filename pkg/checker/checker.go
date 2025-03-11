package checker

type Checker interface {
	Check(path string) (bool, error)
}
