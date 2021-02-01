package metrics

import (
	openmetrics "github.com/prometheus/client_golang/prometheus"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"

	"google.golang.org/grpc"
)

// ServerMetrics represents a collection of metrics to be registered on a
// Prometheus metrics registry for a gRPC server.
type ServerMetrics struct {
	serverRegister openmetrics.Registerer

	serverStartedCounter          *openmetrics.CounterVec
	serverHandledCounter          *openmetrics.CounterVec
	serverStreamMsgReceived       *openmetrics.CounterVec
	serverStreamMsgSent           *openmetrics.CounterVec
	serverHandledHistogramEnabled bool
	serverHandledHistogramOpts    openmetrics.HistogramOpts
	serverHandledHistogram        *openmetrics.HistogramVec
}

// NewServerMetrics returns a ServerMetrics object. Use a new instance of
// ServerMetrics when not using the default Prometheus metrics registry, for
// example when wanting to control which metrics are added to a registry as
// opposed to automatically adding metrics via init functions.
func NewServerMetrics(serverRegistry openmetrics.Registerer, counterOpts ...CounterOption) *ServerMetrics {
	opts := counterOptions(counterOpts)
	return &ServerMetrics{
		serverRegister: serverRegistry,
		serverStartedCounter: openmetrics.NewCounterVec(
			opts.apply(openmetrics.CounterOpts{
				Name: "grpc_server_started_total",
				Help: "Total number of RPCs started on the server.",
			}), []string{"grpc_type", "grpc_service", "grpc_method"}),
		serverHandledCounter: openmetrics.NewCounterVec(
			opts.apply(openmetrics.CounterOpts{
				Name: "grpc_server_handled_total",
				Help: "Total number of RPCs completed on the server, regardless of success or failure.",
			}), []string{"grpc_type", "grpc_service", "grpc_method", "grpc_code"}),
		serverStreamMsgReceived: openmetrics.NewCounterVec(
			opts.apply(openmetrics.CounterOpts{
				Name: "grpc_server_msg_received_total",
				Help: "Total number of RPC stream messages received on the server.",
			}), []string{"grpc_type", "grpc_service", "grpc_method"}),
		serverStreamMsgSent: openmetrics.NewCounterVec(
			opts.apply(openmetrics.CounterOpts{
				Name: "grpc_server_msg_sent_total",
				Help: "Total number of gRPC stream messages sent by the server.",
			}), []string{"grpc_type", "grpc_service", "grpc_method"}),
		serverHandledHistogramEnabled: false,
		serverHandledHistogramOpts: openmetrics.HistogramOpts{
			Name:    "grpc_server_handling_seconds",
			Help:    "Histogram of response latency (seconds) of gRPC that had been application-level handled by the server.",
			Buckets: openmetrics.DefBuckets,
		},
		serverHandledHistogram: nil,
	}
}

// Register registers the provided Collector with the custom register.
// returns error much like DefaultRegisterer of Prometheus.
func (m *ServerMetrics) Register(c openmetrics.Collector) error {
	return m.serverRegister.Register(c)
}

// MustRegister registers the provided Collectors with the custom Registerer
// and panics if any error occurs much like DefaultRegisterer of Prometheus.
func (m *ServerMetrics) MustRegister(c openmetrics.Collector) {
	m.serverRegister.MustRegister(c)
}

// EnableHandlingTimeHistogram turns on recording of handling time
// of RPCs. Histogram metrics can be very expensive for Prometheus
// to retain and query.It takes options to configure histogram
// options such as the defined buckets.
func (m *ServerMetrics) EnableHandlingTimeHistogram(opts ...HistogramOption) error {
	for _, o := range opts {
		o(&m.serverHandledHistogramOpts)
	}
	if !m.serverHandledHistogramEnabled {
		m.serverHandledHistogram = openmetrics.NewHistogramVec(
			m.serverHandledHistogramOpts,
			[]string{"grpc_type", "grpc_service", "grpc_method"},
		)
	}
	m.serverHandledHistogramEnabled = true
	return m.serverRegister.Register(m.serverHandledHistogram)
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector to the provided channel and returns once
// the last descriptor has been sent.
func (m *ServerMetrics) Describe(ch chan<- *openmetrics.Desc) {
	m.serverStartedCounter.Describe(ch)
	m.serverHandledCounter.Describe(ch)
	m.serverStreamMsgReceived.Describe(ch)
	m.serverStreamMsgSent.Describe(ch)
	if m.serverHandledHistogramEnabled {
		m.serverHandledHistogram.Describe(ch)
	}
}

// Collect is called by the Prometheus registry when collecting
// metrics. The implementation sends each collected metric via the
// provided channel and returns once the last metric has been sent.
func (m *ServerMetrics) Collect(ch chan<- openmetrics.Metric) {
	m.serverStartedCounter.Collect(ch)
	m.serverHandledCounter.Collect(ch)
	m.serverStreamMsgReceived.Collect(ch)
	m.serverStreamMsgSent.Collect(ch)
	if m.serverHandledHistogramEnabled {
		m.serverHandledHistogram.Collect(ch)
	}
}

// InitializeMetrics initializes all metrics, with their appropriate null
// value, for all gRPC methods registered on a gRPC server. This is useful, to
// ensure that all metrics exist when collecting and querying.
func (m *ServerMetrics) InitializeMetrics(server *grpc.Server) {
	serviceInfo := server.GetServiceInfo()
	for serviceName, info := range serviceInfo {
		for _, mInfo := range info.Methods {
			m.preRegisterMethod(serviceName, &mInfo)
		}
	}
}

// preRegisterMethod is invoked on Register of a Server, allowing all gRPC services labels to be pre-populated.
func (m *ServerMetrics) preRegisterMethod(serviceName string, mInfo *grpc.MethodInfo) {
	methodName := mInfo.Name
	methodType := string(typeFromMethodInfo(mInfo))
	// These are just references (no increments), as just referencing will create the labels but not set values.
	_, _ = m.serverStartedCounter.GetMetricWithLabelValues(methodType, serviceName, methodName)
	_, _ = m.serverStreamMsgReceived.GetMetricWithLabelValues(methodType, serviceName, methodName)
	_, _ = m.serverStreamMsgSent.GetMetricWithLabelValues(methodType, serviceName, methodName)
	if m.serverHandledHistogramEnabled {
		_, _ = m.serverHandledHistogram.GetMetricWithLabelValues(methodType, serviceName, methodName)
	}
	for _, code := range interceptors.AllCodes {
		_, _ = m.serverHandledCounter.GetMetricWithLabelValues(methodType, serviceName, methodName, code.String())
	}
}
