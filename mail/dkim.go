package mail

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/go-gomail/gomail"
	"github.com/spf13/cast"
	"github.com/toorop/go-dkim"
)

func SendCloserWithDKIM(sc gomail.SendCloser, conf map[string]interface{}) (gomail.SendCloser, error) {
	dOpts := dkim.NewSigOptions()

	// Required: Read private key from keyFile
	keyFile := cast.ToString(conf["keyfile"])
	if keyFile == "" {
		return nil, fmt.Errorf("DKIM requires a keyFile")
	}

	keyBytes, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}

	// Required configuration
	dOpts.PrivateKey = keyBytes
	dOpts.Domain = cast.ToString(conf["domain"])
	dOpts.Selector = cast.ToString(conf["selector"])

	// Optional configuration
	if v, ok := conf["signatureexpirein"]; ok {
		dOpts.SignatureExpireIn = cast.ToUint64(v)
	}

	if v, ok := conf["canonicalization"]; ok {
		dOpts.Canonicalization = cast.ToString(v)
	}

	// TODO: add "headers" configuration
	dOpts.Headers = []string{
		"Mime-Version", "To", "From", "Subject", "Reply-To",
		"Sender", "Content-Transfer-Encoding", "Content-Type",
	}

	return &dkimSendCloser{Options: dOpts, sc: sc}, nil
}

type dkimSendCloser struct {
	Options dkim.SigOptions
	sc      gomail.SendCloser
}

func (d *dkimSendCloser) Send(from string, to []string, msg io.WriterTo) error {
	return d.sc.Send(from, to, dkimMessage{d.Options, msg})
}

func (d *dkimSendCloser) Close() error {
	return d.sc.Close()
}

type dkimMessage struct {
	options dkim.SigOptions
	msg     io.WriterTo
}

func (dm dkimMessage) WriteTo(w io.Writer) (n int64, err error) {
	var b bytes.Buffer
	if _, err := dm.msg.WriteTo(&b); err != nil {
		return 0, err
	}

	email := b.Bytes()
	if err := dkim.Sign(&email, dm.options); err != nil {
		return 0, err
	}

	return bytes.NewBuffer(email).WriteTo(w)
}
