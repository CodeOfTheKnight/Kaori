package kaoriData

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func NormalizeEpNumber(eps []float64) (name string) {

	for i, ep := range eps {
		if i != 0 {
			name += "-"
		}
		name += fmt.Sprintf("%.1f", ep)

		tmp := strings.Split(name, ".")
		if tmp[1] == "0" {
			name = tmp[0]
		}
	}

	return name
}

func SendToKaori(obj interface{}, kaoriUrl string, token string) error {

	//Create JSON
	data, err := json.MarshalIndent(obj, " ", "\t")
	if err != nil {
		return errors.New("Error to create JSON: " + err.Error())
	}

	fmt.Println("DATA:", string(data))

	//Create client
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	//Create request
	req, err := http.NewRequest("PUT", kaoriUrl, bytes.NewReader(data))
	if err != nil {
		return errors.New("Error to create request: " + err.Error())
	}
	req.Header.Set("Authorization", "Bearer " + token)

	//Do request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("Error to send data, status code: " + resp.Status)
	}

	return nil
}

func AppendFile(obj interface{}, filePath string) error {

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()

	data, err := json.MarshalIndent(obj, " ", "\t")
	if err != nil {
		return errors.New("Error to create JSON: " + err.Error())
	}

	_, err = file.Write([]byte(string(data) + ",\n"))
	if err != nil {
		return err
	}

	return nil
}