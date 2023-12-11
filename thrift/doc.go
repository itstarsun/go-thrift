package thrift

// requireKeyedLiterals can be embedded in a struct to require keyed literals.
type requireKeyedLiterals struct{}

// nonComparable can be embedded in a struct to prevent comparability.
type nonComparable [0]func()

// A Set represents a slice of T that will be encoded as a Thrift set.
type Set[T any] []T

func (Set[T]) set() {}

// A List represents a slice of T that will be encoded as a Thrift list.
type List[T any] []T

func (List[T]) list() {}
