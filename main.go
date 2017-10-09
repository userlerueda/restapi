package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/gorilla/mux"
)

type Person struct {
	ID        string   `json:"id,omitempty"`
	Firstname string   `json:"firstname,omitempty"`
	Lastname  string   `json:"lastname,omitempty"`
	Address   *Address `json:"address,omitempty"`
}

type Service struct {
	Type           string          `json:"servicename,omitempty"`
	ID             string          `json:"id,omitempty"`
	ServiceDetails *ServiceDetails `json:"servicedetails,omitempty"`
}

type ServiceDetails struct {
	Vlan     string `json:"vlan,omitempty"`
	RemoteIP string `json:"remoteip,omitempty"`
}

type Address struct {
	City  string `json:"city,omitempty"`
	State string `json:"state,omitempty"`
}

var (
	people   []Person
	services []Service
	port     = "8080"
	Trace    *log.Logger
	Info     *log.Logger
	Warning  *log.Logger
	Error    *log.Logger
)

func GetServiceEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	for _, item := range services {
		if item.ID == params["id"] {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(item)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(404)
}

func GetServicesEndpoint(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

func CreateServiceEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	var service Service
	Info.Println("Creating " + params["id"])
	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		Error.Println(err)
	}
	Trace.Println(string(requestDump))
	_ = json.NewDecoder(req.Body).Decode(&service)
	service.ID = params["id"]
	services = append(services, service)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(service)
}

func DeleteServiceEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	for index, item := range services {
		if item.ID == params["id"] {
			Info.Print("Deleting ", item.ID, "\n")
			Trace.Println(item)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(item)
			services = append(services[:index], services[index+1:]...)
			return
		}
	}
	Error.Println(params["id"], " Not found!")
	w.WriteHeader(404)
	w.Write([]byte(params["id"] + " Not found"))

}

func Init(
	traceHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) {

	Trace = log.New(traceHandle,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {
	// Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
	Init(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	// Trace.Println("I have something standard to say")
	// Info.Println("Special Information")
	// Warning.Println("There is something you need to know about")
	// Error.Println("Something has failed")

	router := mux.NewRouter()
	people = append(people, Person{ID: "1", Firstname: "Nic", Lastname: "Raboy", Address: &Address{City: "Dublin", State: "CA"}})
	people = append(people, Person{ID: "2", Firstname: "Maria", Lastname: "Raboy"})

	services = append(services, Service{Type: "lan2lan", ID: "customer1"})
	services = append(services, Service{Type: "lan2lan", ID: "customer2"})
	services = append(services, Service{Type: "lan2lan", ID: "customer3", ServiceDetails: &ServiceDetails{Vlan: "100", RemoteIP: "1.1.1.1"}})

	router.HandleFunc("/services", GetServicesEndpoint).Methods("GET")
	router.HandleFunc("/services/service/{id}", GetServiceEndpoint).Methods("GET")
	router.HandleFunc("/services/service/{id}", CreateServiceEndpoint).Methods("POST")
	router.HandleFunc("/services/service/{id}", DeleteServiceEndpoint).Methods("DELETE")
	Info.Println("Starting REST Server on port " + port + "...")
	log.Fatal(http.ListenAndServe(":"+port, router))
}
