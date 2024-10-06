package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	vaultToken := "root"

	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = "8085"
		log.Println("PORT environment variable not set, defaulting to", port)
	}

	vaultURL := os.Getenv("VAULT_ADDR")
	if vaultURL == "" {
		vaultURL = "http://vault:8200"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received Request - Port forwarding is working.")

		// If the JWT path is setup then get the new token from Vault using the k8s Auth
		jwtPath := os.Getenv("JWT_PATH")
		if jwtPath != "" {
			jwtFile, err := os.ReadFile(jwtPath)
			if err != nil {
				fmt.Println("Error reading JWT file at", jwtPath, ": ", err)
				return
			}

			jwt := string(jwtFile)
			fmt.Println("Read JWT:", jwt)

			authPath := os.Getenv("AUTH_PATH")
			if authPath == "" {
				authPath = "auth/kubernetes/login"
			}

			role := os.Getenv("ROLE")
			fmt.Println("ROLE:", role)

			// Create the payload for Vault authentication
			pl := VaultJWTPayload { Role: role, JWT: jwt }
			jwtPayload, err := json.Marshal(pl)
			if err != nil {
				fmt.Println("Error encoding Vault request JSON:", err)
				return
			}

			// Send a request to Vault to retrieve a token
			vaultLoginResponse := &VaultLoginResponse{}
			err = SendRequest(vaultURL + "/v1/" + authPath, "", "POST", jwtPayload, vaultLoginResponse)
			if err != nil {
				fmt.Println("Error getting response from Vault k8s login:", err)
				return
			}
			fmt.Println("Retrieved login response: ", vaultLoginResponse)
			vaultToken = vaultLoginResponse.Auth.ClientToken
			fmt.Println("Retrieved token: ", vaultToken)
		}

		secretsPath := "secret/data/webapp/config"

		// Send a request to Vault using the token to retrieve the secret
		vaultSecretResponse := &VaultSecretResponse{}
		err := SendRequest(vaultURL + "/v1/" + secretsPath, vaultToken, "GET", nil, &vaultSecretResponse)
		if err != nil {
			fmt.Println("Error getting secret from Vault:", err)
			return
		}

		secretResponseData, ok := vaultSecretResponse.Data.Data.(map[string]interface{})
		fmt.Print(secretResponseData)
		if ok {
			for key, value := range secretResponseData {
				fmt.Fprintf(w, "%s:%s ",  key, value)
			}
		} else {
			fmt.Println("Error getting the secret from Vault, cannot convert Data to map[string]interface{}")
		}
	})

	log.Println("Listening on port", port)
	if err := http.ListenAndServe(":" + port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// SendRequest sends a request
func SendRequest(url string, token string, requestType string, payload []byte, target interface{}) error {
	req, err := http.NewRequest(requestType, url, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}

	tr := &http.Transport {
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("X-Vault-Token", token)
	}

	client := &http.Client{Timeout: 10 * time.Second, Transport: tr}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request to Vault:", err)
		return err
	}
	defer res.Body.Close()
	return json.NewDecoder(res.Body).Decode(target)
}