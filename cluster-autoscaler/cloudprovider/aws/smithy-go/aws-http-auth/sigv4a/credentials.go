package sigv4a

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
	"math"
	"math/big"
	"sync"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/aws-http-auth/credentials"
)

var (
	p256          elliptic.Curve
	nMinusTwoP256 *big.Int

	one = new(big.Int).SetInt64(1)
)

func init() {
	p256 = elliptic.P256()

	nMinusTwoP256 = new(big.Int).SetBytes(p256.Params().N.Bytes())
	nMinusTwoP256 = nMinusTwoP256.Sub(nMinusTwoP256, new(big.Int).SetInt64(2))
}

// ecdsaCache stores the result of deriving an ECDSA private key from a
// shared-secret identity.
type ecdsaCache struct {
	mu sync.Mutex

	akid string
	priv *ecdsa.PrivateKey
}

// Derive computes and caches the ECDSA key-pair for the identity, returning
// the result.
//
// Future calls to Derive with the same set of credentials (identified by AKID)
// will short-circuit. Future calls with a different set of credentials
// (identified by AKID) will re-derive the value, overwriting the old result.
func (c *ecdsaCache) Derive(creds credentials.Credentials) (*ecdsa.PrivateKey, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if creds.AccessKeyID == c.akid {
		return c.priv, nil
	}

	priv, err := derivePrivateKey(creds)
	if err != nil {
		return nil, err
	}

	c.akid = creds.AccessKeyID
	c.priv = priv
	return priv, nil
}

// derivePrivateKey derives a NIST P-256 PrivateKey from the given IAM
// AccessKey and SecretKey pair.
//
// Based on FIPS.186-4 Appendix B.4.2
func derivePrivateKey(creds credentials.Credentials) (*ecdsa.PrivateKey, error) {
	akid := creds.AccessKeyID
	secret := creds.SecretAccessKey

	params := p256.Params()
	bitLen := params.BitSize // Testing random candidates does not require an additional 64 bits
	counter := 0x01

	buffer := make([]byte, 1+len(akid)) // 1 byte counter + len(accessKey)
	kdfContext := bytes.NewBuffer(buffer)

	inputKey := append([]byte("AWS4A"), []byte(secret)...)

	d := new(big.Int)
	for {
		kdfContext.Reset()
		kdfContext.WriteString(akid)
		kdfContext.WriteByte(byte(counter))

		key, err := deriveHMACKey(sha256.New, bitLen, inputKey, []byte(algorithm), kdfContext.Bytes())
		if err != nil {
			return nil, err
		}

		cmp, err := cmpConst(key, nMinusTwoP256.Bytes())
		if err != nil {
			return nil, err
		}
		if cmp == -1 {
			d.SetBytes(key)
			break
		}

		counter++
		if counter > 0xFF {
			return nil, fmt.Errorf("exhausted single byte external counter")
		}
	}
	d = d.Add(d, one)

	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = p256
	priv.D = d
	priv.PublicKey.X, priv.PublicKey.Y = p256.ScalarBaseMult(d.Bytes())

	return priv, nil
}

// deriveHMACKey provides an implementation of a NIST-800-108 of a KDF (Key
// Derivation Function) in Counter Mode. HMAC is used as the pseudorandom
// function, where the value of `r` is defined as a 4-byte counter.
func deriveHMACKey(hash func() hash.Hash, bitLen int, key []byte, label, context []byte) ([]byte, error) {
	// verify that we won't overflow the counter
	n := int64(math.Ceil((float64(bitLen) / 8) / float64(hash().Size())))
	if n > 0x7FFFFFFF {
		return nil, fmt.Errorf("unable to derive key of size %d using 32-bit counter", bitLen)
	}

	// verify the requested bit length is not larger then the length encoding size
	if int64(bitLen) > 0x7FFFFFFF {
		return nil, fmt.Errorf("bitLen is greater than 32-bits")
	}

	fixedInput := bytes.NewBuffer(nil)
	fixedInput.Write(label)
	fixedInput.WriteByte(0x00)
	fixedInput.Write(context)
	if err := binary.Write(fixedInput, binary.BigEndian, int32(bitLen)); err != nil {
		return nil, fmt.Errorf("failed to write bit length to fixed input string: %v", err)
	}

	var output []byte

	h := hmac.New(hash, key)

	for i := int64(1); i <= n; i++ {
		h.Reset()
		if err := binary.Write(h, binary.BigEndian, int32(i)); err != nil {
			return nil, err
		}
		_, err := h.Write(fixedInput.Bytes())
		if err != nil {
			return nil, err
		}
		output = append(output, h.Sum(nil)...)
	}

	return output[:bitLen/8], nil
}

// constant-time byte slice compare
func cmpConst(x, y []byte) (int, error) {
	if len(x) != len(y) {
		return 0, fmt.Errorf("slice lengths do not match")
	}

	xLarger, yLarger := 0, 0

	for i := 0; i < len(x); i++ {
		xByte, yByte := int(x[i]), int(y[i])

		x := ((yByte - xByte) >> 8) & 1
		y := ((xByte - yByte) >> 8) & 1

		xLarger |= x &^ yLarger
		yLarger |= y &^ xLarger
	}

	return xLarger - yLarger, nil
}
