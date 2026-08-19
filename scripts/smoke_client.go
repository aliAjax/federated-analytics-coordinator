//go:build ignore

package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const address = "http://127.0.0.1:18081"

func main() {
	pub1, priv1, _ := ed25519.GenerateKey(rand.Reader)
	pub2, priv2, _ := ed25519.GenerateKey(rand.Reader)
	p1 := participant("one", pub1)
	p2 := participant("two", pub2)
	deadline := time.Now().UTC().Add(time.Hour).Format(time.RFC3339)
	task := request("POST", "/api/v1/tasks", map[string]any{"template_id": "count-v1", "metric": "count", "participants": []string{p1, p2}, "minimum_participants": 2, "deadline": deadline, "epsilon": 1.0})
	taskID := getString(task, "id")
	request("POST", "/api/v1/tasks/"+taskID+"/start", nil)
	upload(taskID, p1, priv1, 7, "nonce-one")
	upload(taskID, p2, priv2, 5, "nonce-two")
	result := request("POST", "/api/v1/tasks/"+taskID+"/aggregate", nil)
	if value := getFloat(result, "value"); value != 12 {
		panic(fmt.Sprintf("want 12 got %v", value))
	}
	request("GET", "/api/v1/results/"+taskID, nil)
	fmt.Println("smoke passed")
}
func participant(name string, pub ed25519.PublicKey) string {
	return getString(request("POST", "/api/v1/participants", map[string]any{"name": name, "public_key": base64.StdEncoding.EncodeToString(pub), "capabilities": []map[string]any{{"metric": "count"}}}), "id")
}
func upload(task, participant string, key ed25519.PrivateKey, value int64, nonce string) {
	canonical := fmt.Sprintf("%s|%s|%d|%d|%d|%s", task, participant, 1, value, 1, nonce)
	sig := ed25519.Sign(key, []byte(canonical))
	request("POST", "/api/v1/tasks/"+task+"/shares", map[string]any{"participant_id": participant, "round": 1, "value": value, "count": 1, "nonce": nonce, "signature": base64.StdEncoding.EncodeToString(sig)})
}
func request(method, path string, payload any) map[string]any {
	var body io.Reader
	if payload != nil {
		b, e := json.Marshal(payload)
		if e != nil {
			panic(e)
		}
		body = bytes.NewReader(b)
	}
	req, e := http.NewRequest(method, address+path, body)
	if e != nil {
		panic(e)
	}
	req.Header.Set("X-Tenant-ID", "smoke")
	req.Header.Set("Content-Type", "application/json")
	response, e := http.DefaultClient.Do(req)
	if e != nil {
		panic(e)
	}
	defer response.Body.Close()
	b, _ := io.ReadAll(response.Body)
	if response.StatusCode < 200 || response.StatusCode > 299 {
		fmt.Fprintln(os.Stderr, string(b))
		panic(response.Status)
	}
	var v struct {
		Data map[string]any `json:"data"`
	}
	if e = json.Unmarshal(b, &v); e != nil {
		panic(e)
	}
	return v.Data
}
func getString(v map[string]any, key string) string { return v[key].(string) }
func getFloat(v map[string]any, key string) float64 { return v[key].(float64) }
