# SFX Datapoints Generator

This utility generates datapoints using the signalfx format and concurrently sends them to the configured ingest. 

## Options

````
./sfxdatagen metrics --help                                                                                                                                    ✔    18:59:03 
Simulates a client generating metrics

Usage:
  sfxdatagen metrics [flags]

Examples:
sfxdatagen metrics

Flags:
      --access-token string     SignalFX Ingest Access Token
      --batch-writer-size int   How many datapoints batch to submit to the writer (default 100)
      --duration duration       For how long to run the test
  -h, --help                    help for metrics
      --http-timeout duration   Writer client http timeout (default 5s)
      --metrics int             Number of metrics to generate in each worker (ignored if duration is provided) (default 1)
      --rate int                Approximately how many metrics per second each worker should generate. Zero means no throttling.
      --realm string            Realm your Org belongs to (default "us0")
      --workers int             Number of workers (goroutines) to run (default 1)
````
## Additional Info:
The Generated Metric name is `gensfxdp<worker_index>` example: `gensfxdp1`

## Running:

example: 
````
sfxdatagen-linux metrics --rate 900 --realm eu0 --metrics 25000 --batch-writer-size 900 --workers 15 --access-token <ingest token value>
````