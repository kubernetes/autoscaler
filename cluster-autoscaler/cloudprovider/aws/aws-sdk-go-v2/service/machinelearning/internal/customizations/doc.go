/*
Package customizations provides customizations for the Machine Learning API client.

The Machine Learning API client uses one customization to support the PredictEndpoint
input parameter.

# Predict Endpoint

The predict endpoint customization runs after normal endpoint resolution happens. If
the user has provided a value for PredictEndpoint then this customization will
overwrite the request's endpoint with that value.
*/
package customizations
