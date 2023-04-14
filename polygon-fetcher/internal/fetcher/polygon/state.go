package polygon

import (
	"fmt"
	"main/internal/domain"
	"time"

	"github.com/UshakovN/stock-predictor-service/utils"
	log "github.com/sirupsen/logrus"
)

type state struct {
	finished         bool
	updatedAt        *time.Time
	modeCode         int
	modeTotalHours   int
	modeCurrentHours int
	ticker           *stateReq
	tickerDetails    *stateReq
	stocks           *stateReq
}

type stateReq struct {
	reqURL string
	used   bool
}

func newFetcherState(modeTotalHours, modeCurrentHours int) *state {
	return &state{
		modeCode:         fetcherModeTotal,
		modeTotalHours:   modeTotalHours,
		modeCurrentHours: modeCurrentHours,
		ticker:           &stateReq{},
		tickerDetails:    &stateReq{},
		stocks:           &stateReq{},
	}
}

func (s *state) SetFinished() {
	s.finished = true
}

func (s *state) ResetFinished() {
	s.finished = false
}

func (s *state) SetUpdatedTime(t time.Time) {
	s.updatedAt = &t
}

func (s *state) SetModeCode(mode int) {
	if mode != fetcherModeTotal && mode != fetcherModeCurrent {
		log.Warnf("invalid fetcher mode code: %d. mode code do not set. possible: %d - total, %d - current",
			mode, fetcherModeTotal, fetcherModeCurrent)
		return
	}
	s.modeCode = mode
	log.Infof("current fetcher mode: %d. possible: %d - total, %d - current",
		mode, fetcherModeTotal, fetcherModeCurrent)
}

func (f *Fetcher) SetTickerId(tickerId string) {
	if tickerId == "" {
		log.Warnf("ticker id is empty. ticker id do not set")
		return
	}
	f.specTickerId = tickerId
}

func (f *Fetcher) hasRecentlyFetched() bool {
	if !f.state.finished || f.state.updatedAt == nil {
		return false
	}
	updatedAt := *f.state.updatedAt
	thresholdTime := updatedAt.Add(recentlyThresholdInterval)
	return thresholdTime.After(time.Now())
}

func (f *Fetcher) hasSpecTickerId() bool {
	return f.specTickerId != ""
}

func (f *Fetcher) ContinuouslyFetch() {
	if !f.hasSpecTickerId() {
		// if spec ticker id not set
		if err := f.loadFetcherState(); err != nil {
			log.Errorf("state loading from storage failed. : %v", err)
		}
	}
	f.state.SetModeCode(fetcherModeTotal)

	tryLeft := fetcherRetryCount
	// fetch with retries
	for tryLeft >= 0 {
		if f.hasRecentlyFetched() {
			log.Printf("recently fetched. wait %v before the next fetch",
				recentlyFetchedSleepInterval)

			time.Sleep(recentlyFetchedSleepInterval)
			f.state.ResetFinished() // reset finished field
			continue
		}
		var err error

		if f.hasSpecTickerId() {
			err = f.fetchSpecTicker(f.specTickerId)
		} else {
			err = f.fetchTickers()
		}
		if err != nil {
			log.Errorf("fetching error: %v. wait %v before the next fetch",
				err, encounteredErrorSleepInterval)

			time.Sleep(encounteredErrorSleepInterval)
			tryLeft--
			continue
		}
		f.state.SetUpdatedTime(utils.NotTimeUTC())
		f.state.SetFinished() // set finished field

		// set retry count again
		tryLeft = fetcherRetryCount
		log.Println("successfully fetching finished")

		if f.hasSpecTickerId() {
			return
		}
		f.once.Do(func() {
			f.state.SetModeCode(fetcherModeCurrent)
		})
	}
	log.Fatalf("fetching failed and stopped")
}

func (f *Fetcher) SaveFetcherState() {
	if err := f.storage.PutFetcherState(createFetcherState(f.state)); err != nil {
		log.Fatalf("cannot put fetcher state to storage: %v", err)
	}
}

func (f *Fetcher) loadFetcherState() error {
	state, found, err := f.storage.GetFetcherState()
	if err != nil {
		return fmt.Errorf("cannot get fetcher state from storage: %v", err)
	}
	if !found {
		return nil
	}
	// set fields from storage state
	f.state.ticker.reqURL = utils.StripString(state.TickerReqUrl)
	f.state.tickerDetails.reqURL = utils.StripString(state.TickerDetailsReqUrl)
	f.state.stocks.reqURL = utils.StripString(state.StockReqUrl)
	f.state.updatedAt = &state.CreatedAt
	f.state.finished = state.Finished

	return nil
}

func createFetcherState(state *state) *domain.FetcherState {
	if state == nil {
		return nil
	}
	return &domain.FetcherState{
		TickerReqUrl:        state.ticker.reqURL,
		TickerDetailsReqUrl: state.tickerDetails.reqURL,
		StockReqUrl:         state.stocks.reqURL,
		CreatedAt:           utils.NotTimeUTC(),
	}
}
