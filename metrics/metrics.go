package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/qietv/kits/utils"
	"net/http"
	"sync"
)

var (
	metricSgt sync.Once
)
var DefaultMetrics = &Metric{
	Path:    "/metrics",
	Address: utils.GetIP(),
	Port:    9909,
}

type Metric struct {
	NodeName   string                 `yaml:"node"`
	Path       string                 `yaml:"path"`
	Address    string                 `yaml:"address"`
	Port       int                    `yaml:"port"`
	Listen     string                 `yaml:"listen"`
	Handler    http.Handler           `yaml:"-"`
	Collectors []prometheus.Collector `yaml:"-"`
	server     *http.Server           `yaml:"-"`
}

func InitMetrics(opts *Metric) {

	if opts.Handler != nil {
		h := opts.Handler
		metricSgt.Do(func() {
			switch h.(type) {
			case *gin.Engine:
				h.(*gin.Engine).GET(opts.Path, func() gin.HandlerFunc {
					h := promhttp.Handler()
					return func(c *gin.Context) {
						h.ServeHTTP(c.Writer, c.Request)
					}
				}())
			default:
				http.Handle(opts.Path, promhttp.Handler())
			}
		})
	} else {
		http.Handle(opts.Path, promhttp.Handler())
	}
	if opts.Listen != "" {
		opts.server = &http.Server{
			Addr: opts.Listen,
		}
		go func() {
			if err := opts.server.ListenAndServe(); err != nil {
				utils.GetLogger().Info("metrics listen fail, %s", err.Error())
			}
		}()
	}

}

func Register(collector ...prometheus.Collector) {
	prometheus.MustRegister(collector...)
}
