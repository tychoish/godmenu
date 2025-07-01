package godmenu

type Optional[T int] struct {
	v       T
	defined bool
}

func NewOptional[T int](in T) Optional[T]  { return Optional[T]{v: in, defined: true} }
func (o Optional[T]) Set(in T) Optional[T] { o.defined = true; o.v = in; return o }
func (o Optional[T]) Reset() Optional[T]   { return Optional[T]{} }
func (o Optional[T]) Resolve() T           { return o.v }
func (o Optional[T]) OK() bool             { return o.defined }
