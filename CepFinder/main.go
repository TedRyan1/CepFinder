package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	API1 = "https://cdn.apicep.com/file/apicep/%s.json"
	API2 = "http://viacep.com.br/ws/%s/json/"
)

type Address struct {
	CEP       string `json:"cep"`
	Logradouro string `json:"logradouro"`
	Bairro    string `json:"bairro"`
	Cidade    string `json:"localidade"`
	Estado    string `json:"uf"`
}

type Response struct {
	Address *Address
	API     string
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
	address := &Address{}
	json.Unmarshal(data, address)

	ch <- Response{Address: address, API: apiURL}
}

func main() {
	fmt.Print("Informe o CEP: ")
	var cep string
	fmt.Scanln(&cep)

	ch := make(chan Response, 2)

	go fetchFromAPI(cep, API1, ch)
	go fetchFromAPI(cep, API2, ch)

	select {
	case result := <-ch:
		if result.Address != nil {
			fmt.Printf("Resultado de %s:\n", result.API)
			fmt.Printf("CEP: %s, Logradouro: %s, Bairro: %s, Cidade: %s, Estado: %s\n",
				result.Address.CEP, result.Address.Logradouro, result.Address.Bairro,
				result.Address.Cidade, result.Address.Estado)
		} else {
			fmt.Printf("Erro na API %s\n", result.API)
		}
	case <-time.After(1 * time.Second):
		fmt.Println("Erro: timeout excedido para ambas as APIs.")
	}
}