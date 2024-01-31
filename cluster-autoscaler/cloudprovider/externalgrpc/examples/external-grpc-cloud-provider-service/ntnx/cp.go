package ntnx

type ntnxCP struct {
}

func (cp *ntnxCP) Name() string {
	return "ntnx"
}

func NodeGroups() []NodeGroup {
	// /v2/nodepool/list for cmsp

	// convert msp nodepools to NodeGroup
}
