/*
 * Copyright 2018 Kopano and its licensors
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License, version 3,
 * as published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package main

import (
	"fmt"
	"os"

	kcc "stash.kopano.io/kgol/kcc-go"

	"stash.kopano.io/kc/konnect/identifier"
	identifierBackends "stash.kopano.io/kc/konnect/identifier/backends"
	"stash.kopano.io/kc/konnect/identity"
	identityManagers "stash.kopano.io/kc/konnect/identity/managers"
)

func newKCIdentityManager(bs *bootstrap) (identity.Manager, error) {
	logger := bs.cfg.Logger

	if bs.authorizationEndpointURI.String() != "" {
		return nil, fmt.Errorf("kc backend is incompatible with authorization-endpoint-uri parameter")
	}
	bs.authorizationEndpointURI.Path = "/signin/v1/identifier/_/authorize"

	if bs.endSessionEndpointURI.String() != "" {
		return nil, fmt.Errorf("kc backend is incompatible with endsession-endpoint-uri parameter")
	}
	bs.endSessionEndpointURI.Path = "/signin/v1/identifier/_/endsession"

	if bs.signInFormURI.EscapedPath() == "" {
		bs.signInFormURI.Path = "/signin/v1/identifier"
	}

	if bs.signedOutURI.EscapedPath() == "" {
		bs.signedOutURI.Path = "/signin/v1/goodbye"
	}

	identifierBackend, identifierErr := identifierBackends.NewKCIdentifierBackend(
		bs.cfg,
		kcc.NewKCC(nil),
		os.Getenv("KOPANO_SERVER_USERNAME"),
		os.Getenv("KOPANO_SERVER_PASSWORD"),
	)
	if identifierErr != nil {
		return nil, fmt.Errorf("failed to create identifier backend: %v", identifierErr)
	}

	fullAuthorizationEndpointURL := withSchemeAndHost(bs.authorizationEndpointURI, bs.issuerIdentifierURI)

	activeIdentifier, err := identifier.NewIdentifier(&identifier.Config{
		Config: bs.cfg,

		PathPrefix:      "/signin/v1",
		StaticFolder:    bs.identifierClientPath,
		LogonCookieName: "__Secure-KKT", // Kopano-Konnect-Token

		AuthorizationEndpointURI: fullAuthorizationEndpointURL,

		Backend: identifierBackend,
		Clients: bs.managers.clients,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create identifier: %v", err)
	}
	err = activeIdentifier.SetKey(bs.encryptionSecret)
	if err != nil {
		return nil, fmt.Errorf("invalid --encryption-secret parameter value for identifier: %v", err)
	}

	identityManagerConfig := &identity.Config{
		SignInFormURI: withSchemeAndHost(bs.signInFormURI, bs.issuerIdentifierURI),
		SignedOutURI:  withSchemeAndHost(bs.signedOutURI, bs.issuerIdentifierURI),

		Logger: logger,

		ScopesSupported: bs.cfg.AllowedScopes,
	}

	identifierIdentityManager := identityManagers.NewIdentifierIdentityManager(identityManagerConfig, activeIdentifier, bs.managers.clients)
	logger.Infoln("using identifier backed identity manager")

	return identifierIdentityManager, nil
}
