package enum

// Transformer transforms a source enum members to target enum members
//
// The transformer must only return keys present inside the
// context.Source.Members and context.Target.Members, if something cannot be
// mapped by the transformer just skip the key and don't return it. An error by
// this methods aborts the aborts the whole goverter conversion, so only use it
// when there are config errors.
type Transformer func(context TransformContext) (map[string]string, error)

type TransformContext struct {
	Source Enum
	Target Enum
	// Config is user definable config
	Config string
}
