package util

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// we create a new custom metric of type counter
var connectionStatus = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_request_get_connection_status", // metric name
		Help: "Count of status returned for each connections and iterations.",
	},
	[]string{"ConnectionsSucceded", "ConnectionsFailed", "Iteration", "FaliledConnections"}, // labels
)

func PublishConnStatusToPrometheus(passedConnCount, failedConnCount, iterationNum int, failedConns string) {
	connectionStatus.WithLabelValues(strconv.Itoa(passedConnCount), strconv.Itoa(failedConnCount), strconv.Itoa(iterationNum), failedConns).Inc()
}

func StartPrometheus() {
	prometheus.MustRegister(connectionStatus)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
