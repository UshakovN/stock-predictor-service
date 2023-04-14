package fetcher

type Fetcher interface {
  ContinuouslyFetch()
  SaveFetcherState()
  SetTickerId(tickerId string)
}
