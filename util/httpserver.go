package util

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type HttpHandler struct {
	Handler     http.HandlerFunc
	Description string
}

var HttpPrefixes = map[string]HttpHandler{
	HttpPrefixPushMetrics: {PushMetrics, "Prometheus pushmetrics endpoint"},
}

type Metric struct {
	TotalConnections     string `json:"total_connections,omitempty"`
	DstIpAddress         string `json:"dst_ip_address,omitempty"`
	DstPort              string `json:"dst_port,omitempty"`
	Proto                string `json:"proto,omitempty"`
	Delay                string `json:"delay,omitempty"`
	KeepAlive            string `json:"keep_alive,omitempty"`
	ConnectionsSucceeded string `json:"connections_succeeded,omitempty"`
	ConnectionsFailed    string `json:"connections_failed,omitempty"`
	TimeStamp            string `json:"timestamp,omitempty"`
	Iteration            int    `json:"iteration,omitempty"`
	StartTimestamp       string `json:"start_timestamp,omitempty"`
	EndTimestamp         string `json:"end_timestamp,omitempty"`
	TotalDuration        string `json:"total_duration,omitempty"`
	TotalRequests        string `json:"total_requests,omitempty"`
	RequestsSucceeded    string `json:"requests_succeeded,omitempty"`
	RequestsFailed       string `json:"requests_failed,omitempty"`
	RequestsPerSecond    string `json:"requests_per_second,omitempty"`
}

func PushMetrics(w http.ResponseWriter, req *http.Request) {
	log.Printf("Push Metrics Handler called ...")
	var m Metric

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(req.Body).Decode(&m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("Metrics received : %+v \n", m)
	// Print the metric struct...
	fmt.Fprintf(w, "Metric: %+v", m)

	PublishConnStatusToPrometheus(&m)
}

func StartHttpServer() {
	for prefix, handler := range HttpPrefixes {
		log.Printf("%s available at 'http://<pod-ip>%s%s' \n", handler.Description, NodeExporterPort, prefix)
		http.Handle(prefix, handler.Handler)
	}
	log.Println("NodeExporter http request handler started on port : ", NodeExporterPort)
	http.ListenAndServe(NodeExporterPort, nil)
}
