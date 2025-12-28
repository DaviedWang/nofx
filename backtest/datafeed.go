package backtest

import (
	"fmt"
	"sort"
	"time"

	"nofx/market"
)

type timeframeSeries struct {
	klines     []market.Kline
	closeTimes []int64
}

type symbolSeries struct {
	byTF map[string]*timeframeSeries
}

// DataFeed manages historical kline data and provides time-progressive snapshots for backtesting.
type DataFeed struct {
	cfg           BacktestConfig
	symbols       []string
	timeframes    []string
	symbolSeries  map[string]*symbolSeries
	decisionTimes []int64
	primaryTF     string
	longerTF      string
}

func NewDataFeed(cfg BacktestConfig) (*DataFeed, error) {
	// Ensure DecisionTimeframe is included in the timeframes list
	timeframes := make([]string, 0, len(cfg.Timeframes)+1)
	decisionTFIncluded := false
	for _, tf := range cfg.Timeframes {
		timeframes = append(timeframes, tf)
		if tf == cfg.DecisionTimeframe {
			decisionTFIncluded = true
		}
	}
	// Add DecisionTimeframe if not already in the list
	if !decisionTFIncluded && cfg.DecisionTimeframe != "" {
		timeframes = append(timeframes, cfg.DecisionTimeframe)
	}

	df := &DataFeed{
		cfg:          cfg,
		symbols:      make([]string, len(cfg.Symbols)),
		timeframes:   timeframes,
		symbolSeries: make(map[string]*symbolSeries),
		primaryTF:    cfg.DecisionTimeframe,
	}
	copy(df.symbols, cfg.Symbols)

	if err := df.loadAll(); err != nil {
		return nil, err
	}

	return df, nil
}

func (df *DataFeed) loadAll() error {
	start := time.Unix(df.cfg.StartTS, 0)
	end := time.Unix(df.cfg.EndTS, 0)

	// longest timeframe used for auxiliary indicators
	var longestDur time.Duration
	for _, tf := range df.timeframes {
		dur, err := market.TFDuration(tf)
		if err != nil {
			return err
		}
		if dur > longestDur {
			longestDur = dur
			df.longerTF = tf
		}
	}

	for _, symbol := range df.symbols {
		ss := &symbolSeries{byTF: make(map[string]*timeframeSeries)}
		for _, tf := range df.timeframes {
			dur, _ := market.TFDuration(tf)
			buffer := dur * 200
			fetchStart := start.Add(-buffer)
			if fetchStart.Before(time.Unix(0, 0)) {
				fetchStart = time.Unix(0, 0)
			}
			fetchEnd := end.Add(dur)

			klines, err := market.GetKlinesRange(symbol, tf, fetchStart, fetchEnd)
			if err != nil {
				return fmt.Errorf("fetch klines for %s %s: %w", symbol, tf, err)
			}
			if len(klines) == 0 {
				return fmt.Errorf("no klines for %s %s", symbol, tf)
			}

			series := &timeframeSeries{
				klines:     klines,
				closeTimes: make([]int64, len(klines)),
			}
			for i, k := range klines {
				series.closeTimes[i] = k.CloseTime
			}
			ss.byTF[tf] = series
		}
		df.symbolSeries[symbol] = ss
	}

	// Generate backtest progress timeline using the primary timeframe of the first symbol
	firstSymbol := df.symbols[0]
	primarySeries := df.symbolSeries[firstSymbol].byTF[df.primaryTF]
	if primarySeries == nil {
		return fmt.Errorf("primary timeframe %s data not found for symbol %s", df.primaryTF, firstSymbol)
	}
	startMs := start.UnixMilli()
	endMs := end.UnixMilli()
	for _, ts := range primarySeries.closeTimes {
		if ts < startMs {
			continue
		}
		if ts > endMs {
			break
		}
		df.decisionTimes = append(df.decisionTimes, ts)
		// Align other symbols; report error early if data is missing
		for _, symbol := range df.symbols[1:] {
			if _, ok := df.symbolSeries[symbol].byTF[df.primaryTF]; !ok {
				return fmt.Errorf("symbol %s missing timeframe %s", symbol, df.primaryTF)
			}
		}
	}
	if len(df.decisionTimes) == 0 {
		return fmt.Errorf("no decision bars in range")
	}
	return nil
}

func (df *DataFeed) DecisionBarCount() int {
	return len(df.decisionTimes)
}

func (df *DataFeed) DecisionTimestamp(index int) int64 {
	return df.decisionTimes[index]
}

func (df *DataFeed) sliceUpTo(symbol, tf string, ts int64) []market.Kline {
	series := df.symbolSeries[symbol].byTF[tf]
	idx := sort.Search(len(series.closeTimes), func(i int) bool {
		return series.closeTimes[i] > ts
	})
	if idx <= 0 {
		return nil
	}
	return series.klines[:idx]
}

func (df *DataFeed) BuildMarketData(ts int64) (map[string]*market.Data, map[string]map[string]*market.Data, error) {
	result := make(map[string]*market.Data, len(df.symbols))
	multi := make(map[string]map[string]*market.Data, len(df.symbols))

	for _, symbol := range df.symbols {
		perTF := make(map[string]*market.Data, len(df.timeframes))
		for _, tf := range df.timeframes {
			series := df.sliceUpTo(symbol, tf, ts)
			if len(series) == 0 {
				continue
			}
			var longer []market.Kline
			if df.longerTF != "" && df.longerTF != tf {
				longer = df.sliceUpTo(symbol, df.longerTF, ts)
			}
			data, err := market.BuildDataFromKlines(symbol, series, longer)
			if err != nil {
				return nil, nil, err
			}
			perTF[tf] = data
			if tf == df.primaryTF {
				result[symbol] = data
			}
		}
		if _, ok := perTF[df.primaryTF]; !ok {
			return nil, nil, fmt.Errorf("no primary data for %s at %d", symbol, ts)
		}

		// Build TimeframeData for primary timeframe data to support strategy engine
		// This is needed for formatMarketData to work correctly
		if primaryData, hasPrimary := perTF[df.primaryTF]; hasPrimary {
			primaryData.TimeframeData = make(map[string]*market.TimeframeSeriesData, len(df.timeframes))
			for _, tf := range df.timeframes {
				// Re-fetch the raw Kline data for this timeframe
				series := df.sliceUpTo(symbol, tf, ts)
				if len(series) == 0 {
					continue
				}

				// Convert Kline to KlineBar for TimeframeSeriesData
				klineBars := make([]market.KlineBar, len(series))
				for i, k := range series {
					klineBars[i] = market.KlineBar{
						Time:   k.OpenTime,
						Open:   k.Open,
						High:   k.High,
						Low:    k.Low,
						Close:  k.Close,
						Volume: k.Volume,
					}
				}

				// Get the calculated indicator data from the built market data
				tfData, hasData := perTF[tf]
				if !hasData || tfData.IntradaySeries == nil {
					// Skip if no data for this timeframe
					continue
				}

				primaryData.TimeframeData[tf] = &market.TimeframeSeriesData{
					Timeframe:   tf,
					Klines:      klineBars,
					MidPrices:   tfData.IntradaySeries.MidPrices,
					EMA20Values: tfData.IntradaySeries.EMA20Values,
					MACDValues:  tfData.IntradaySeries.MACDValues,
					RSI7Values:  tfData.IntradaySeries.RSI7Values,
					RSI14Values: tfData.IntradaySeries.RSI14Values,
					Volume:      tfData.IntradaySeries.Volume,
					ATR14:       tfData.IntradaySeries.ATR14,
				}
			}
		}

		multi[symbol] = perTF
	}
	return result, multi, nil
}

func (df *DataFeed) decisionBarSnapshot(symbol string, ts int64) (*market.Kline, *market.Kline) {
	ss, ok := df.symbolSeries[symbol]
	if !ok {
		return nil, nil
	}
	series, ok := ss.byTF[df.primaryTF]
	if !ok {
		return nil, nil
	}
	idx := sort.Search(len(series.closeTimes), func(i int) bool {
		return series.closeTimes[i] >= ts
	})
	if idx >= len(series.closeTimes) || series.closeTimes[idx] != ts {
		return nil, nil
	}
	curr := &series.klines[idx]
	var next *market.Kline
	if idx+1 < len(series.klines) {
		next = &series.klines[idx+1]
	}
	return curr, next
}
