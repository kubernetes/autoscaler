//go:build codegen
// +build codegen

package api

type legacyJSONValues struct {
	Type          string
	StructMembers map[string]struct{}
	ListMemberRef bool
	MapValueRef   bool
}

var legacyJSONValueShapes = map[string]map[string]legacyJSONValues{
	"braket": {
		"CreateQuantumTaskRequest": {
			Type: "structure",
			StructMembers: map[string]struct{}{
				"action":           {},
				"deviceParameters": {},
			},
		},
		"GetDeviceResponse": {
			Type: "structure",
			StructMembers: map[string]struct{}{
				"deviceCapabilities": {},
			},
		},
		"GetQuantumTaskResponse": {
			Type: "structure",
			StructMembers: map[string]struct{}{
				"deviceParameters": {},
			},
		},
	},
	"cloudwatchrum": {
		"RumEvent": {
			Type: "structure",
			StructMembers: map[string]struct{}{
				"details":  {},
				"metadata": {},
			},
		},
	},
	"lexruntimeservice": {
		"PostContentRequest": {
			Type: "structure",
			StructMembers: map[string]struct{}{
				"requestAttributes": {},
				//"ActiveContexts":    struct{}{}, - Disabled because JSON List
				"sessionAttributes": {},
			},
		},
		"PostContentResponse": {
			Type: "structure",
			StructMembers: map[string]struct{}{
				// "alternativeIntents":  struct{}{}, - Disabled because JSON List
				"sessionAttributes":   {},
				"nluIntentConfidence": {},
				"slots":               {},
				//"activeContexts":      struct{}{}, - Disabled because JSON List
			},
		},
		"PutSessionResponse": {
			Type: "structure",
			StructMembers: map[string]struct{}{
				// "activeContexts":    struct{}{}, - Disabled because JSON List
				"slots":             {},
				"sessionAttributes": {},
			},
		},
	},
	"lookoutequipment": {
		"DatasetSchema": {
			Type: "structure",
			StructMembers: map[string]struct{}{
				"InlineDataSchema": {},
			},
		},
		"DescribeDatasetResponse": {
			Type: "structure",
			StructMembers: map[string]struct{}{
				"Schema": {},
			},
		},
		"DescribeModelResponse": {
			Type: "structure",
			StructMembers: map[string]struct{}{
				"Schema":       {},
				"ModelMetrics": {},
			},
		},
	},
	"networkmanager": {
		"CoreNetworkPolicy": {
			Type: "structure",
			StructMembers: map[string]struct{}{
				"PolicyDocument": {},
			},
		},
		"GetResourcePolicyResponse": {
			Type: "structure",
			StructMembers: map[string]struct{}{
				"PolicyDocument": {},
			},
		},
		"PutCoreNetworkPolicyRequest": {
			Type: "structure",
			StructMembers: map[string]struct{}{
				"PolicyDocument": {},
			},
		},
		"PutResourcePolicyRequest": {
			Type: "structure",
			StructMembers: map[string]struct{}{
				"PolicyDocument": {},
			},
		},
	},
	"personalizeevents": {
		"Event": {
			Type: "structure",
			StructMembers: map[string]struct{}{
				"properties": {},
			},
		},
		"Item": {
			Type: "structure",
			StructMembers: map[string]struct{}{
				"properties": {},
			},
		},
		"User": {
			Type: "structure",
			StructMembers: map[string]struct{}{
				"properties": {},
			},
		},
	},
	"pricing": {
		"PriceList": {
			Type:          "list",
			ListMemberRef: true,
		},
	},
	"rekognition": {
		"HumanLoopActivationOutput": {
			Type: "structure",
			StructMembers: map[string]struct{}{
				"HumanLoopActivationConditionsEvaluationResults": {},
			},
		},
	},
	"sagemaker": {
		"HumanLoopActivationConditionsConfig": {
			Type: "structure",
			StructMembers: map[string]struct{}{
				"HumanLoopActivationConditions": {},
			},
		},
	},
	"schemas": {
		"GetResourcePolicyResponse": {
			Type: "structure",
			StructMembers: map[string]struct{}{
				"Policy": {},
			},
		},
		"PutResourcePolicyRequest": {
			Type: "structure",
			StructMembers: map[string]struct{}{
				"Policy": {},
			},
		},
		"PutResourcePolicyResponse": {
			Type: "structure",
			StructMembers: map[string]struct{}{
				"Policy": {},
			},
		},
		"GetResourcePolicyOutput": {
			Type: "structure",
			StructMembers: map[string]struct{}{
				"Policy": {},
			},
		},
		"PutResourcePolicyInput": {
			Type: "structure",
			StructMembers: map[string]struct{}{
				"Policy": {},
			},
		},
		"PutResourcePolicyOutput": {
			Type: "structure",
			StructMembers: map[string]struct{}{
				"Policy": {},
			},
		},
	},
	"textract": {
		"HumanLoopActivationOutput": {
			Type: "structure",
			StructMembers: map[string]struct{}{
				"HumanLoopActivationConditionsEvaluationResults": {},
			},
		},
	},
}
