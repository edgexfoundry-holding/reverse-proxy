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
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dghubble/sling"
	"github.com/dgrijalva/jwt-go"
)

const (
	servicesPath     = "services/"
	consumersPath    = "consumers/"
	certificatesPath = "certificates/"
	vaultToken       = "X-Vault-Token"
)

func main() {
	//read config from toml file
	//TODO: get data from consul to overwrite the local default config
	//TODO: change the log format from console output to log service of edgex
	config := LoadTomlConfig("res/configuration.toml")
	proxyBaseURL := fmt.Sprintf("http://%s:%s/", config.KongURL.Server, config.KongURL.AdminPort)
	secretServiceBaseURL := fmt.Sprintf("http://%s:%s/", config.SecretService.Server, config.SecretService.Port)
	client := &http.Client{Timeout: 10 * time.Second}

	checkProxyStatus(proxyBaseURL, client)
	checkSecretServiceStatus(secretServiceBaseURL+config.SecretService.HealthcheckPath, client)

	for _, service := range config.EdgexServices {
		serviceParams := &KongService{
			Name:     service.Name,
			Host:     service.Host,
			Port:     service.Port,
			Protocol: service.Protocol,
		}

		initKongService(proxyBaseURL, client, serviceParams)
		jwtServicePath := fmt.Sprintf("services/%s/plugins", service.Name)
		initJWTAuthForService(proxyBaseURL, client, jwtServicePath, service.Name)
	}

	for _, service := range config.EdgexServices {
		routeParams := &KongRoute{
			Paths: []string{"/" + service.Name},
		}
		routePath := fmt.Sprintf("services/%s/routes", service.Name)
		initKongRoutes(proxyBaseURL, client, routeParams, routePath, service.Name)
	}

	initKongAdminInterface(config, proxyBaseURL, client)
	//loadKongCerts(config, proxyBaseURL, client)

}

func checkProxyStatus(url string, c *http.Client) {
	req, err := sling.New().Get(url).Request()
	resp, err := c.Do(req)
	if err != nil {
		fmt.Println("The status of reverse proxy is unknown, the initialization is terminated.")
		os.Exit(0)
	} else {
		fmt.Println(resp.StatusCode)
		if resp.StatusCode == 200 {
			fmt.Println("Reverse proxy is up successfully.")
		} else {
			fmt.Println("The status of reverse proxy is unknown, the initialization is terminated.")
			os.Exit(0)
		}
	}
}

func initKongService(url string, c *http.Client, service *KongService) {
	req, err := sling.New().Base(url).Post(servicesPath).BodyForm(service).Request()
	resp, err := c.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println(fmt.Sprintf("Failed to set up proxy service for %s.", service.Name))
	} else {
		fmt.Println(resp.StatusCode)
		if resp.StatusCode == 201 || resp.StatusCode == 409 {
			fmt.Println(fmt.Sprintf("Successful to set up proxy service for %s.", service.Name))
		} else {
			fmt.Println(fmt.Sprintf("Failed to set up proxy service for %s.", service.Name))
		}
	}
}

func initJWTAuthForService(url string, c *http.Client, path string, name string) {
	jwtParams := &KongPlugin{
		Name: "jwt",
	}

	req, err := sling.New().Base(url).Post(path).BodyForm(jwtParams).Request()
	resp, err := c.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println(fmt.Sprintf("Failed to set up jwt authentication for service %s.", name))
	} else {
		fmt.Println(resp.StatusCode)
		if resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 409 {
			fmt.Println(fmt.Sprintf("Successful to set up jwt authentication for service %s.", name))
		} else {
			fmt.Println(fmt.Sprintf("Failed to set up jwt authentication for service %s.", name))
		}
	}
}

func initKongRoutes(url string, c *http.Client, r *KongRoute, path string, name string) {
	req, err := sling.New().Base(url).Post(path).BodyForm(r).Request()
	resp, err := c.Do(req)
	if err != nil {
		log.Println(err.Error())
	} else {
		fmt.Println(resp.StatusCode)
		if resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 409 {
			fmt.Println(fmt.Sprintf("Successful to set up route for %s.", name))
		} else {
			fmt.Println(fmt.Sprintf("Failed to set up route for %s.", name))
		}
	}
}

//redirect request for 8001 to an admin service of 8000, and add authentication
func initKongAdminInterface(config *tomlConfig, url string, c *http.Client) {
	adminServiceParams := &KongService{
		Name:     "admin",
		Host:     config.KongURL.Server,
		Port:     config.KongURL.AdminPort,
		Protocol: "http",
	}
	req, err := sling.New().Base(url).Post(servicesPath).BodyForm(adminServiceParams).Request()
	resp, err := c.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Failed to set up service for admin loopback.")
	} else {
		fmt.Println(resp.StatusCode)
		if resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 409 {
			fmt.Println("Successful to set up admin loopback.")
		} else {
			fmt.Println("Failed to set up admin loopback.")
		}
	}

	adminRouteParams := &KongRoute{Paths: []string{"/admin"}}
	adminRoutePath := fmt.Sprintf("%sadmin/routes", servicesPath)
	req, err = sling.New().Base(url).Post(adminRoutePath).BodyForm(adminRouteParams).Request()
	resp, err = c.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Failed to set up admin service route.")
	} else {
		fmt.Println(resp.StatusCode)
		if resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 409 {
			fmt.Println("Successful to set up admin service routes.")
		} else {
			fmt.Println("Failed to set up admin service routes.")
		}
	}

	jwtAdminServicePath := "services/admin/plugins"
	initJWTAuthForService(url, c, jwtAdminServicePath, "admin")

	createConsumer(config.KongAdmin.UserName, url, c, "admin")

	t, err := createJWTForConsumer(config.KongAdmin.UserName, url, c, "admin")
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Failed to create jwt token for admin service.")
	} else {
		fmt.Println(fmt.Sprintf("The JWT for consumer %s is: %s. Please keep the jwt for future use.", config.KongAdmin.UserName, t))
	}
}

func createConsumer(user string, url string, c *http.Client, name string) {
	userNameParams := &KongUser{UserName: user}
	req, err := sling.New().Base(url).Post(consumersPath).BodyForm(userNameParams).Request()
	resp, err := c.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println(fmt.Sprintf("Failed to create consumer for %s service.", name))
		os.Exit(0)
	} else {
		fmt.Println(resp.StatusCode)
		if resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 409 {
			fmt.Println(fmt.Sprintf("Successful to create consumer for %s service.", name))
		} else {
			fmt.Println(fmt.Sprintf("Failed to create consumer for %s service.", name))
		}
	}
}

func createJWTForConsumer(user string, url string, c *http.Client, name string) (string, error) {
	jwtCred := JWTCred{}
	s := sling.New().Set("Content-Type", "application/x-www-form-urlencoded")
	req, err := s.New().Get(url).Post(fmt.Sprintf("consumers/%s/jwt", user)).Request()
	resp, err := c.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println(fmt.Sprintf("Failed to create jwt token for consumer %s.", user))
		return "", errors.New("Error: unable to create JWT")
	}
	fmt.Println(resp.StatusCode)
	if resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 409 {
		fmt.Println(fmt.Sprintf("Successful to create JWT for consumer %s.", user))
		defer resp.Body.Close()
		json.NewDecoder(resp.Body).Decode(&jwtCred)
		fmt.Println(fmt.Sprintf("successful on retrieving JWT credential for consumer %s.", user))

		// Create the Claims
		claims := KongJWTClaims{
			jwtCred.Key,
			user,
			jwt.StandardClaims{
				Issuer: "edgex",
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		return token.SignedString([]byte(jwtCred.Secret))
	}
	fmt.Println(fmt.Sprintf("Failed to create JWT for consumer %s.", user))
	return "", errors.New("Error: unable to create JWT")
}

func checkSecretServiceStatus(url string, c *http.Client) {
	req, err := sling.New().Get(url).Request()
	resp, err := c.Do(req)
	if err != nil {
		fmt.Println("The status of secret service is unknown, the initialization is terminated.")
		os.Exit(0)
	} else {
		fmt.Println(resp.StatusCode)
		if resp.StatusCode == 200 {
			fmt.Println("Secret management service is up successfully.")
		} else {
			fmt.Println(fmt.Sprintf("Secret management service is down. Please check the status of secret service with endpoint %s.", url))
			os.Exit(0)
		}
	}
}

func loadKongCerts(config *tomlConfig, url string, c *http.Client) {
	cert, key, err := getCertKeyPair(config, c)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}

	body := &CertInfo{
		Cert: cert,
		Key:  key,
		Snis: []string{config.SecretService.SNIS},
	}
	req, err := sling.New().Base(url).Post(certificatesPath).BodyJSON(body).Request()
	resp, err := c.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println(fmt.Sprintf("Failed to add certificate with cert path as %s.", config.SecretService.CertPath))
	} else {
		fmt.Println(resp.StatusCode)
		if resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 409 {
			fmt.Println(fmt.Sprintf("Successful to add certificate to the reverse proxy."))
		} else {
			fmt.Println(fmt.Sprintf("Failed to add certificate to the reverse proxy."))
		}
	}
}

func getCertKeyPair(config *tomlConfig, c *http.Client) (string, string, error) {
	certs := Cert{}
	s := sling.New().Set(vaultToken, config.SecretService.Token)
	req, err := s.New().Get(config.SecretService.CertPath).Request()
	resp, err := c.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println(fmt.Sprintf("Failed to retrieve certificate with path as %s.", config.SecretService.CertPath))
	} else {
		defer resp.Body.Close()
		json.NewDecoder(resp.Body).Decode(&certs)
		fmt.Println(fmt.Sprintf("successful on retrieving certificate from %s.", config.SecretService.CertPath))
		return certs.Data.Cert, certs.Data.Key, nil
	}
	errInfo := fmt.Sprintf("Failed to retrieve certificate with path as %s.", config.SecretService.CertPath)
	fmt.Println(errInfo)
	return "", "", errors.New(errInfo)
}
