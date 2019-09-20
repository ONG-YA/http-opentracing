# opentracing net/http



[OpenTracing](http://opentracing.io/) instrumentation for [net/http]

## Usage 

implements http RoundTrip by `NewTraceTracesport(rt http.RoundTripper, activeSpanKey  string,peerService string, extraTags ...opentracing.Tag)` .

Example :

```go
tracertan := httpinvoke.NewTraceTracesport(http.DefaultTransport,"","",opentracing.Tag{Key:"ab",Value:"b"})

	client := &http.Client{
		Transport:tracertan,
	}
```
Example for rpcx :

```go
	tracertan := httpinvoke.NewTraceTracesport(http.DefaultTransport,share.OpentracingSpanServerKey,"",opentracing.Tag{Key:"ab",Value:"b"})

	client := &http.Client{
		Transport:tracertan,
	}
```