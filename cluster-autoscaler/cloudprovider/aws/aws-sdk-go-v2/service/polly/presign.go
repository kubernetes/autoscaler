package polly

import (
	"bytes"
	"context"
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/aws/protocol/query"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/middleware"
	smithyhttp "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/transport/http"
)

// AddPresignSynthesizeSpeechMiddleware adds presignOpSynthesizeSpeechInput into middleware stack to
// parse SynthesizeSpeechInput into request stream
func AddPresignSynthesizeSpeechMiddleware(stack *middleware.Stack) error {
	return stack.Serialize.Insert(&presignOpSynthesizeSpeechInput{}, "Query:AsGetRequest", middleware.Before)
}

// presignOpSynthesizeSpeechInput encodes SynthesizeSpeechInput into url format
// query string and put that into request stream for later presign-url build
type presignOpSynthesizeSpeechInput struct {
}

func (*presignOpSynthesizeSpeechInput) ID() string {
	return "PresignSerializer"
}

func (m *presignOpSynthesizeSpeechInput) HandleSerialize(ctx context.Context, in middleware.SerializeInput, next middleware.SerializeHandler) (
	out middleware.SerializeOutput, metadata middleware.Metadata, err error,
) {
	request, ok := in.Request.(*smithyhttp.Request)
	if !ok {
		return out, metadata, &smithy.SerializationError{Err: fmt.Errorf("unknown transport type %T", in.Request)}
	}

	input, ok := in.Parameters.(*SynthesizeSpeechInput)
	_ = input
	if !ok {
		return out, metadata, &smithy.SerializationError{Err: fmt.Errorf("unknown input parameters type %T", in.Parameters)}
	}

	bodyWriter := bytes.NewBuffer(nil)
	bodyEncoder := query.NewEncoder(bodyWriter)

	if err := presignSerializeOpDocumentSynthesizeSpeechInput(input, bodyEncoder.Value); err != nil {
		return out, metadata, &smithy.SerializationError{Err: err}
	}

	err = bodyEncoder.Encode()
	if err != nil {
		return out, metadata, &smithy.SerializationError{Err: err}
	}

	if request, err = request.SetStream(bytes.NewReader(bodyWriter.Bytes())); err != nil {
		return out, metadata, &smithy.SerializationError{Err: err}
	}

	in.Request = request

	return next.HandleSerialize(ctx, in)
}

func presignSerializeOpDocumentSynthesizeSpeechInput(v *SynthesizeSpeechInput, value query.Value) error {
	object := value.Object()
	_ = object

	if v.LexiconNames != nil && len(v.LexiconNames) > 0 {
		objectKey := object.KeyWithValues("LexiconNames")
		for _, name := range v.LexiconNames {
			objectKey.String(name)
		}
	}

	if len(v.OutputFormat) > 0 {
		objectKey := object.Key("OutputFormat")
		objectKey.String(string(v.OutputFormat))
	}

	if v.SampleRate != nil {
		objectKey := object.Key("SampleRate")
		objectKey.String(*v.SampleRate)
	}

	if v.Text != nil {
		objectKey := object.Key("Text")
		objectKey.String(*v.Text)
	}

	if len(v.TextType) > 0 {
		objectKey := object.Key("TextType")
		objectKey.String(string(v.TextType))
	}

	if len(v.VoiceId) > 0 {
		objectKey := object.Key("VoiceId")
		objectKey.String(string(v.VoiceId))
	}

	return nil
}
