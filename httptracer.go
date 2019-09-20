package httpinvoke

import (
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"net/http"
)

const _defaultComponentName = "net/http"


// TraceTransport wraps a RoundTripper. If a request is being traced with
// Tracer, Transport will inject the current span into the headers,
// and set HTTP related tags on the span.
type TraceTransport struct {
	//spankey from parentTrace
	activeSpanKey  string
	//peerService addr
	peerService  string
	//extraTags to be set
	extraTags []opentracing.Tag
	// The actual RoundTripper to use for the request. A nil
	// RoundTripper defaults to http.DefaultTransport.
	http.RoundTripper
}

// NewTraceTracesport NewTraceTracesport
func NewTraceTracesport(rt http.RoundTripper, activeSpanKey  string,peerService string, extraTags ...opentracing.Tag) *TraceTransport {
	return &TraceTransport{RoundTripper: rt,activeSpanKey:activeSpanKey, peerService: peerService, extraTags: extraTags}
}

// RoundTrip implements the RoundTripper interface
func (t *TraceTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rt := t.RoundTripper
	if rt == nil {
		rt = http.DefaultTransport
	}
	var tr opentracing.Span
	if t.activeSpanKey != "" {
		acvspans := req.Context().Value(t.activeSpanKey)
		if acvspans == nil{
			return rt.RoundTrip(req)
		}
		tr = opentracing.SpanFromContext(opentracing.ContextWithSpan(req.Context(), acvspans.(opentracing.Span)))
	} else {
		tr = opentracing.SpanFromContext(req.Context())
	}

	if tr == nil {
		return rt.RoundTrip(req)
	}
	operationName := "HTTP:" + req.Method
	//get start new tracer
	tracer := tr.Tracer()
	span := tracer.StartSpan(operationName, opentracing.ChildOf(tr.Context()))

	ext.DBType.Set(span, "http")
	//ext.PeerAddress.Set(span, req.URL.String())
	ext.HTTPMethod.Set(span,req.Method)
	ext.HTTPUrl.Set(span, req.URL.String())
	/*ext.SpanKind.Set(span, ext.SpanKindRPCClientEnum)*/
	ext.Component.Set(span,_defaultComponentName)
	//end
	if t.peerService != "" {
		ext.PeerService.Set(span,t.peerService)
	}
	for _, v := range t.extraTags {
		span.SetTag(v.Key, v.Value)
	}
	// inject trace to http header
	tracer.Inject(span.Context(),opentracing.HTTPHeaders,req.Header)

	// ct := clientTracer{tr: tr}
	// req = req.WithContext(httptrace.WithClientTrace(req.Context(), ct.clientTrace()))
	resp, err := rt.RoundTrip(req)

	if err != nil {
		ext.Error.Set(span, true)
		span.LogFields(log.String("event", "error"), log.String("message", err.Error()))
		span.Finish()
		return resp, err
	}

	ext.HTTPStatusCode.Set(span,uint16(resp.StatusCode))
	if resp.StatusCode >= 400 {
		ext.Error.Set(span, true)
	}
	span.Finish()
	return resp, err
}

