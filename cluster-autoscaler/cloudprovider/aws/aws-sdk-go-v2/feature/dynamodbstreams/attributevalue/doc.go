// Package attributevalue provides marshaling and unmarshaling utilities to
// convert between Go types and Amazon DynamoDB Streams AttributeValues.
//
// These utilities allow you to marshal slices, maps, structs, and scalar
// values to and from the AttributeValue type. These utilities make it
// easier to convert between AttributeValue and Go types when working with
// DynamoDB resources.
//
// This package only converts between Go types and DynamoDBStreams
// AttributeValue. See the feature/dynamodb/attributevalue package for
// converting to DynamoDB AttributeValue types.
//
// # Converting AttributeValue between DynamoDB and DynamoDBStreams
//
// The FromDynamoDBMap, FromDynamoDBList, and FromDynamoDB functions provide
// the conversion utilities to convert a DynamoDB AttributeValue type to a
// DynamoDBStreams AttributeValue type. Use these utilities when you need to
// convert the AttributeValue type between the two APIs.
//
// # AttributeValue Unmarshaling
//
// To unmarshal an AttributeValue to a Go type you can use the Unmarshal,
// UnmarshalList, UnmarshalMap, and UnmarshalListOfMaps functions. The List and
// Map functions are specialized versions of the Unmarshal function for
// unmarshal slices and maps of Attributevalues.
//
// The following example will unmarshal Items result from the DynamoDBStreams
// GetRecords operation. The items returned will be unmarshaled into the slice
// of the Records struct.
//
//	type Record struct {
//	    ID     string
//	    URLs   []string
//	}
//
//	//...
//
//	result, err := client.GetRecords(context.Context(), &dynamodbstreams.GetRecordsInput{
//	    ShardIterator: &shardIterator,
//	})
//	if err != nil {
//	    return fmt.Errorf("failed to get records from stream, %w", err)
//	}
//
//	var records []Record
//	for _, ddbRecord := range result.Records {
//	    if record.DynamoDB == nil {
//	        continue
//	    }
//
//	    var record
//	    err := attributevalue.UnmarshalMap(ddbRecord.NewImage, &record)
//	    if err != nil {
//	         return fmt.Errorf("failed to unmarshal record, %w", err))
//	    }
//	    records = append(records, record)
//	}
//
// # Struct tags
//
// The AttributeValue Marshal and Unmarshal functions support the `dynamodbav`
// struct tag by default. Additional tags can be enabled with the
// EncoderOptions and DecoderOptions, TagKey option.
//
// See the Marshal and Unmarshal function for information on how struct tags
// and fields are marshaled and unmarshaled.
package attributevalue
