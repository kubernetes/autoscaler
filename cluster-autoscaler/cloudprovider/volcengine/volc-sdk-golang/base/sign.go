/*
Copyright 2023 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package base

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

func (c Credentials) Sign(request *http.Request) *http.Request {
	query := request.URL.Query()
	request.URL.RawQuery = query.Encode()

	if request.URL.Path == "" {
		request.URL.Path += "/"
	}
	requestParam := RequestParam{
		IsSignUrl: false,
		Body:      readAndReplaceBody(request),
		Host:      request.Host,
		Path:      request.URL.Path,
		Method:    request.Method,
		Date:      now(),
		QueryList: query,
		Headers:   request.Header,
	}
	signRequest := GetSignRequest(requestParam, c)

	request.Header.Set("Host", signRequest.Host)
	request.Header.Set("Content-Type", signRequest.ContentType)
	request.Header.Set("X-Date", signRequest.XDate)
	request.Header.Set("X-Content-Sha256", signRequest.XContentSha256)
	request.Header.Set("Authorization", signRequest.Authorization)
	if signRequest.XSecurityToken != "" {
		request.Header.Set("X-Security-Token", signRequest.XSecurityToken)
	}
	return request
}

func (c Credentials) SignUrl(request *http.Request) string {
	query := request.URL.Query()

	requestParam := RequestParam{
		IsSignUrl: true,
		Body:      readAndReplaceBody(request),
		Host:      request.Host,
		Path:      request.URL.Path,
		Method:    request.Method,
		Date:      now(),
		QueryList: query,
		Headers:   request.Header,
	}
	signRequest := GetSignRequest(requestParam, c)

	query.Set("X-Date", signRequest.XDate)
	query.Set("X-NotSignBody", signRequest.XNotSignBody)
	query.Set("X-Credential", signRequest.XCredential)
	query.Set("X-Algorithm", signRequest.XAlgorithm)
	query.Set("X-SignedHeaders", signRequest.XSignedHeaders)
	query.Set("X-SignedQueries", signRequest.XSignedQueries)
	query.Set("X-Signature", signRequest.XSignature)
	if signRequest.XSecurityToken != "" {
		query.Set("X-Security-Token", signRequest.XSecurityToken)
	}
	return query.Encode()
}

func GetSignRequest(requestParam RequestParam, credentials Credentials) SignRequest {
	formatDate := appointTimestampV4(requestParam.Date)
	meta := getMetaData(credentials, tsDateV4(formatDate))

	requestSignMap := make(map[string][]string)
	if credentials.SessionToken != "" {
		requestSignMap["X-Security-Token"] = []string{credentials.SessionToken}
	}
	signRequest := SignRequest{
		XDate:          formatDate,
		XSecurityToken: credentials.SessionToken,
	}
	var bodyHash string
	if requestParam.IsSignUrl {
		for k, v := range requestParam.QueryList {
			requestSignMap[k] = v
		}
		requestSignMap["X-Date"], requestSignMap["X-NotSignBody"], requestSignMap["X-Credential"], requestSignMap["X-Algorithm"], requestSignMap["X-SignedHeaders"], requestSignMap["X-SignedQueries"] =
			[]string{formatDate}, []string{""}, []string{credentials.AccessKeyID + "/" + meta.credentialScope}, []string{meta.algorithm}, []string{meta.signedHeaders}, []string{""}

		keys := make([]string, 0, len(requestSignMap))
		for k := range requestSignMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		requestSignMap["X-SignedQueries"] = []string{strings.Join(keys, ";")}

		signRequest.XNotSignBody, signRequest.XCredential, signRequest.XAlgorithm, signRequest.XSignedHeaders, signRequest.XSignedQueries =
			"", credentials.AccessKeyID+"/"+meta.credentialScope, meta.algorithm, meta.signedHeaders, strings.Join(keys, ";")
		bodyHash = hashSHA256([]byte{})
	} else {
		for k, v := range requestParam.Headers {
			requestSignMap[k] = v
		}
		if requestSignMap["Content-Type"] == nil {
			signRequest.ContentType = "application/x-www-form-urlencoded; charset=utf-8"
		} else {
			signRequest.ContentType = requestSignMap["Content-Type"][0]
		}
		requestSignMap["X-Date"], requestSignMap["Host"], requestSignMap["Content-Type"] = []string{formatDate}, []string{requestParam.Host}, []string{signRequest.ContentType}

		if len(requestParam.Body) == 0 {
			bodyHash = hashSHA256([]byte{})
		} else {
			bodyHash = hashSHA256(requestParam.Body)
		}
		requestSignMap["X-Content-Sha256"] = []string{bodyHash}
		signRequest.Host, signRequest.XContentSha256 = requestParam.Host, bodyHash
	}

	signature := getSignatureStr(requestParam, meta, credentials.SecretAccessKey, formatDate, requestSignMap, bodyHash)
	if requestParam.IsSignUrl {
		signRequest.XSignature = signature
	} else {
		signRequest.Authorization = buildAuthHeaderV4(signature, meta, credentials)
	}
	return signRequest
}

func getSignatureStr(requestParam RequestParam, meta *metadata, secretAccessKey string,
	formatDate string, requestSignMap map[string][]string, bodyHash string) string {
	// Task 1
	hashedCanonReq := hashedCanonicalRequestV4(requestParam, meta, requestSignMap, bodyHash)

	// Task 2
	stringToSign := concat("\n", meta.algorithm, formatDate, meta.credentialScope, hashedCanonReq)

	// Task 3
	signingKey := signingKeyV4(secretAccessKey, meta.date, meta.region, meta.service)
	return signatureV4(signingKey, stringToSign)
}

func hashedCanonicalRequestV4(param RequestParam, meta *metadata, requestSignMap map[string][]string, bodyHash string) string {
	var canonicalRequest string
	if param.IsSignUrl {
		queryList := make(url.Values)
		for k, v := range requestSignMap {
			for i := range v {
				queryList.Set(k, v[i])
			}
		}
		canonicalRequest = concat("\n", param.Method, normuri(param.Path), normquery(queryList), "\n", meta.signedHeaders, bodyHash)
	} else {
		canonicalHeaders := getCanonicalHeaders(param, meta, requestSignMap)
		canonicalRequest = concat("\n", param.Method, normuri(param.Path), normquery(param.QueryList), canonicalHeaders, meta.signedHeaders, bodyHash)
	}
	return hashSHA256([]byte(canonicalRequest))
}

func getCanonicalHeaders(param RequestParam, meta *metadata, requestSignMap map[string][]string) string {
	signMap := make(map[string][]string)
	signedHeaders := sortHeaders(requestSignMap, signMap)
	if !param.IsSignUrl {
		meta.signedHeaders = concat(";", signedHeaders...)
	}
	if param.Path == "" {
		param.Path = "/"
	}
	var headersToSign string
	for _, key := range signedHeaders {
		value := strings.TrimSpace(signMap[key][0])
		if key == "host" {
			if strings.Contains(value, ":") {
				split := strings.Split(value, ":")
				port := split[1]
				if port == "80" || port == "443" {
					value = split[0]
				}
			}
		}
		headersToSign += key + ":" + value + "\n"
	}
	return headersToSign
}

func sortHeaders(requestSignMap map[string][]string, signMap map[string][]string) []string {
	var sortedHeaderKeys []string
	for k, v := range requestSignMap {
		signMap[strings.ToLower(k)] = v
		switch k {
		case "Content-Type", "Content-Md5", "Host", "X-Security-Token":
		default:
			if !strings.HasPrefix(k, "X-") {
				continue
			}
		}
		sortedHeaderKeys = append(sortedHeaderKeys, strings.ToLower(k))
	}
	sort.Strings(sortedHeaderKeys)
	return sortedHeaderKeys
}

func getMetaData(credentials Credentials, date string) *metadata {
	meta := new(metadata)
	meta.date, meta.service, meta.region, meta.signedHeaders, meta.algorithm = date, credentials.Service, credentials.Region, "", "HMAC-SHA256"
	meta.credentialScope = concat("/", meta.date, meta.region, meta.service, "request")
	return meta
}

func signatureV4(signingKey []byte, stringToSign string) string {
	return hex.EncodeToString(hmacSHA256(signingKey, stringToSign))
}

func signingKeyV4(secretKey, date, region, service string) []byte {
	kDate := hmacSHA256([]byte(secretKey), date)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, service)
	kSigning := hmacSHA256(kService, "request")
	return kSigning
}

func buildAuthHeaderV4(signature string, meta *metadata, keys Credentials) string {
	credential := keys.AccessKeyID + "/" + meta.credentialScope

	return meta.algorithm +
		" Credential=" + credential +
		", SignedHeaders=" + meta.signedHeaders +
		", Signature=" + signature
}

func timestampV4() string {
	return now().Format(timeFormatV4)
}

func appointTimestampV4(date time.Time) string {
	return date.Format(timeFormatV4)
}
func tsDateV4(timestamp string) string {
	return timestamp[:8]
}

func hmacSHA256(key []byte, content string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(content))
	return mac.Sum(nil)
}

func hashSHA256(content []byte) string {
	h := sha256.New()
	h.Write(content)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func hashMD5(content []byte) string {
	h := md5.New()
	h.Write(content)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func readAndReplaceBody(request *http.Request) []byte {
	if request.Body == nil {
		return []byte{}
	}
	payload, _ := ioutil.ReadAll(request.Body)
	request.Body = ioutil.NopCloser(bytes.NewReader(payload))
	return payload
}

func concat(delim string, str ...string) string {
	return strings.Join(str, delim)
}

var now = func() time.Time {
	return time.Now().UTC()
}

func normuri(uri string) string {
	parts := strings.Split(uri, "/")
	for i := range parts {
		parts[i] = encodePathFrag(parts[i])
	}
	return strings.Join(parts, "/")
}

func encodePathFrag(s string) string {
	hexCount := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if shouldEscape(c) {
			hexCount++
		}
	}
	t := make([]byte, len(s)+2*hexCount)
	j := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if shouldEscape(c) {
			t[j] = '%'
			t[j+1] = "0123456789ABCDEF"[c>>4]
			t[j+2] = "0123456789ABCDEF"[c&15]
			j += 3
		} else {
			t[j] = c
			j++
		}
	}
	return string(t)
}

func shouldEscape(c byte) bool {
	if 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' {
		return false
	}
	if '0' <= c && c <= '9' {
		return false
	}
	if c == '-' || c == '_' || c == '.' || c == '~' {
		return false
	}
	return true
}

func normquery(v url.Values) string {
	queryString := v.Encode()

	return strings.Replace(queryString, "+", "%20", -1)
}
