package util

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	// PrometheusPort is the port on which the prometheus server will run
	PrometheusPort = ":9100"
	// NodeExporterPort is the port on which the node exporter will run
	NodeExporterPort = ":9101"
	// PrometheusEndpoint is the endpoint where the prometheus metrics will be available
	PrometheusEndpoint = "/metrics"
	// HttpPrefixPushMetrics is the endpoint where the metrics will be available to Promtheus server
	HttpPrefixPushMetrics = "/pushmetrics"
)

type ApiHandler struct {
	Handler     http.Handler
	Description string
}

var ApiPrefixes = map[string]ApiHandler{
	PrometheusEndpoint: {promhttp.Handler(), "Prometheus metrics endpoint"},
}

// we create a new custom metric of type counter
var ConnectionStatus = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_request_get_connection_status", // metric name
		Help: "Count of status returned for each connections and iterations.",
	},
	[]string{
		"total_connections",
		"dst_ip_address",
		"dst_port",
		"proto",
		"delay",
		"keep_alive",
		"connections_succeeded",
		"connections_failed",
		"timestamp",
		"iteration",
		"start_timestamp",
		"end_timestamp",
		"total_duration",
		"total_requests",
		"requests_succeeded",
		"requests_failed",
		"requests_per_second",
	},
)

func ConstructDefaultMetric() *Metric {
	return &Metric{
		TotalConnections:     "0",
		ConnectionsSucceeded: "0",
		ConnectionsFailed:    "0",
		TimeStamp:            "0",
		Iteration:            0,
		StartTimestamp:       "0",
		EndTimestamp:         "0",
		TotalDuration:        "0",
		TotalRequests:        ErrorNotImplemented,
		RequestsSucceeded:    ErrorNotImplemented,
		RequestsFailed:       ErrorNotImplemented,
		RequestsPerSecond:    ErrorNotImplemented,
	}
}

func ConstructMetric(args map[string]string, iter int, startTS string) *Metric {
	keepAlive := args[AtribDisableKeepAlive]
	if keepAlive == ConstTrue {
		keepAlive = "KeepAliveDisabled"
	} else {
		keepAliveTime, _ := strconv.Atoi(args[AtribTimeoutKeepAlive])
		keepAliveTime = keepAliveTime / 1000
		keepAlive = fmt.Sprintf("KeepAliveTimeout: %d Seconds", keepAliveTime)
	}
	return &Metric{
		TotalConnections: args[AtribCons],
		DstIpAddress:     args[AtribIpAddr],
		DstPort:          args[AtribPort],
		Proto:            args[AtribProto],
		Delay:            args[AtribDelay],
		KeepAlive:        keepAlive,
		Iteration:        iter,
		StartTimestamp:   startTS,
		TotalRequests:    args[AtribReqs],
	}
}

func UpdateMetricResult(m *Metric, startTS time.Time, connsSuccess, connsFail int) {
	endTS := time.Now()
	totalDur := endTS.Sub(startTS).String()
	m.EndTimestamp = endTS.Format(time.RFC3339)
	m.TotalDuration = totalDur
	m.ConnectionsSucceeded = strconv.Itoa(connsSuccess)
	m.ConnectionsFailed = strconv.Itoa(connsFail)
	m.TimeStamp = strconv.FormatInt(time.Now().Unix(), 10)
}

func PublishConnStatusToPrometheus(m *Metric) {
	ConnectionStatus.WithLabelValues(
		m.TotalConnections,
		m.DstIpAddress,
		m.DstPort,
		m.Proto,
		m.Delay,
		m.KeepAlive,
		m.ConnectionsSucceeded,
		m.ConnectionsFailed,
		m.TimeStamp,
		strconv.Itoa(m.Iteration),
		m.StartTimestamp,
		m.EndTimestamp,
		m.TotalDuration,
		m.TotalRequests,
		m.RequestsSucceeded,
		m.RequestsFailed,
		m.RequestsPerSecond,
	).Inc()
}

func StartPrometheus() {
	prometheus.MustRegister(ConnectionStatus)
	for prefix, handler := range ApiPrefixes {
		log.Printf("%s available at 'http://<pod-ip>%s%s' \n", handler.Description, PrometheusPort, prefix)
		http.Handle(prefix, handler.Handler)
	}
	log.Println("NodeExporter prometheus handler started on port : ", PrometheusPort)
	http.ListenAndServe(PrometheusPort, nil)
}
