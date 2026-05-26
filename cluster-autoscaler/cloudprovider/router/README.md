# Router Package

Package router provides a centralized way to include cloud provider implementations into the Cluster Autoscaler binary using Go build tags.

This package is primarily a convenience for the main Cluster Autoscaler build, allowing it to easily include all supported providers by default or a specific subset via tags (e.g., `-tags aws`).

## Note for forks and specialized builds

While this package is useful for the main distribution, external forks or specialized deployments that only require a specific set of providers are encouraged to bypass this package and instead use blank imports of the specific cloud provider packages directly in their main entry point. This avoids maintaining the tag logic and ensures a cleaner dependency graph.
