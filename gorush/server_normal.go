// +build !lambda

package gorush

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"net/http"
)

// RunHTTPServer provide run http or https protocol.
func RunHTTPServer() (err error) {
	if !PushConf.Core.Enabled {
		LogAccess.Debug("httpd server is disabled.")
		return nil
	}

	server := &http.Server{
		Addr:    PushConf.Core.Address + ":" + PushConf.Core.Port,
		Handler: routerEngine(),
	}

	LogAccess.Debug("HTTPD server is running on " + PushConf.Core.Port + " port.")
	if PushConf.Core.AutoTLS.Enabled {
		return startServer(autoTLSServer())
	} else if PushConf.Core.SSL {
		config := &tls.Config{
			MinVersion: tls.VersionTLS10,
		}

		if config.NextProtos == nil {
			config.NextProtos = []string{"http/1.1"}
		}

		config.Certificates = make([]tls.Certificate, 1)
		if PushConf.Core.CertPath != "" && PushConf.Core.KeyPath != "" {
			config.Certificates[0], err = tls.LoadX509KeyPair(PushConf.Core.CertPath, PushConf.Core.KeyPath)
			if err != nil {
				LogError.Error("Failed to load https cert file: ", err)
				return err
			}
		} else if PushConf.Core.CertBase64 != "" && PushConf.Core.KeyBase64 != "" {
			cert, err := base64.StdEncoding.DecodeString(PushConf.Core.CertBase64)
			if err != nil {
				LogError.Error("base64 decode error:", err.Error())
				return err
			}
			key, err := base64.StdEncoding.DecodeString(PushConf.Core.KeyBase64)
			if err != nil {
				LogError.Error("base64 decode error:", err.Error())
				return err
			}
			if config.Certificates[0], err = tls.X509KeyPair(cert, key); err != nil {
				LogError.Error("tls key pair error:", err.Error())
				return err
			}
		} else {
			return errors.New("missing https cert config")
		}

		server.TLSConfig = config
	}

	return startServer(server)
}

func startServer(s *http.Server) error {
	if s.TLSConfig == nil {
		return s.ListenAndServe()
	}
	return s.ListenAndServeTLS("", "")
}
