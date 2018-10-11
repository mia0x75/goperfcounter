Bench Test
====

Run
----

```bash
cd $GOPATH/src/github.com/mia0x75/gopfc && go test -test.bench=".*"

```

Result
----

```
PASS
BenchmarkMeter   2000000           918 ns/op
BenchmarkMeterMulti  2000000           916 ns/op
BenchmarkGauge   5000000           342 ns/op
BenchmarkGaugeMulti  5000000           333 ns/op
BenchmarkHistogram   2000000           800 ns/op
BenchmarkHistogramMulti  2000000           780 ns/op
ok      github.com/mia0x75/gopfc  14.568s

```