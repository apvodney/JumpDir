package fatalError

type Error struct{ error }

func Panic(err error) {
	panic(Error{error: err})
}
