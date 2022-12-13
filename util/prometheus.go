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
	[]string{"Connections", "ConnectionsSucceded", "ConnectionsFailed", "Iteration"}, // labels
)

func PublishConnStatusToPrometheus(totalConnCount, passedConnCount, failedConnCount, iterationNum int) {
	connectionStatus.WithLabelValues(strconv.Itoa(totalConnCount), strconv.Itoa(passedConnCount), strconv.Itoa(failedConnCount), strconv.Itoa(iterationNum)).Inc()
}

func StartPrometheus() {
	prometheus.MustRegister(connectionStatus)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
