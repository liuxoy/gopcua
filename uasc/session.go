package uasc

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/wmnsk/gopcua/datatypes"
	"github.com/wmnsk/gopcua/services"
)

type Session struct {
	sechan *SecureChannel
	cfg    *SessionConfig

	maxRequestMessageSize uint32

	// mySignature is is the client/serverSignature expected to receive from the other endpoint.
	// This parameter is automatically calculated and kept temporarily until being used to verify
	// received client/serverSignature.
	mySignature *services.SignatureData

	// signatureToSend is the client/serverSignature defined in Part4, Table 15 and Table 17.
	// This parameter is automatically calculated and kept temporarily until it is sent in next message.
	signatureToSend *services.SignatureData
}

func NewSession(sechan *SecureChannel, cfg *SessionConfig) *Session {
	return &Session{sechan: sechan, cfg: cfg}
}

func (s *Session) Open() error {
	if err := s.createSession(); err != nil {
		return err
	}
	return s.activateSession()
}

func (s *Session) Close() error {
	return nil
}

func (s *Session) createSession() error {
	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		return err
	}

	req := &services.CreateSessionRequest{
		ClientDescription: s.cfg.ClientDescription,
		EndpointURL:       s.sechan.EndpointURL,
		SessionName:       fmt.Sprintf("gopcua-%d", time.Now().UnixNano()),
		ClientNonce:       nonce,
		ClientCertificate: s.sechan.cfg.Certificate,
	}

	return s.sechan.Send(req, func(v interface{}) error {
		resp, ok := v.(*services.CreateSessionResponse)
		if !ok {
			return fmt.Errorf("invalid response. Got %T, want CreateSessionResponse", v)
		}

		s.sechan.reqhdr.AuthenticationToken = resp.AuthenticationToken
		s.cfg.ServerEndpoints = resp.ServerEndpoints
		s.cfg.SessionTimeout = resp.RevisedSessionTimeout
		s.signatureToSend = services.NewSignatureDataFrom(resp.ServerCertificate, resp.ServerNonce)
		s.maxRequestMessageSize = resp.MaxRequestMessageSize

		// keep SignatureData to verify serverSignature in CreateSessionResponse.
		s.mySignature = services.NewSignatureDataFrom(s.sechan.cfg.Certificate, nonce)
		return nil
	})
}

func (s *Session) activateSession() error {
	req := &services.ActivateSessionRequest{
		ClientSignature:            s.signatureToSend,
		ClientSoftwareCertificates: nil,
		LocaleIDs:                  s.cfg.LocaleIDs,
		UserIdentityToken:          datatypes.NewExtensionObject(1, s.cfg.UserIdentityToken),
		UserTokenSignature:         s.cfg.UserTokenSignature,
	}
	return s.sechan.Send(req, func(v interface{}) error {
		resp, ok := v.(*services.ActivateSessionResponse)
		if !ok {
			return fmt.Errorf("invalid response. Got %T, want ActivateSessionResponse", v)
		}

		for _, result := range resp.Results {
			if result != 0 {
				return fmt.Errorf("rejected")
			}
		}
		return nil
	})
}
