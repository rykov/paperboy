package mail

import (
	msgd "github.com/emersion/go-msgauth/dkim"
	"github.com/rykov/paperboy/config"
	"github.com/spf13/afero"
	"github.com/spf13/cast"
	"github.com/wneessen/go-mail-middleware/dkim"

	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"time"
)

func DKIMMiddleware(ac *config.AConfig) (*dkim.Middleware, error) {
	conf, appFs := ac.DKIM, ac.AppFs

	// Required: Read private key from keyFile
	keyFile := cast.ToString(conf["keyfile"])
	if keyFile == "" {
		return nil, fmt.Errorf("DKIM requires a keyFile")
	}

	keyBytes, err := afero.ReadFile(appFs, keyFile)
	if err != nil {
		return nil, err
	}

	// Domain and selector for DKIM
	domain := cast.ToString(conf["domain"])
	sel := cast.ToString(conf["selector"])

	// Header fields to sign
	opts := []dkim.SignerOption{
		dkim.WithHeaderFields(
			"Mime-Version", "To", "From", "Subject", "Reply-To",
			"Sender", "Content-Transfer-Encoding", "Content-Type",
		),
	}

	sc, err := dkim.NewConfig(domain, sel, opts...)
	if err != nil {
		return nil, err
	}

	// (Optional) Expiration configuration
	if v, ok := conf["signatureexpirein"]; ok {
		sc.SetExpiration(time.Unix(0, cast.ToInt64(v)))
	}

	// (Optional) Canonicalization configuration
	if v, ok := conf["canonicalization"]; ok && v != "" {
		parts := strings.SplitN(cast.ToString(v), "/", 2)
		if len(parts) != 2 {
			return nil, dkim.ErrInvalidCanonicalization
		}
		errH := sc.SetHeaderCanonicalization(msgd.Canonicalization(parts[0]))
		errB := sc.SetBodyCanonicalization(msgd.Canonicalization(parts[1]))
		if err := errors.Join(errH, errB); err != nil {
			return nil, err
		}
	}

	// Create the middleware with key & config
	return newDkimMiddleware(keyBytes, sc)
}

// Create the middleware with key & config with auto-detection
func newDkimMiddleware(keyBytes []byte, sc *dkim.SignerConfig) (*dkim.Middleware, error) {
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Try PKCS8 first (supports both RSA and Ed25519)
	if privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		switch key := privateKey.(type) {
		case ed25519.PrivateKey:
			return dkim.NewFromEd25519Key(keyBytes, sc)
		case *rsa.PrivateKey:
			return dkim.NewFromRSAKey(pem.EncodeToMemory(&pem.Block{
				Bytes: x509.MarshalPKCS1PrivateKey(key),
				Type:  "RSA PRIVATE KEY",
			}), sc)
		default:
			return nil, fmt.Errorf("unsupported key type: %T", privateKey)
		}
	}

	// Try PKCS1 (RSA only)
	if _, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return dkim.NewFromRSAKey(keyBytes, sc)
	}

	return nil, fmt.Errorf("unsupported key format or type")
}
