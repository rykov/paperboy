package mail

import (
	msgd "github.com/emersion/go-msgauth/dkim"
	"github.com/rykov/paperboy/config"
	"github.com/spf13/afero"
	"github.com/spf13/cast"
	"github.com/wneessen/go-mail-middleware/dkim"

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
	return dkim.NewFromRSAKey(keyBytes, sc)
}
