package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	API1 = "https://cdn.apicep.com/file/apicep/%s.json"
	API2 = "http://viacep.com.br/ws/%s/json/"
)

type AddressViaCEP struct {
	CEP       string `json:"cep"`
	Logradouro string `json:"logradouro"`
	Bairro    string `json:"bairro"`
	Cidade    string `json:"localidade"`
	Estado    string `json:"uf"`
}

type AddressAPIcep struct {
	CEP       string `json:"code"`
	Logradouro string `json:"address"`
	Bairro    string `json:"district"`
	Cidade    string `json:"city"`
	Estado    string `json:"state"`
}

type Response struct {
	AddressViaCEP *AddressViaCEP
	AddressAPIcep *AddressAPIcep
	API           string
}

func cleanCEP(cep string) string {
	return strings.ReplaceAll(cep, "-", "")
}

func formatCEP(cep string) string {
	return cep[:5] + "-" + cep[5:] 
}

func fetchFromAPI(cep, apiURL string, ch chan Response) {
	client := http.Client{
		Timeout: 1 * time.Second,
	}

	resp, err := client.Get(fmt.Sprintf(apiURL, cep))
	if err != nil {
		ch <- Response{API: apiURL}
		return
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	if apiURL == API1 {
		address := &AddressAPIcep{}
		json.Unmarshal(data, address)
		ch <- Response{AddressAPIcep: address, API: apiURL}
		return
	}

	address := &AddressViaCEP{}
	json.Unmarshal(data, address)
	ch <- Response{AddressViaCEP: address, API: apiURL}
}

func main() {
	cep := getCEPFromUser()
	if cep == "" {
		return
	} 

	ch := make(chan Response, 2)
    go fetchFromAPI(formatCEP(cep), API1, ch) 
	go fetchFromAPI(cep, API2, ch)  
	displayResult(ch)
}

func getCEPFromUser() string {
	fmt.Print("Informe o CEP: ")
	var inputCEP string
	fmt.Scanln(&inputCEP)

	cleanedCEP := cleanCEP(inputCEP)

	if len(cleanedCEP) != 8 {
		fmt.Println("Formato inválido. O CEP deve ter 8 dígitos.")
		return ""
	}
	return cleanedCEP
}

func displayResult(ch chan Response) {
	select {
	case result := <-ch:
		printAPIResponse(result)
	case <-time.After(1 * time.Second):
		fmt.Println("Erro: timeout excedido para ambas as APIs.")
	}
}

func printAPIResponse(result Response) {
	if result.API == API1 && result.AddressAPIcep.CEP != "" {
		fmt.Printf("Resultado de %s:\n", result.API)
		fmt.Printf("CEP: %s, Logradouro: %s, Bairro: %s, Cidade: %s, Estado: %s\n",
			result.AddressAPIcep.CEP, result.AddressAPIcep.Logradouro, result.AddressAPIcep.Bairro,
			result.AddressAPIcep.Cidade, result.AddressAPIcep.Estado)
		return
	}

	if result.API == API2 && result.AddressViaCEP.CEP != "" {
		fmt.Printf("Resultado de %s:\n", result.API)
		fmt.Printf("CEP: %s, Logradouro: %s, Bairro: %s, Cidade: %s, Estado: %s\n",
			result.AddressViaCEP.CEP, result.AddressViaCEP.Logradouro, result.AddressViaCEP.Bairro,
			result.AddressViaCEP.Cidade, result.AddressViaCEP.Estado)
		return
	}

	fmt.Printf("Erro na API %s\n", result.API)
} 
