package pkgload

import (
	"fmt"
	"path"
	"strings"
)

func ParseMethodString(cwd, fullMethod string) (pkg, name string, err error) {
	parts := strings.SplitN(fullMethod, ":", 2)
	switch len(parts) {
	case 0:
		return pkg, name, fmt.Errorf("invalid custom method: %s", fullMethod)
	case 1:
		name = parts[0]
		pkg = cwd
	case 2:
		pkg = parts[0]
		name = parts[1]
		if strings.HasPrefix(pkg, "./") || strings.HasPrefix(pkg, "../") {
			pkg = path.Join(cwd, pkg)
		}

		if pkg == "" {
			// example: goverter:extend :MyLocalConvert
			// the purpose of the ':' in this case is confusing, do not allow such case
			return pkg, name, fmt.Errorf(`package path must not be empty in the custom method %q.
See https://goverter.jmattheis.de/reference/extend`, fullMethod)
		}
	}

	if name == "" {
		return pkg, name, fmt.Errorf(`method name pattern is required in the custom method %q.
See https://goverter.jmattheis.de/reference/extend`, fullMethod)
	}
	if strings.Contains(pkg, "...") {
		return pkg, name, fmt.Errorf(`package wildcard pattern ... is not supported: %q`, fullMethod)
	}
	return pkg, name, nil
}
