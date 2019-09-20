package httpinvoke

import (
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/reporter"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"log"
	"net/http"
	"testing"
	"time"
)

func TestHttpTrace(t *testing.T) {

	initZipkinV2()
	span := opentracing.StartSpan("test-case")
	span.Finish()
	ctx := opentracing.ContextWithSpan(context.Background(),span)
	mux := http.NewServeMux()
	wait := make(chan bool,1)
	mux.HandleFunc("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		keys := []string{"x-b3-traceid","x-b3-spanid","x-b3-parentspanid","x-b3-sampled"}
		for _, key := range keys {
			fmt.Println(r.Header.Get(key))
			if r.Header.Get(key) == "" {
				t.Errorf("empty key: %s", key)
			}
		}
		wait <- true
	}))

	server := &http.Server{
		Addr:         ":4400",
		WriteTimeout: 4 * time.Second,
		Handler:      mux,
	}

	defer server.Close()
	//创建http服务监听
	go func(){
		err := server.ListenAndServe()
		if err != nil {
			if err == http.ErrServerClosed {
				log.Print("Server closed under requeset!!")
			} else {
				log.Fatal("Server closed unexpecteed!!")
			}

		}
	}()

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1"+server.Addr, nil)
	if err != nil {
		t.Fatal(err)
	}
	req = req.WithContext(ctx)
	tracertan := NewTraceTracesport(http.DefaultTransport,"","",opentracing.Tag{Key:"ab",Value:"b"})
	client := &http.Client{Transport: tracertan}
	if _, err = client.Do(req); err != nil {
		t.Fatal(err)
	}
	<-wait

}

/**
 *init zipkin tracer
 */
func initZipkinV2() reporter.Reporter {
	msg := "init_zipkin-v2_err"

	// set up a span reporter
	reporter := zipkinhttp.NewReporter("")
	if reporter == nil {

		panic(msg)
	}

	// create our local service endpoint
	endpoint, err := zipkin.NewEndpoint("test", "127.0.0.1:4400")
	if err != nil || endpoint == nil {
		msg := fmt.Sprintf("%v\t%v", msg, err)
		panic(msg)
	}

	// initialize our tracer
	nativeTracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))
	if err != nil || nativeTracer == nil {
		msg := fmt.Sprintf("%v\t%v", msg, err)
		panic(msg)
	}

	// use zipkin-go-opentracing to wrap our tracer
	tracer := zipkinot.Wrap(nativeTracer)
	if tracer == nil {
		panic(msg)
	}

	// optionally set as Global OpenTracing tracer instance
	opentracing.SetGlobalTracer(tracer)

	return reporter
}