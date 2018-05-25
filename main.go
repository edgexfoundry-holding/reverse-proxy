package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/dghubble/sling"
)

const (
	servicesPath     = "services/"
	consumersPath    = "consumers/"
	certificatesPath = "certificates/"
)

func main() {
	//read config from toml file
	//TODO: get data from consul to overwrite the local default config
	//TODO: change the log format from console output to log service of edgex
	config := LoadTomlConfig("res/configuration.toml")

	baseURL := fmt.Sprintf("http://%s:%s/", config.KongUrl.Server, config.KongUrl.AdminPort)
	client := &http.Client{}

	checkKongStatus(baseURL, client)

	for _, service := range config.EdgexServices {
		serviceParams := &KongService{
			Name:     service.Name,
			Host:     service.Host,
			Port:     service.Port,
			Protocol: service.Protocol,
		}

		initKongServices(baseURL, client, serviceParams)

		jwtServicePath := fmt.Sprintf("services/%s/plugins", service.Name)
		initKongJWT(baseURL, client, jwtServicePath, service.Name)
	}

	for _, service := range config.EdgexServices {
		routeParams := &KongRoute{
			Paths: []string{"/" + service.Name},
		}
		routePath := fmt.Sprintf("services/%s/routes", service.Name)
		initKongRoutes(baseURL, client, routeParams, routePath, service.Name)
	}

	initKongAdminInterface(config, baseURL, client)
	loadKongCerts(baseURL, client)
	os.Exit(0)
}

func checkKongStatus(url string, c *http.Client) {
	req, err := sling.New().Get(url).Request()
	resp, err := c.Do(req)
	if err != nil {
		fmt.Println("KONG is failed to start up, please verify...")
		os.Exit(0)
	} else {
		fmt.Println(resp.StatusCode)
		if resp.StatusCode == 200 {
			fmt.Println("Ping successful with KONG service.")
		} else {
			fmt.Println("Failed to ping KONG service. Please check KONG service status.")
		}
	}
}

func initKongServices(url string, c *http.Client, service *KongService) {
	req, err := sling.New().Base(url).Post(servicesPath).BodyForm(service).Request()
	resp, err := c.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Failed to set up service for " + service.Name)
	} else {
		fmt.Println(resp.StatusCode)
		if resp.StatusCode == 201 || resp.StatusCode == 409 {
			fmt.Println("Successful to set up service for " + service.Name)
		} else {
			fmt.Println("Failed to set up service for " + service.Name)
		}
	}
}

func initKongJWT(url string, c *http.Client, path string, name string) {
	jwtParams := &KongPlugin{
		Name: "jwt",
	}

	req, err := sling.New().Base(url).Post(path).BodyForm(jwtParams).Request()
	resp, err := c.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Failed to set up jwt authentication for service " + name)
	} else {
		fmt.Println(resp.StatusCode)
		if resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 409 {
			fmt.Println("Successful to set up jwt authentication for service " + name)
		} else {
			fmt.Println("Failed to set up jwt authentication for service " + name)
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
			fmt.Println("Successful to set up route for " + name)
		} else {
			fmt.Println("Failed to set up route for " + name)
		}
	}

}

func initKongAdminInterface(config *tomlConfig, url string, c *http.Client) {
	//redirect request for 8001 to an admin service of 8000, and add authentication
	adminServiceParams := &KongService{
		Name:     "admin",
		Host:     config.KongUrl.Server,
		Port:     config.KongUrl.AdminPort,
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

	//enable basic-auth for admin service
	//url -X POST http://kong:8001/services/{service}/plugins \
	//--data "name=basic-auth"  \
	//--data "config.hide_credentials=true"
	basicAuthParams := &KongBasicAuthPlugin{
		Name:            "basic-auth",
		HideCredentials: "true",
	}
	req, err = sling.New().Base(url).Post("services/admin/plugins/").BodyForm(basicAuthParams).Request()
	resp, err = c.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Failed to enable basic-auth for admin service.")
	} else {
		fmt.Println(resp.StatusCode)
		if resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 409 {
			fmt.Println("Successful to enable basic-auth for admin service.")
		} else {
			fmt.Println("Failed to enable basic-auth for admin service.")
		}
	}

	//create consumer administrator so that it can be used to consume admin service
	userNameParams := &KongUser{UserName: config.KongAdmin.UserName}
	req, err = sling.New().Base(url).Post(consumersPath).BodyForm(userNameParams).Request()
	resp, err = c.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Failed to create consumer for admin service.")
	} else {
		fmt.Println(resp.StatusCode)
		if resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 409 {
			fmt.Println("Successful to create consumer for admin service.")
		} else {
			fmt.Println("Failed to create consumer for admin service.")
		}
	}

	//curl -X POST http://kong:8001/consumers/{consumer}/basic-auth \
	//--data "username=administrator" \
	//--data "password=changeme"
	adminCredential := &KongUser{
		UserName: config.KongAdmin.UserName,
		Password: config.KongAdmin.Password,
	}
	adminAuthPath := fmt.Sprintf("consumers/%s/basic-auth", config.KongAdmin.UserName)
	req, err = sling.New().Base(url).Post(adminAuthPath).BodyForm(adminCredential).Request()
	resp, err = c.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println(fmt.Sprintf("Failed to add credential for consumer %s.", config.KongAdmin.UserName))
	} else {
		fmt.Println(resp.StatusCode)
		if resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 409 {
			fmt.Println(fmt.Sprintf("Successful to add credential for consumer %s.", config.KongAdmin.UserName))
		} else {
			fmt.Println(fmt.Sprintf("Failed to add credential for consumer %s.", config.KongAdmin.UserName))
		}
	}

}

func loadKongCerts(url string, c *http.Client) {
	certInfo := &KongCert{
		Cert: "pemcert",
		Key:  "key",
		Snis: "",
	}
	req, err := sling.New().Base(url).Post(certificatesPath).BodyForm(certInfo).Request()
	resp, err := c.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println(fmt.Sprintf("Failed to add certificate with cert path as %s.", "pem-path"))
	} else {
		fmt.Println(resp.StatusCode)
		if resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 409 {
			fmt.Println(fmt.Sprintf("Successful to add cert."))
		} else {
			fmt.Println(fmt.Sprintf("Failed to add cert."))
		}
	}
}
