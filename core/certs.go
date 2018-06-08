/*******************************************************************************
 * Copyright 2018 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 * @author: Tingyu Zeng, Dell
 * @version: 0.1.0
 *******************************************************************************/
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/dghubble/sling"
)

func loadKongCerts(config *tomlConfig, url string, c *http.Client) error {
	cert, key, err := getCertKeyPair(config, c)
	if err != nil {
		return err
	}
	body := &CertInfo{
		Cert: cert,
		Key:  key,
		Snis: []string{config.SecretService.SNIS},
	}
	req, err := sling.New().Base(url).Post(CertificatesPath).BodyJSON(body).Request()
	resp, err := c.Do(req)
	if err != nil {
		s := fmt.Sprintf("Failed to add certificate with cert path of %s with error %s.", config.SecretService.CertPath, err.Error())
		return errors.New(s)
	} else {
		if resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 409 {
			lc.Info("Successful to add certificate to the reverse proxy.")
		} else {
			s := fmt.Sprintf("Failed to add certificate with cert path of %s with errorcode %s.", config.SecretService.CertPath, resp.StatusCode)
			return errors.New(s)
		}
	}
	return nil
}

func getCertKeyPair(config *tomlConfig, c *http.Client) (string, string, error) {
	certs := Cert{}
	s := sling.New().Set(VaultToken, config.SecretService.Token)
	req, err := s.New().Get(config.SecretService.CertPath).Request()
	resp, err := c.Do(req)
	if err != nil {
		errStr := fmt.Sprintf("Failed to retrieve certificate with path as %s with error %s", config.SecretService.CertPath, err.Error())
		return "", "", errors.New(errStr)
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&certs)
	lc.Info(fmt.Sprintf("successful on retrieving certificate from %s.", config.SecretService.CertPath))
	return certs.Data.Cert, certs.Data.Key, nil
}
