package status

// Status constants. Can be used to compare to api status fields
const (
	Available   = "available"
	Creating    = "creating"
	Deleted     = "deleted"
	Deleting    = "deleting"
	Deprecated  = "deprecated"
	Failing     = "failing"
	Failed      = "failed"
	Pending     = "pending"
	Unavailable = "unavailable"
)

// Account additional status constants. Compare with Account status fields.
const (
	Closed     = "closed"
	Overdue    = "overdue"
	Suspended  = "suspended"
	Terminated = "terminated"
	Warning    = "warning"
)

// Cloud IP additional status constants. Compare with Cloud IP status fields.
const (
	Mapped   = "mapped"
	Reserved = "reserved"
	Unmapped = "unmapped"
)

// Collaboration additional status constants. Compare with Collaboration status fields.
const (
	Accepted  = "accepted"
	Cancelled = "cancelled"
	Ended     = "ended"
	Rejected  = "rejected"
)

// Server Type additional status constants. Compare with Server Type status fields.
const (
	Experimental = "experimental"
)

// Server additional status constants. Compare with Server status fields.
const (
	Active   = "active"
	Inactive = "inactive"
)
