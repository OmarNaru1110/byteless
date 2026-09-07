package command

type Command interface {
	Execute(args any) (any, error)
}
