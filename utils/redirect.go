/*
 * Copyright 2017 Kopano and its licensors
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

package utils

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/go-querystring/query"
)

// WriteRedirect crates a URL out of the provided uri and params and writes a
// a HTTP response with the provided HTTP status code to the provided
// http.ResponseWriter incliding HTTP caching headers to prevent caching. If
// asFragment is true, the provided params are added as URL fragment, otherwise
// they replace the query. If params is nil, the provided uri is taken as is.
func WriteRedirect(rw http.ResponseWriter, code int, uri *url.URL, params interface{}, asFragment bool) error {
	uriString := uri.String()

	if params != nil {
		queryString, err := query.Values(params)
		if err != nil {
			return err
		}

		seperator := "#"
		if !asFragment {
			seperator = "?"
		}

		if strings.Contains(uriString, seperator) {
			// Avoid generating invalid URLs if the seperator is already part
			// of the target URL - instead append it in the most likely way.
			seperator = "&"
		}
		queryStringEncoded := strings.Replace(queryString.Encode(), "+", "%20", -1) // NOTE(longsleep): Ensure we use %20 instead of +.
		uriString = fmt.Sprintf("%s%s%s", uriString, seperator, queryStringEncoded)
	}

	rw.Header().Set("Location", uriString)
	rw.Header().Set("Cache-Control", "no-store")
	rw.Header().Set("Pragma", "no-cache")

	rw.WriteHeader(code)

	return nil
}
