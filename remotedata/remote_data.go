package remotedata

type State[T any] interface {
	sealed()
}

type NotAsked struct{}

func (NotAsked) sealed() {}

type Loading struct{}

func (Loading) sealed() {}

type Failure struct {
	Error error
}

func (Failure) sealed() {}

type Success[T any] struct {
	Data T
}

func (Success[T]) sealed() {}

// CustomFieldTypeSwitch is a mechanism to switch over all possible
// values of a CustomFieldType, allowing the compiler to check all
// routes have been handled.
func Match[T any, R any](fieldType State[T],
	notAsked func(NotAsked) (R, error),
	loading func(Loading) (R, error),
	failure func(Failure) (R, error),
	success func(Success[T]) (R, error),
) (res R, err error) {
	switch fieldType.(type) {
	case NotAsked:
		return notAsked(fieldType.(NotAsked))
	case Loading:
		return loading(fieldType.(Loading))

	case Failure:
		return failure(fieldType.(Failure))
	case Success[T]:
		return success(fieldType.(Success[T]))
	}

	return res, err
}
