package attributevalue

import (
	"fmt"

	ddb "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/service/dynamodb/types"
	streams "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/service/dynamodbstreams/types"
)

// FromDynamoDBMap converts a map of Amazon DynamoDB
// AttributeValues, and all nested members.
func FromDynamoDBMap(from map[string]ddb.AttributeValue) (to map[string]streams.AttributeValue, err error) {
	to = make(map[string]streams.AttributeValue, len(from))
	for field, value := range from {
		to[field], err = FromDynamoDB(value)
		if err != nil {
			return nil, err
		}
	}

	return to, nil
}

// FromDynamoDBList converts a slice of Amazon DynamoDB
// AttributeValues, and all nested members.
func FromDynamoDBList(from []ddb.AttributeValue) (to []streams.AttributeValue, err error) {
	to = make([]streams.AttributeValue, len(from))
	for i := 0; i < len(from); i++ {
		to[i], err = FromDynamoDB(from[i])
		if err != nil {
			return nil, err
		}
	}

	return to, nil
}

// FromDynamoDB converts an Amazon DynamoDB  AttributeValue, and
// all nested members.
func FromDynamoDB(from ddb.AttributeValue) (streams.AttributeValue, error) {
	switch tv := from.(type) {
	case *ddb.AttributeValueMemberNULL:
		return &streams.AttributeValueMemberNULL{Value: tv.Value}, nil

	case *ddb.AttributeValueMemberBOOL:
		return &streams.AttributeValueMemberBOOL{Value: tv.Value}, nil

	case *ddb.AttributeValueMemberB:
		return &streams.AttributeValueMemberB{Value: tv.Value}, nil

	case *ddb.AttributeValueMemberBS:
		bs := make([][]byte, len(tv.Value))
		for i := 0; i < len(tv.Value); i++ {
			bs[i] = append([]byte{}, tv.Value[i]...)
		}
		return &streams.AttributeValueMemberBS{Value: bs}, nil

	case *ddb.AttributeValueMemberN:
		return &streams.AttributeValueMemberN{Value: tv.Value}, nil

	case *ddb.AttributeValueMemberNS:
		return &streams.AttributeValueMemberNS{Value: append([]string{}, tv.Value...)}, nil

	case *ddb.AttributeValueMemberS:
		return &streams.AttributeValueMemberS{Value: tv.Value}, nil

	case *ddb.AttributeValueMemberSS:
		return &streams.AttributeValueMemberSS{Value: append([]string{}, tv.Value...)}, nil

	case *ddb.AttributeValueMemberL:
		values, err := FromDynamoDBList(tv.Value)
		if err != nil {
			return nil, err
		}
		return &streams.AttributeValueMemberL{Value: values}, nil

	case *ddb.AttributeValueMemberM:
		values, err := FromDynamoDBMap(tv.Value)
		if err != nil {
			return nil, err
		}
		return &streams.AttributeValueMemberM{Value: values}, nil

	default:
		return nil, fmt.Errorf("unknown AttributeValue union member, %T", from)
	}
}
